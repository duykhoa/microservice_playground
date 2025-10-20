
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"microservice-playground/services/internal/common"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
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

		// Wait for the response
		msgs, err := ch.Consume(
			replyQueue.Name, // queue
			"",              // consumer
			false,           // auto-ack
			false,           // exclusive
			false,           // no-local
			false,           // no-wait
			nil,             // args
		)
		common.FailOnError(err, "Failed to register a consumer")

		for d := range msgs {
			if d.CorrelationId == correlationID {
				var response common.FulfillmentResponse
				err := json.Unmarshal(d.Body, &response)
				if err != nil {
					log.Printf("Error decoding response: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if response.CanFulfillOrder {
					log.Printf("Order can be fulfilled.")
					w.WriteHeader(http.StatusCreated)
				} else {
					log.Printf("Order cannot be fulfilled.")
					w.WriteHeader(http.StatusUnprocessableEntity)
				}
				d.Ack(false)
				return // Exit after processing the correct message
			} else {
				log.Printf("Received message with wrong correlation ID. Nacking.")
				d.Nack(false, true) // Requeue for other consumers
			}
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
