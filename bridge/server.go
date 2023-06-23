package bridge

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/imroc/req/v3"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	ENV_RABBITMQ_BROKER_URL    = "RABBITMQ_BROKER_URL"
	ENV_RABBITMQ_INPUT_QUEUE   = "RABBITMQ_INPUT_QUEUE"
	ENV_RABBITMQ_OUTPUT_QUEUE  = "RABBITMQ_OUTPUT_QUEUE"
	ENV_RABBITMQ_FULL_GRAPH    = "RABBITMQ_FULL_GRAPH"
	UNHANDLED_ERROR            = "Unhandled error from predictor process"
	DEFAULT_MAX_MSG_SIZE_BYTES = 10240
)

type RabbitMQServer struct {
	ServerUrl       string
	BrokerUrl       string
	InputQueueName  string
	OutputQueueName string
	Log             logr.Logger
	HttpClient      *req.Client
}

type RabbitMQServerOptions struct {
	ServerUrl       string
	BrokerUrl       string
	InputQueueName  string
	OutputQueueName string
	Log             logr.Logger
}

func CreateRabbitMQServer(args RabbitMQServerOptions) (*RabbitMQServer, error) {

	return &RabbitMQServer{
		ServerUrl:       args.ServerUrl,
		BrokerUrl:       args.BrokerUrl,
		InputQueueName:  args.InputQueueName,
		OutputQueueName: args.OutputQueueName,
		Log:             args.Log.WithName("RabbitMqServer"),
		HttpClient:      req.C(),
	}, nil
}

func (rs *RabbitMQServer) Serve() error {
	conn, err := createRabbitMQConnection(rs.BrokerUrl, rs.Log)
	if err != nil {
		rs.Log.Error(err, "error connecting to rabbitmq")
		return fmt.Errorf("error '%w' connecting to rabbitmq", err)
	}
	defer func(conn ConnectionWrapper) {
		err := conn.Close()
		if err != nil {
			rs.Log.Error(err, "error closing rabbitMQ connection")
		}
	}(conn)

	wg := new(sync.WaitGroup)
	terminateChan, err := rs.serve(conn, wg)
	if err != nil {
		rs.Log.Error(err, "error starting rabbitmq server")
		return fmt.Errorf("error '%w' starting rabbitmq server", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	// wait for shutdown signal and terminate if received
	go func() {
		rs.Log.Info("awaiting OS shutdown signals")
		sig := <-sigs
		rs.Log.Info("sending termination message due to signal", "signal", sig)
		terminateChan <- true
		wg.Done()
	}()

	wg.Wait()

	rs.Log.Info("RabbitMQ server terminated normally")
	return nil
}

func (rs *RabbitMQServer) serve(conn ConnectionWrapper, wg *sync.WaitGroup) (chan<- bool, error) {
	// not sure if this is the best pattern or better to pass in pod name explicitly somehow
	consumerTag, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error '%w' retrieving hostname", err)
	}

	publisher, err := conn.NewPublisher()
	if err != nil {
		return nil, fmt.Errorf("error '%w' creating RMQ publisher", err)
	}
	rs.Log.Info("Created", "publisher", publisher)

	consumerHandler := CreateConsumerHandler(
		func(reqPl func(*req.Request)) error { return rs.predictAndPublishResponse(reqPl, publisher) },
		func(args ConsumerError) error { return rs.createAndPublishErrorResponse(args, publisher) },
		rs.Log,
	)

	consumer, err := conn.NewConsumer(consumerHandler, rs.InputQueueName, consumerTag)
	if err != nil {
		return nil, fmt.Errorf("error '%w' creating RMQ consumer", err)
	}
	rs.Log.Info("Created", "consumer", consumer, "input queue", rs.InputQueueName)

	// provide a channel to terminate the server
	terminate := make(chan bool, 1)

	wg.Add(1)
	go func() {
		rs.Log.Info("awaiting group termination")
		<-terminate
		rs.Log.Info("termination initiated, shutting down")
		consumer.Close()
		publisher.Close()
		wg.Done()
	}()

	return terminate, nil
}

func (rs *RabbitMQServer) predictAndPublishResponse(
	reqLoader func(*req.Request),
	publisher PublisherWrapper,
) error {
	request := rs.HttpClient.R()
	reqLoader(request)

	res, err := request.Post(rs.ServerUrl)
	if err != nil {
		rs.Log.Error(err, UNHANDLED_ERROR)
		return fmt.Errorf("unhandled error %w from HTTP POST", err)
	}

	return rs.publishPayload(publisher, *res)
}

func (rs *RabbitMQServer) createAndPublishErrorResponse(errorArgs ConsumerError, publisher PublisherWrapper) error {
	errorPayload := req.Response{
		// todo something with headers
		Err: errorArgs.err,
		//Request: errorArgs.pl,
	}

	return rs.publishPayload(publisher, errorPayload)
}

func (rs *RabbitMQServer) publishPayload(publisher PublisherWrapper, pl req.Response) error {
	return publisher.Publish(pl, rs.OutputQueueName)
}
