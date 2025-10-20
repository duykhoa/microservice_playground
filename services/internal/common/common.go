
package common

import "log"

type CreateFulfillmentEvent struct {
	CorrelationID string `json:"correlation_id"`
	ReplyTo       string `json:"reply_to"`
	Order         Order  `json:"order"`
}

type FulfillmentResponse struct {
	CorrelationID   string `json:"correlation_id"`
	CanFulfillOrder bool   `json:"can_fulfill_order"`
}

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Order struct {
	ID    string    `json:"id"`
	Items []Product `json:"items"`
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
