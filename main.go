package main

import (
	"context"
	"log"
	"net"

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

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, &server{})
	log.Println("gRPC server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
