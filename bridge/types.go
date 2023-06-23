package bridge

import (
	"github.com/imroc/req/v3"
	"github.com/wagslane/go-rabbitmq"
	"io"
)

const (
	CONTENT_TYPE_JSON = "application/json"
)

// wrapper around `rabbitmq.Conn`
type ConnectionWrapper interface {
	io.Closer

	NewPublisher() (PublisherWrapper, error)
	NewConsumer(handler rabbitmq.Handler, queue string, consumerTag string) (ConsumerWrapper, error)
}

type ConsumerWrapper interface {
	Close()
}

type PublisherWrapper interface {
	Close()
	Publish(
		payload req.Response,
		queueName string,
	) error
}

type ConsumerError struct {
	err      error
	delivery rabbitmq.Delivery
}
