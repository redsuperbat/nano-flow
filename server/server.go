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
}

func (s *Server) AppendMessage(ctx context.Context, req *pb.AppendRequest) (*pb.Empty, error) {
	message := data.NewMessage(req.Data)
	s.messageService.AppendMessage(&message)
	return &pb.Empty{}, nil
}

func (s *Server) SubscribeToMessages(req *pb.SubscriptionRequest, cb pb.MessageService_SubscribeToMessagesServer) error {
	messages, _ := s.messageService.GetAllMessages()
	for _, msg := range messages {
		cb.Send(&pb.SubscriptionStream{
			Crc:       int32(msg.Crc),
			Version:   int32(msg.Version),
			Timestamp: msg.Timestamp,
			Data:      msg.Data,
		})
	}
	return nil
}

func (s *Server) Start(lis net.Listener) error {
	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func New(ms *data.MessageService, logger *zap.SugaredLogger) *Server {
	return &Server{
		messageService: ms,
		logger:         logger,
	}
}
