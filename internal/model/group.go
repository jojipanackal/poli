package model

import "time"

// Group represents a collection of requests (like a Postman folder).
type Group struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
