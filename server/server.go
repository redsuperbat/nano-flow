package server

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/redsuperbat/nano-flow/data"
	pb "github.com/redsuperbat/nano-flow/rpc/messages"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Listeners = map[string]map[string]data.MessageChannel

type Server struct {
	pb.UnimplementedMessageServiceServer
	messageService *data.MessageService
	logger         *zap.SugaredLogger
	listeners      Listeners
	clients        int64
}

func (s *Server) AppendMessage(ctx context.Context, req *pb.AppendRequest) (*pb.Empty, error) {
	message := data.NewMessage(req.Data)
	s.messageService.AppendMessage(&message)
	for _, group := range s.listeners {
		for _, channel := range group {
			channel <- &message
		}
	}
	return &pb.Empty{}, nil
}

func getMessageChannelFromGroupId(groupId string, listeners Listeners, requestId string) map[string]data.MessageChannel {
	if value, found := listeners[groupId]; found {
		value[requestId] = make(data.MessageChannel)
		return value
	} else {
		listeners[groupId] = make(map[string]data.MessageChannel)
		listeners[groupId][requestId] = make(data.MessageChannel)
		return listeners[groupId]
	}
}

func (s *Server) SubscribeToMessages(req *pb.SubscriptionRequest, cb pb.MessageService_SubscribeToMessagesServer) error {
	_, found := s.listeners[req.GroupId]
	s.clients += 1
	s.logger.Infof("client connected \t#clients %d", s.clients)
	if !found {
		messages, _ := s.messageService.GetAllMessages()
		for _, msg := range messages {
			nMsg := pb.NanoMessage{
				Crc:       msg.Crc,
				Version:   uint32(msg.Version),
				Data:      msg.Data,
				Timestamp: msg.Timestamp,
			}
			err := cb.Send(&nMsg)
			if err != nil {
				return err
			}
		}
	}
	requestId := uuid.New().String()
	// Add listener to the server
	channelMap := getMessageChannelFromGroupId(req.GroupId, s.listeners, requestId)
	// Subscribe to messages
	for {
		select {
		case msg := <-channelMap[requestId]:
			cb.Send(&pb.NanoMessage{
				Crc:       msg.Crc,
				Timestamp: msg.Timestamp,
				Version:   uint32(msg.Version),
				Data:      msg.Data,
			})
		case <-cb.Context().Done():
			s.clients -= 1
			s.logger.Infof("client disconnected \t#clients %d", s.clients)
			delete(channelMap, requestId)
			return nil
		}
	}

}

func (s *Server) Start(lis net.Listener, grpcServer *grpc.Server) error {
	pb.RegisterMessageServiceServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func New(ms *data.MessageService, logger *zap.SugaredLogger) *Server {
	return &Server{
		messageService: ms,
		logger:         logger,
		listeners:      Listeners{},
		clients:        0,
	}
}
