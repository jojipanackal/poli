package curl

import (
	"strings"
	"time"

	"github.com/jojipanackal/poli/internal/model"
)

// Parse converts a curl command string into a model.Request.
func Parse(name, raw string) (model.Request, error) {
	req := model.Request{
		Name:      name,
		Method:    "GET",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Strip leading "curl " if present
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "curl ") {
		raw = raw[5:]
	} else if strings.HasPrefix(raw, "curl\t") {
		raw = raw[5:]
	}

	tokens := tokenize(raw)

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		switch {
		case t == "-X" || t == "--request":
			if i+1 < len(tokens) {
				i++
				req.Method = strings.ToUpper(tokens[i])
			}

		case t == "-H" || t == "--header":
			if i+1 < len(tokens) {
				i++
				parts := strings.SplitN(tokens[i], ":", 2)
				if len(parts) == 2 {
					req.Headers = append(req.Headers, model.Header{
						Key:   strings.TrimSpace(parts[0]),
						Value: strings.TrimSpace(parts[1]),
					})
				}
			}

		case t == "-d" || t == "--data" || t == "--data-raw" || t == "--data-binary":
			if i+1 < len(tokens) {
				i++
				req.Body = tokens[i]
				// If method is still GET and we have a body, assume POST
				if req.Method == "GET" {
					req.Method = "POST"
				}
			}

		case !strings.HasPrefix(t, "-"):
			// Positional arg = URL
			url := strings.Trim(t, "'\"")
			if url != "" && req.URL == "" {
				req.URL = url
			}
		}
	}

	return req, nil
}

// tokenize splits a curl command into tokens, respecting quoted strings.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, r := range s {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
			escaped = true
			continue
		}

		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if (r == ' ' || r == '\t' || r == '\n') && !inSingle && !inDouble {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
