package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"microservice-playground/services/internal/common"

	"github.com/rabbitmq/amqp091-go"
)

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Order struct {
	ID    string    `json:"id"`
	Items []Product `json:"items"`
}

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"orders", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Received order: %+v", order)

		body, err := json.Marshal(order)
		common.FailOnError(err, "Failed to marshal order")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = ch.PublishWithContext(ctx,
			"wms",  // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		common.FailOnError(err, "Failed to publish a message")

		w.WriteHeader(http.StatusCreated)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
