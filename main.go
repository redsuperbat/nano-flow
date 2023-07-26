package main

import (
	"net"
	"os"

	"github.com/redsuperbat/nano-flow/data"
	"github.com/redsuperbat/nano-flow/logging"
	"github.com/redsuperbat/nano-flow/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func getDbPath() string {
	const filepath = "nano-flow.db"
	path, found := os.LookupEnv("NANO_DATABASE_PATH")
	if !found {
		return filepath
	}
	return path
}

func main() {
	filepath := getDbPath()
	logger := logging.New()
	file, err := data.InitDatabase(filepath)
	if err != nil {
		logger.Fatalln(err)
	}
	if err != nil {
		logger.Fatalln(err)
	}
	messageService := data.NewMessageService(file)
	server := server.New(messageService, logger)
	if err != nil {
		logger.Fatalln(err)
	}
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}
	logger.Infoln("server is running on port 50051 ready to accept connections")
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	if err = server.Start(lis, grpcServer); err != nil {
		logger.Fatalf("failed to start server %s", err)
	}
}
