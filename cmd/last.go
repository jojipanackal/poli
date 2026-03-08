package cmd

import (
	"fmt"

	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	lastShowHeaders bool
	lastShowFull    bool
	lastExpandKey   string
	lastShowRaw     bool
	lastRowIndex    int
	lastSearchQuery string
)

var lastCmd = &cobra.Command{
	Use:   "last [name]",
	Short: "Show the last response for a request",
	Long: `Retrieve and display the last saved response from a previous ping.

By default shows only the status code and response body.
Use flags to see more details:

  --headers   Show response headers
  --full      Show full response (status + headers + body)
  --expand    Expand a nested key in the response body
  --raw       Show raw JSON response instead of table
  --row       Expand a specific row in an array response (1-indexed)
  --search    Filter array by full/partial value or key=value

Examples:
  poli last "Get Users"
  poli last "Get Users" --headers
  poli last "Get Users" --full
  poli last "Get Users" --expand address
  poli last "Get Users" --raw
  poli last "Get All Posts" --row 12
  poli last "Get All Posts" --search "userId=5"`,
	Args:      cobra.ExactArgs(1),
	GroupID:   "management",
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		group := mustCurrentGroup()

		// Verify request exists
		req, err := store.FindRequest(group, name)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		resp, err := store.LoadResponse(group, req.Name)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		ui.Info(fmt.Sprintf("Last response for \"%s\"  %s", req.Name, resp.Timestamp.Format("Jan 02 3:04 PM")))

		// Determine what to show
		opts := ui.DefaultResponseOptions()
		opts.RequestName = req.Name
		opts.ExpandKey = lastExpandKey
		opts.ShowRaw = lastShowRaw
		opts.RowIndex = lastRowIndex
		opts.SearchQuery = lastSearchQuery
		if lastShowFull {
			opts = ui.FullResponseOptions()
			opts.RequestName = req.Name
			opts.ExpandKey = lastExpandKey
			opts.ShowRaw = lastShowRaw
			opts.RowIndex = lastRowIndex
			opts.SearchQuery = lastSearchQuery
		} else if lastShowHeaders {
			opts.ShowHeaders = true
		}

		ui.PrintResponseFromSaved(resp.StatusCode, resp.Status, resp.Headers, resp.Body, resp.DurationMs, opts)
	},
}

func init() {
	rootCmd.AddCommand(lastCmd)

	lastCmd.Flags().BoolVar(&lastShowHeaders, "headers", false, "show response headers")
	lastCmd.Flags().BoolVar(&lastShowFull, "full", false, "show full response (status + headers + body)")
	lastCmd.Flags().StringVar(&lastExpandKey, "expand", "", "expand a nested key in the response body")
	lastCmd.Flags().BoolVar(&lastShowRaw, "raw", false, "show raw JSON response instead of table")
	lastCmd.Flags().IntVar(&lastRowIndex, "row", 0, "expand a specific row in an array response (1-indexed)")
	lastCmd.Flags().StringVar(&lastSearchQuery, "search", "", "filter array items by full/partial value or key=value")
}
