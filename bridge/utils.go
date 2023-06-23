package bridge

import (
	"fmt"
	"github.com/imroc/req/v3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wagslane/go-rabbitmq"
)

func TableToStringMap(t amqp.Table) map[string]string {
	stringMap := make(map[string]string)
	for key, value := range t {
		stringMap[key] = fmt.Sprintf("%v", value)
	}
	return stringMap
}

func StringMapToTable(m map[string][]string) rabbitmq.Table {
	table := make(map[string]interface{})
	for key, values := range m {
		// just take the first value, at least for now
		table[key] = values[0]
	}
	return table
}

func DeliveryToPayload(delivery rabbitmq.Delivery) func(*req.Request) {
	return func(request *req.Request) {
		request.SetBody(delivery.Body)
		if delivery.Headers != nil {
			request.SetHeaders(TableToStringMap(delivery.Headers))
		}
		if delivery.ContentType != "" {
			request.SetHeader("Content-Type", delivery.ContentType)
		}
		if delivery.ContentEncoding != "" {
			request.SetHeader("Content-Encoding", delivery.ContentEncoding)
		}
	}
}
