package bridge

import (
	"github.com/imroc/req/v3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/wagslane/go-rabbitmq"
	"testing"
)

func TestStringMapTableFunctions(t *testing.T) {
	origTable1 := amqp.Table{
		"key1": "value1",
		"key2": 45,
	}
	derivedStringMap1 := map[string]string{
		"key1": "value1",
		"key2": "45",
	}
	stringMap1 := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {"45"},
	}
	derivedTable1 := rabbitmq.Table{
		"key1": "value1",
		"key2": "45",
	}

	t.Run("TableToStringMap", func(t *testing.T) {
		mappedOrigTable1 := TableToStringMap(origTable1)
		assert.Equal(t, derivedStringMap1, mappedOrigTable1)
	})

	t.Run("StringMapToTable", func(t *testing.T) {
		mappedStringMap1 := StringMapToTable(stringMap1)
		assert.Equal(t, derivedTable1, mappedStringMap1)
	})
}

func TestDeliveryToPayload(t *testing.T) {
	bytesBody := []byte(`{"status":{"status":0},"strData":"\"hello\""}`)
	testDeliveryRest := rabbitmq.Delivery{
		Delivery: amqp.Delivery{
			Body:            bytesBody,
			ContentType:     CONTENT_TYPE_JSON,
			ContentEncoding: "",
			// TODO headers
		},
	}

	t.Run("rest payload", func(t *testing.T) {
		pl := DeliveryToPayload(testDeliveryRest)

		request := req.C().R()
		pl(request)

		assert.Equal(t, bytesBody, request.Body)
		// TODO headers, content type / encoding
	})
}
