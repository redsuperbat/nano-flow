package main

import (
	"context"
	"net"

	"github.com/redsuperbat/nano-flow/data"
	"github.com/redsuperbat/nano-flow/logging"
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
	s.logger.Infoln("appending a message")
	message := data.NewMessage(req.Body)
	s.messageService.AppendMessage(&message)

	return &pb.Empty{}, nil
}

func (s *Server) SubscribeToMessages(*pb.SubscriptionRequest, pb.MessageService_SubscribeToMessagesServer) error {
	return nil
}

const filepath = "messages.db"

func main() {
	logger := logging.New()
	file, err := data.Init(filepath)
	if err != nil {
		logger.Fatalln(err)
	}
	messageService := data.NewMessageService(file)
	msgs, err := messageService.GetAllMessages()
	server := Server{messageService: &messageService, logger: logger}
	if err != nil {
		logger.Fatalln(err)
	}

	for i, msg := range msgs {
		logger.Infof("%d. len: %d, content: %s", i, msg.ContentLength, msg.Data)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, &server)
	logger.Infoln("gRPC server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}
