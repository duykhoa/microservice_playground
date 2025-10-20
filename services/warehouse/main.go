
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"microservice-playground/services/internal/common"

	"github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"fulfillment_exchange", // name
		"direct",               // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	common.FailOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"fulfillment_requests", // name
		false,                  // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,                 // queue name
		"fulfillment_requests", // routing key
		"fulfillment_exchange", // exchange
		false,
		nil,
	)
	common.FailOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	common.FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var event common.CreateFulfillmentEvent
			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Printf("Error decoding message: %s", err)
				continue
			}

			// Simulate processing and make a random decision
			rand.Seed(time.Now().UnixNano())
			canFulfill := rand.Intn(2) == 1

			response := common.FulfillmentResponse{
				CorrelationID:   event.CorrelationID,
				CanFulfillOrder: canFulfill,
			}

			body, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error encoding response: %s", err)
				continue
			}

			err = ch.Publish(
				"",          // exchange
				event.ReplyTo, // routing key (the reply-to queue)
				false,       // mandatory
				false,       // immediate
				amqp091.Publishing{
					ContentType:   "application/json",
					CorrelationId: event.CorrelationID,
					Body:          body,
				})

			if err != nil {
				log.Printf("Failed to publish reply: %s", err)
			} else {
				log.Printf("Replied with canFulfill: %v", canFulfill)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
