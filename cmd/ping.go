package cmd

import (
	"fmt"
	"time"

	httppkg "github.com/jojipanackal/poli/internal/http"
	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	pingShowHeaders bool
	pingShowFull    bool
	pingExpandKey   string
	pingShowRaw     bool
	pingRowIndex    int
	pingSearchQuery string
)

var pingCmd = &cobra.Command{
	Use:   "ping [name]",
	Short: "Execute a saved request",
	Long: `Send the HTTP request and display the response.

By default shows only the status code and response body.
Use flags to see more details:

  --headers   Show response headers
  --full      Show full response (status + headers + body)
  --expand    Expand a nested key in the response body
  --raw       Show raw JSON response instead of table
  --row       Expand a specific row in an array response (1-indexed)
  --search    Filter array by full/partial value or key=value

Examples:
  poli ping "Get Users"
  poli ping "Get Users" --headers
  poli ping "Get Users" --full
  poli ping "Get Users" --expand address
  poli ping "Get Users" --raw
  poli ping "Get All Posts" --row 12
  poli ping "Get All Posts" --search "userId=5"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		group := mustCurrentGroup()

		req, err := store.FindRequest(group, name)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		ui.Info(fmt.Sprintf("%s %s", req.Method, req.URL))

		resp, err := httppkg.Execute(req)
		if err != nil {
			ui.Error(fmt.Sprintf("Request failed: %s", err))
			return
		}

		// Save response for `poli last`
		flatHeaders := make(map[string]string)
		for k, vals := range resp.Headers {
			if len(vals) > 0 {
				flatHeaders[k] = vals[0]
			}
		}
		store.SaveResponse(group, req.Name, store.SavedResponse{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    flatHeaders,
			Body:       resp.Body,
			DurationMs: resp.Duration.Milliseconds(),
			Timestamp:  time.Now(),
		})

		// Determine what to show
		opts := ui.DefaultResponseOptions()
		opts.RequestName = req.Name
		opts.ExpandKey = pingExpandKey
		opts.ShowRaw = pingShowRaw
		opts.RowIndex = pingRowIndex
		opts.SearchQuery = pingSearchQuery
		if pingShowFull {
			opts = ui.FullResponseOptions()
			opts.RequestName = req.Name
			opts.ExpandKey = pingExpandKey
			opts.ShowRaw = pingShowRaw
			opts.RowIndex = pingRowIndex
			opts.SearchQuery = pingSearchQuery
		} else if pingShowHeaders {
			opts.ShowHeaders = true
		}

		ui.PrintResponseCompact(resp.StatusCode, resp.Status, resp.Headers, resp.Body, resp.Duration, opts)
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	pingCmd.Flags().BoolVar(&pingShowHeaders, "headers", false, "show response headers")
	pingCmd.Flags().BoolVar(&pingShowFull, "full", false, "show full response (status + headers + body)")
	pingCmd.Flags().StringVar(&pingExpandKey, "expand", "", "expand a nested key in the response body")
	pingCmd.Flags().BoolVar(&pingShowRaw, "raw", false, "show raw JSON response instead of table")
	pingCmd.Flags().IntVar(&pingRowIndex, "row", 0, "expand a specific row in an array response (1-indexed)")
	pingCmd.Flags().StringVar(&pingSearchQuery, "search", "", "filter array items by full/partial value or key=value")
}
