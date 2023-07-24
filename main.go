package main

import (
	"context"
	"net"

	"github.com/redsuperbat/nano-flow/data"
	"github.com/redsuperbat/nano-flow/logging"
	pb "github.com/redsuperbat/nano-flow/rpc/messages"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMessageServiceServer
}

func (s *server) AppendMessage(ctx context.Context, req *pb.AppendRequest) (*pb.AppendResponse, error) {
	return &pb.AppendResponse{
		Id: "",
	}, nil
}

func (s *server) SubscribeToMessages(*pb.SubscriptionRequest, pb.MessageService_SubscribeToMessagesServer) error {
	return nil
}

const filepath = "messages.db"

func main() {
	logger := logging.New()
	file, err := data.Init(filepath)
	if err != nil {
		logger.Fatalln(err)
	}
	recordService := data.NewMessageService(file)
	_, err = recordService.GetAllMessages()
	if err != nil {
		logger.Fatalln(err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, &server{})
	logger.Infoln("gRPC server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}
