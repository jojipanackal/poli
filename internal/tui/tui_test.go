package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
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
	for _, label := range []string{"EDIT REQUEST", "NAME", "URL", "HEADERS", "BODY"} {
		if !strings.Contains(view, label) {
			t.Fatalf("view is missing %q:\n%s", label, view)
		}
	}
}

func TestWrapTextKeepsLongResponsesInsideThePane(t *testing.T) {
	got := wrapText("abcdefghi", 4)
	if got != "abcd\nefgh\ni" {
		t.Fatalf("wrapText returned %q", got)
	}

	unicode := wrapText("你好世界", 2)
	if unicode != "你好\n世界" {
		t.Fatalf("wrapText split unicode text incorrectly: %q", unicode)
	}
}

func TestNarrowWorkbenchShowsOneBrowsePaneAtATime(t *testing.T) {
	m, err := newModel("__poli_tui_test_group__", "curl https://example.com")
	if err != nil {
		t.Fatalf("newModel returned error: %v", err)
	}
	m.screen = screenBrowse
	m.width = 80
	m.height = 24
	m.pane = paneList
	listView := m.View()
	if !strings.Contains(listView, "COLLECTION") || strings.Contains(listView, "REQUEST PREVIEW") {
		t.Fatalf("narrow list view is not isolated:\n%s", listView)
	}

	m.pane = paneEditor
	previewView := m.View()
	if !strings.Contains(previewView, "REQUEST PREVIEW") || strings.Contains(previewView, "COLLECTION") {
		t.Fatalf("narrow preview view is not isolated:\n%s", previewView)
	}
}

func TestActivePanelUsesHeavyBorder(t *testing.T) {
	m := &appModel{}
	view := m.panel("ACTIVE", "content", 30, 10, true)
	if !strings.Contains(view, "┃") {
		t.Fatalf("active panel does not use a visibly heavy border:\n%s", view)
	}
}

func TestWorkbenchFillsTheRequestedWidth(t *testing.T) {
	m, err := newModel("__poli_tui_test_group__", "curl https://example.com")
	if err != nil {
		t.Fatalf("newModel returned error: %v", err)
	}
	m.screen = screenBrowse
	m.pane = paneList
	view := m.renderWorkbench(120, 20)
	if got := lipgloss.Width(view); got != 120 {
		t.Fatalf("workbench width is %d, want 120", got)
	}
}

func TestFormatJSONBody(t *testing.T) {
	got, err := formatJSONBody(`{"name":"poli","enabled":true}`)
	if err != nil {
		t.Fatalf("formatJSONBody returned error: %v", err)
	}
	if got != "{\n  \"name\": \"poli\",\n  \"enabled\": true\n}" {
		t.Fatalf("unexpected formatted body:\n%s", got)
	}
	if _, err := formatJSONBody(`{"broken":}`); err == nil {
		t.Fatal("formatJSONBody accepted invalid JSON")
	}
}

func TestFormatResponseLeavesPlainTextReadable(t *testing.T) {
	got := formatResponse(204, "204 No Content", nil, "done", 0, false)
	if !strings.Contains(got, "BODY\ndone") {
		t.Fatalf("plain-text body was not preserved:\n%s", got)
	}
}
