package tui

import (
	"strings"
	"testing"
	"time"
)

func TestParseHeaders(t *testing.T) {
	headers, err := parseHeaders("Authorization: Bearer token\nContent-Type: application/json\nX-Trace: a:b")
	if err != nil {
		t.Fatalf("parseHeaders returned error: %v", err)
	}
	if len(headers) != 3 {
		t.Fatalf("got %d headers, want 3", len(headers))
	}
	if headers[2].Value != "a:b" {
		t.Fatalf("colon in header value was not preserved: %q", headers[2].Value)
	}
}

func TestParseHeadersRejectsMalformedLine(t *testing.T) {
	_, err := parseHeaders("Accept: application/json\nnot-a-header")
	if err == nil {
		t.Fatal("parseHeaders accepted a malformed line")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("error does not identify the line: %v", err)
	}
}

func TestFormatResponsePrettyPrintsJSONAndSortsHeaders(t *testing.T) {
	got := formatResponse(
		200,
		"200 OK",
		map[string]string{"X-Zed": "last", "Content-Type": "application/json"},
		`{"message":"ok","items":[1,2]}`,
		125*time.Millisecond,
		true,
	)

	for _, expected := range []string{
		"200 OK  125ms",
		"Content-Type: application/json\nX-Zed: last",
		"{\n  \"message\": \"ok\",\n  \"items\": [\n    1,\n    2\n  ]\n}",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("response output is missing %q:\n%s", expected, got)
		}
	}
}

func TestNewModelSupportsInitialCurlAndRendersPanels(t *testing.T) {
	m, err := newModel("__poli_tui_test_group__", "curl -X POST -H 'Content-Type: application/json' -d '{\"ok\":true}' https://example.com")
	if err != nil {
		t.Fatalf("newModel returned error: %v", err)
	}
	if m.current.Method != "POST" || m.current.URL != "https://example.com" {
		t.Fatalf("cURL was not imported into the request: %#v", m.current)
	}
	if m.screen != screenEdit || !m.nameInput.Focused() {
		t.Fatalf("initial cURL did not open the focused editor")
	}

	m.width = 120
	m.height = 30
	view := m.View()
	for _, label := range []string{"COLLECTION", "REQUEST", "RESPONSE"} {
		if !strings.Contains(view, label) {
			t.Fatalf("view is missing %q:\n%s", label, view)
		}
	}
}

func TestFormatResponseLeavesPlainTextReadable(t *testing.T) {
	got := formatResponse(204, "204 No Content", nil, "done", 0, false)
	if !strings.Contains(got, "BODY\ndone") {
		t.Fatalf("plain-text body was not preserved:\n%s", got)
	}
}
