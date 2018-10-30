package service

import (
	spartaEvents "github.com/mweagle/Sparta/aws/events"
)

// Todo is the type we return
type Todo struct {
	ID        string `json:"-"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Order     int    `json:"order"`
	URL       string `json:"url"`
}

// TodoRequest is the typed input to the
// Post handler
type TodoRequest struct {
	spartaEvents.APIGatewayEnvelope
	Body Todo `json:"body"`
}
