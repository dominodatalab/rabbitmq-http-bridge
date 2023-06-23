package main

import (
	"bridge/bridge"
	"fmt"
	"github.com/go-logr/stdr"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	logger := stdr.New(log.New(os.Stdout, "", log.Lshortfile))
	if len(args) != 4 {
		logger.Error(fmt.Errorf("exactly 4 arguments required, found %q", args), "")
		os.Exit(1)
	}
	serverOptions := bridge.RabbitMQServerOptions{
		ServerUrl:       args[0],
		BrokerUrl:       args[1],
		InputQueueName:  args[2],
		OutputQueueName: args[3],
		Log:             logger,
	}
	server, err := bridge.CreateRabbitMQServer(serverOptions)
	if err != nil {
		logger.Error(err, "error creating rabbitmq server")
		os.Exit(1)
	}
	err = server.Serve()
	if err != nil {
		logger.Error(err, "error while running rabbitmq server")
		os.Exit(1)
	}
	logger.Info("rabbitmq server terminated normally")
}
