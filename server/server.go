package server

import (
	"context"
	"net"

	"github.com/redsuperbat/nano-flow/data"
	pb "github.com/redsuperbat/nano-flow/rpc/messages"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedMessageServiceServer
	messageService *data.MessageService
	logger         *zap.SugaredLogger
	listeners      map[string]data.MessageChannel
}

func (s *Server) AppendMessage(ctx context.Context, req *pb.AppendRequest) (*pb.Empty, error) {
	message := data.NewMessage(req.Data)
	s.messageService.AppendMessage(&message)
	for _, listener := range s.listeners {
		listener <- &message
	}
	return &pb.Empty{}, nil
}

func (s *Server) SubscribeToMessages(req *pb.SubscriptionRequest, cb pb.MessageService_SubscribeToMessagesServer) error {
	messages, _ := s.messageService.GetAllMessages()
	for _, msg := range messages {
		stream := pb.SubscriptionStream{
			Crc:       int32(msg.Crc),
			Version:   int32(msg.Version),
			Data:      msg.Data,
			Timestamp: msg.Timestamp,
		}
		err := cb.Send(&stream)
		if err != nil {
			return err
		}
	}

	// Add listener to the server
	msgChan := make(data.MessageChannel)
	s.listeners[req.GroupId] = msgChan

	// Subscribe to messages
	for msg := range msgChan {
		cb.Send(&pb.SubscriptionStream{
			Crc:       int32(msg.Crc),
			Timestamp: msg.Timestamp,
			Version:   int32(msg.Version),
			Data:      msg.Data,
		})
	}

	return nil
}

func (s *Server) Start(lis net.Listener, grpcServer *grpc.Server) error {
	pb.RegisterMessageServiceServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func New(ms *data.MessageService, logger *zap.SugaredLogger) *Server {
	return &Server{
		messageService: ms,
		logger:         logger,
		listeners:      map[string]data.MessageChannel{},
	}
}
