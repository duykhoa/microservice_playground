
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"microservice-playground/services/internal/common"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

// A map to store response channels, keyed by correlationID
var (
	responseChannels = make(map[string]chan common.FulfillmentResponse)
	mapMutex         = &sync.Mutex{}
)

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	replyQueue, err := ch.QueueDeclare(
		"",    // name (amqp-assigned)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	common.FailOnError(err, "Failed to declare a reply queue")

	// Start a single consumer for the reply queue
	go consumeReplies(ch, replyQueue.Name)

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var order common.Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Received order: %+v", order)

		correlationID := uuid.New().String()
		responseChan := make(chan common.FulfillmentResponse)

		mapMutex.Lock()
		responseChannels[correlationID] = responseChan
		mapMutex.Unlock()

		defer func() {
			mapMutex.Lock()
			delete(responseChannels, correlationID)
			mapMutex.Unlock()
		}()

		event := common.CreateFulfillmentEvent{
			CorrelationID: correlationID,
			ReplyTo:       replyQueue.Name,
			Order:         order,
		}

		body, err := json.Marshal(event)
		common.FailOnError(err, "Failed to marshal event")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = ch.PublishWithContext(ctx,
			"fulfillment_exchange", // exchange
			"fulfillment_requests", // routing key
			false,                  // mandatory
			false,                  // immediate
			amqp091.Publishing{
				ContentType:   "application/json",
				CorrelationId: correlationID,
				ReplyTo:       replyQueue.Name,
				Body:          body,
			})
		common.FailOnError(err, "Failed to publish a message")

		select {
		case response := <-responseChan:
			if response.CanFulfillOrder {
				log.Printf("Order can be fulfilled.")
				w.WriteHeader(http.StatusCreated)
			} else {
				log.Printf("Order cannot be fulfilled.")
				w.WriteHeader(http.StatusUnprocessableEntity)
			}
		case <-ctx.Done():
			log.Printf("Request timed out.")
			w.WriteHeader(http.StatusRequestTimeout)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func consumeReplies(ch *amqp091.Channel, queueName string) {
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	common.FailOnError(err, "Failed to register a consumer")

	for d := range msgs {
		mapMutex.Lock()
		responseChan, ok := responseChannels[d.CorrelationId]
		mapMutex.Unlock()

		if ok {
			var response common.FulfillmentResponse
			if err := json.Unmarshal(d.Body, &response); err != nil {
				log.Printf("Error decoding response: %s", err)
			} else {
				responseChan <- response
			}
		} else {
			log.Printf("Received message with unknown correlation ID: %s", d.CorrelationId)
		}
	}
}
