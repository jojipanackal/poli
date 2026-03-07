package model

import "time"

// Header represents a single HTTP header key-value pair.
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Request represents a saved HTTP request.
type Request struct {
	Name      string   `json:"name"`
	Method    string   `json:"method"`
	URL       string   `json:"url"`
	Headers   []Header `json:"headers,omitempty"`
	Body      string   `json:"body,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
