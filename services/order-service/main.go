
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

		ch, err := conn.Channel()
		if err != nil {
			log.Printf("Failed to open a channel: %s", err)
			http.Error(w, "Failed to open a channel", http.StatusInternalServerError)
			return
		}
		defer ch.Close()

		replyQueue, err := ch.QueueDeclare(
			"",    // name (amqp-assigned)
			false, // durable
			true,  // delete when unused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Printf("Failed to declare a reply queue: %s", err)
			http.Error(w, "Failed to declare a reply queue", http.StatusInternalServerError)
			return
		}

		msgs, err := ch.Consume(
			replyQueue.Name, // queue
			"",              // consumer
			true,            // auto-ack
			false,           // exclusive
			false,           // no-local
			false,           // no-wait
			nil,             // args
		)
		if err != nil {
			log.Printf("Failed to register a consumer: %s", err)
			http.Error(w, "Failed to register a consumer", http.StatusInternalServerError)
			return
		}

		correlationID := uuid.New().String()

		event := common.CreateFulfillmentEvent{
			CorrelationID: correlationID,
			ReplyTo:       replyQueue.Name,
			Order:         order,
		}

		body, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %s", err)
			http.Error(w, "Failed to marshal event", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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
		if err != nil {
			log.Printf("Failed to publish a message: %s", err)
			http.Error(w, "Failed to publish a message", http.StatusInternalServerError)
			return
		}

		// Loop with a timeout to find our message. Although the queue is exclusive
		// and we should only get one message, this makes the handler more robust.
		for {
			select {
			case d := <-msgs:
				// Check if the message is the one we're waiting for
				if d.CorrelationId == correlationID {
					var response common.FulfillmentResponse
					if err := json.Unmarshal(d.Body, &response); err != nil {
						log.Printf("Error decoding response: %s", err)
						http.Error(w, "Error decoding response", http.StatusInternalServerError)
					} else {
						if response.CanFulfillOrder {
							log.Printf("Order can be fulfilled.")
							w.WriteHeader(http.StatusCreated)
						} else {
							log.Printf("Order cannot be fulfilled.")
							w.WriteHeader(http.StatusUnprocessableEntity)
						}
					}
					// Our work is done, so we can return from the handler.
					return
				} else {
					log.Printf("Received message with wrong correlation ID on exclusive queue. Expected %s, got %s. Discarding.", correlationID, d.CorrelationId)
					// This is unexpected, but we'll loop again to wait for the correct message
				}
			case <-ctx.Done():
				// The request context timed out
				log.Printf("Request timed out.")
				w.WriteHeader(http.StatusRequestTimeout)
				return
			}
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
