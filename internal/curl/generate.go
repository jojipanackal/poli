package curl

import (
	"fmt"
	"strings"

	"github.com/jojipanackal/poli/internal/model"
)

// Generate converts a model.Request back into a curl command string.
func Generate(req model.Request) string {
	var parts []string
	parts = append(parts, "curl")

	if req.Method != "" && req.Method != "GET" {
		parts = append(parts, "-X", req.Method)
	}

	for _, h := range req.Headers {
		parts = append(parts, "-H", fmt.Sprintf("'%s: %s'", h.Key, h.Value))
	}

	if req.Body != "" {
		// Escape single quotes in body
		escaped := strings.ReplaceAll(req.Body, "'", "'\\''")
		parts = append(parts, "-d", fmt.Sprintf("'%s'", escaped))
	}

	if req.URL != "" {
		parts = append(parts, fmt.Sprintf("'%s'", req.URL))
	}

	return strings.Join(parts, " ")
}
