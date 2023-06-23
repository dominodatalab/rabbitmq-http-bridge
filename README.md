# RabbitMQ HTTP Bridge

## Prototype

This simple prototype is based on the work done as a part of our Seldon-core RabbitMQ server.
It specifically uses code from [the branch to convert to the wagslane rabbitMQ 
client](https://github.com/dominodatalab/seldon-core/tree/rabbitmq_wagslane_client/executor/api/rabbitmq). 

## Unit tests

The <bridge/utils_test.go> test runs in IDEA, but not using `go test`.

## End to end test

### Test setup -- RabbitMQ

1. Port-forward RabbitMQ to local
   `kubectl -n <platform-namesapce> port-forward svc/rabbitmq-ha 5672:5672`
2. Retrieve rabbitmq creds for seldon user
   `kubectl get secrets -n <platform-namespace> rabbitmq-ha.seldon -ojsonpath='{.data.rabbitmq-password}' | base64 -D ; echo`
3. Create an async model API, but stop all running versions
   - note the model ID

### Test setup -- HTTP dummy server

1. Download [`dummyhttp`](https://github.com/svenstaro/dummyhttp) and place in your `PATH`
2. Run `dummyhttp` with a specific response code and body (it will respond to every request with this code & body)
   `dummyhttp -c 200 -b '{ "data": "some content" }' -vv`

### Test -- run RabbitMQ HTTP Bridge

Command:

`go run . "http://localhost:8080" "amqp://seldon:<seldon RMQ password>@localhost:5672" model-<modelId>-input-queue model-output-queue`

Four arguments after `go run .`:
- RabbitMQ broker URL (fill in <seldon RMQ password>)
- RabbitMQ input queue (fill in <modelId>)
- RabbitMQ output queue
- HTTP URL to relay the incoming messages to (using HTTP POST)

### Test -- send a message from RMQ web console
1. Port-forward RabbitMQ management port to local
   `kubectl -n <platform-namesapce> port-forward svc/rabbitmq-ha 15672:15672`
2. Retrieve rabbitmq creds for rabbitmq user
   `kubectl get secrets -n <platform-namespace> rabbitmq-ha.rabbitmq -ojsonpath='{.data.rabbitmq-password}' | base64 -D ; echo`
3. Go to http://localhost:15672 and login with user `rabbitmq` and password retrieved above
4. Go to Queues and then click on the name of your model input queue
5. Expand "Publish Message" 
6. Put some content in the Payload (e.g. `{ "data": "example" }`)
7. Click Publish Message

Now you should see the running `dummyhttp` server receive the incoming message and respond with its preconfigured response.

### Test -- observe output message failure in model-hosting
Tail the model hosting logs to see the invalid message on the output queue causing an error
`kubectl -n <platform namespace> logs <model hosting pod>`


