package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/jojipanackal/poli/internal/model"
)

var (
	green  = color.New(color.FgGreen, color.Bold)
	blue   = color.New(color.FgBlue, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	cyan   = color.New(color.FgCyan)
	dim    = color.New(color.Faint)
	bold   = color.New(color.Bold)
	white  = color.New(color.FgWhite, color.Bold)
)

// Success prints a green success message.
func Success(msg string) {
	green.Print("✓ ")
	fmt.Println(msg)
}

// Info prints a blue info message.
func Info(msg string) {
	blue.Print("→ ")
	fmt.Println(msg)
}

// Error prints a red error message.
func Error(msg string) {
	red.Print("✗ ")
	fmt.Println(msg)
}

// Warning prints a yellow warning message.
func Warning(msg string) {
	yellow.Print("! ")
	fmt.Println(msg)
}

// ResponseOptions controls what parts of a response to show.
type ResponseOptions struct {
	ShowHeaders bool
	ShowBody    bool
	ShowStatus  bool
	ShowRaw     bool   // show raw JSON instead of tables
	ExpandKey   string // key to expand in JSON body
	RequestName string // request name for expand hints
	RowIndex    int    // expand a specific row in array responses (1-indexed, 0 = off)
	SearchQuery string // filter array items by key=value
}

// DefaultResponseOptions shows just status + body (clean minimal view).
func DefaultResponseOptions() ResponseOptions {
	return ResponseOptions{
		ShowStatus:  true,
		ShowBody:    true,
		ShowHeaders: false,
	}
}

// FullResponseOptions shows everything.
func FullResponseOptions() ResponseOptions {
	return ResponseOptions{
		ShowStatus:  true,
		ShowBody:    true,
		ShowHeaders: true,
	}
}

// PrintResponseCompact renders a clean, minimal HTTP response (status + body).
func PrintResponseCompact(statusCode int, status string, headers map[string][]string, body string, duration fmt.Stringer, opts ResponseOptions) {
	fmt.Println()

	// Status line — always shown
	if opts.ShowStatus {
		statusColor := green
		if statusCode >= 400 {
			statusColor = red
		} else if statusCode >= 300 {
			statusColor = yellow
		}

		fmt.Print("  ")
		statusColor.Printf("%-4d", statusCode)
		fmt.Print(" ")
		dim.Print(statusText(statusCode))
		fmt.Print("  ")
		dim.Println(duration)
	}

	// Headers — only with --headers or --full
	if opts.ShowHeaders && len(headers) > 0 {
		fmt.Println()
		cyan.Println("  Headers")
		printHeadersTable(headers)
	}

	// Body — shown by default
	if opts.ShowBody && body != "" {
		fmt.Println()
		printBodyAsTable(body, "  ", opts)
	}

	fmt.Println()
}

// PrintResponseFromSaved renders a saved response with the same formatting.
func PrintResponseFromSaved(statusCode int, status string, headers map[string]string, body string, durationMs int64, opts ResponseOptions) {
	fmt.Println()

	if opts.ShowStatus {
		statusColor := green
		if statusCode >= 400 {
			statusColor = red
		} else if statusCode >= 300 {
			statusColor = yellow
		}

		fmt.Print("  ")
		statusColor.Printf("%-4d", statusCode)
		fmt.Print(" ")
		dim.Print(statusText(statusCode))
		fmt.Print("  ")
		dim.Printf("%dms\n", durationMs)
	}

	if opts.ShowHeaders && len(headers) > 0 {
		fmt.Println()
		cyan.Println("  Headers")
		printSavedHeadersTable(headers)
	}

	if opts.ShowBody && body != "" {
		fmt.Println()
		printBodyAsTable(body, "  ", opts)
	}

	fmt.Println()
}

// PrintRequest renders a request in a structured readable format.
func PrintRequest(req model.Request) {
	fmt.Println()
	bold.Printf("  %s ", req.Name)
	dim.Printf("(%s)\n", req.UpdatedAt.Format("Jan 02, 2006 3:04 PM"))
	fmt.Println()

	printField("Method", methodColor(req.Method))
	printField("URL", req.URL)

	if len(req.Headers) > 0 {
		cyan.Print("  Headers\n")
		for _, h := range req.Headers {
			dim.Printf("    %s", h.Key)
			fmt.Printf(": %s\n", h.Value)
		}
	}

	if req.Body != "" {
		cyan.Print("  Body\n")
		printBodyAsTable(req.Body, "    ", ResponseOptions{RequestName: req.Name})
	}

	fmt.Println()
}

// PrintGroupList renders a list of groups.
func PrintGroupList(groups []model.Group, currentGroup string) {
	if len(groups) == 0 {
		Warning("No groups yet. Create one with: poli new group \"Name\"")
		return
	}

	fmt.Println()
	bold.Println("  Groups")
	fmt.Println()

	for i, g := range groups {
		marker := "  "
		if strings.EqualFold(g.Name, currentGroup) {
			marker = green.Sprint("▸ ")
		}

		fmt.Printf("  %2d.", i+1)
		fmt.Printf(" %s %s", marker, g.Name)
		dim.Printf("  %s\n", g.CreatedAt.Format("Jan 02, 2006"))
	}
	fmt.Println()
}

// PrintRequestList renders a list of requests in a table.
func PrintRequestList(group string, reqs []model.Request) {
	if len(reqs) == 0 {
		Warning(fmt.Sprintf("No requests in %q. Create one with: poli new \"Name\"", group))
		return
	}

	fmt.Println()
	bold.Printf("  %s", group)
	dim.Printf("  (%d requests)\n", len(reqs))
	fmt.Println()

	for i, r := range reqs {
		dim.Printf("  %2d.", i+1)
		fmt.Printf("  %s %s", methodColor(r.Method), r.Name)
		dim.Printf("  %s\n", truncateURL(r.URL, 60))
	}
	fmt.Println()
}

func printField(label, value string) {
	cyan.Printf("  %s  ", label)
	fmt.Println(value)
}

func methodColor(method string) string {
	padded := fmt.Sprintf("%-7s", method)
	switch method {
	case "GET":
		return green.Sprint(padded)
	case "POST":
		return blue.Sprint(padded)
	case "PUT":
		return yellow.Sprint(padded)
	case "PATCH":
		return yellow.Sprint(padded)
	case "DELETE":
		return red.Sprint(padded)
	default:
		return white.Sprint(padded)
	}
}

func statusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Found"
	case 304:
		return "Not Modified"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 409:
		return "Conflict"
	case 422:
		return "Unprocessable Entity"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return ""
	}
}

