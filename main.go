package main

import (
	"net"

	"github.com/redsuperbat/nano-flow/data"
	"github.com/redsuperbat/nano-flow/logging"
	"github.com/redsuperbat/nano-flow/server"
)

const filepath = "messages.db"

func main() {
	logger := logging.New()
	file, err := data.InitDatabase(filepath)
	if err != nil {
		logger.Fatalln(err)
	}
	_, err = data.InitWatcher(filepath)
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
	if err = server.Start(lis); err != nil {
		logger.Fatalf("failed to start server %s", err)
	}
}