// printHeadersTable renders HTTP headers as a tabular key-value table.
func printHeadersTable(headers map[string][]string) {
	// Find max key width
	maxKeyLen := 0
	for k := range headers {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	if maxKeyLen > 30 {
		maxKeyLen = 30
	}

	tableWidth := maxKeyLen + 4 + 50
	fmt.Printf("  %s\n", dim.Sprint(strings.Repeat("─", tableWidth)))

	for k, vals := range headers {
		for _, v := range vals {
			paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)
			fmt.Print("  ")
			bold.Printf("%s", paddedKey)
			fmt.Printf("    %s\n", truncate(v, 50))
		}
	}

	fmt.Printf("  %s\n", dim.Sprint(strings.Repeat("─", tableWidth)))
}

// printSavedHeadersTable renders saved headers (flat map) as a table.
func printSavedHeadersTable(headers map[string]string) {
	maxKeyLen := 0
	for k := range headers {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	if maxKeyLen > 30 {
		maxKeyLen = 30
	}

	tableWidth := maxKeyLen + 4 + 50
	fmt.Printf("  %s\n", dim.Sprint(strings.Repeat("─", tableWidth)))

	for k, v := range headers {
		paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)
		fmt.Print("  ")
		bold.Printf("%s", paddedKey)
		fmt.Printf("    %s\n", truncate(v, 50))
	}

	fmt.Printf("  %s\n", dim.Sprint(strings.Repeat("─", tableWidth)))
}

// printRawBody prints JSON body with pretty indentation (raw format).
func printRawBody(body, indent string) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(body), indent, "  "); err == nil {
		lines := strings.Split(pretty.String(), "\n")
		for _, line := range lines {
			fmt.Printf("%s%s\n", indent, line)
		}
	} else {
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			fmt.Printf("%s%s\n", indent, line)
		}
	}
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// truncateURL shortens a URL for display, keeping host+path and truncating query params.
func truncateURL(rawURL string, maxLen int) string {
	if len(rawURL) <= maxLen {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return truncate(rawURL, maxLen)
	}

	// Build base: scheme://host/path
	base := u.Scheme + "://" + u.Host + u.Path

	if u.RawQuery == "" {
		return truncate(base, maxLen)
	}

	// Show base + truncated query
	remaining := maxLen - len(base) - 1 // -1 for "?"
	if remaining <= 3 {
		return base + "?..."
	}
	return base + "?" + truncate(u.RawQuery, remaining)
}

func PrintLogo() {
	logo := `
    ____        ___
   / __ \____  / (_)
  / /_/ / __ \/ / / 
 / ____/ /_/ / / /  
/_/    \____/_/_/   `

	color.Cyan(logo)
	fmt.Println()
}
