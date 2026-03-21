package cmd

import (
	"fmt"
	"time"

	"github.com/jojipanackal/poli/internal/jsonutil"
	"github.com/jojipanackal/poli/internal/model"
	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a saved request",
	Long: `Interactively edit the fields of a saved request.
Supports index (r1, r2, ...) or name.
Press Enter to keep the current value.

Examples:
  poli edit "Get Users"
  poli edit r1`,
	Args:    cobra.ExactArgs(1),
	GroupID: "request",
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		group := mustCurrentGroup()

		req, err := store.FindRequest(group, name)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		fmt.Println()
		ui.Info(fmt.Sprintf("Editing \"%s\" — press Enter to keep current value", req.Name))
		fmt.Println()

		// Method
		req.Method = ui.PromptMethod(req.Method)

		// URL
		req.URL = ui.PromptString("URL", req.URL)

		// Headers
		existingHeaders := make([]struct{ Key, Value string }, len(req.Headers))
		for i, h := range req.Headers {
			existingHeaders[i] = struct{ Key, Value string }{h.Key, h.Value}
		}

		if ui.PromptConfirm("Edit headers?") {
			headerPairs := ui.PromptHeaders(existingHeaders)
			var headers []model.Header
			for _, h := range headerPairs {
				headers = append(headers, model.Header{Key: h.Key, Value: h.Value})
			}
			req.Headers = headers
		}

		// Body
		if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
			if ui.PromptConfirm("Edit body?") {
				if req.Body != "" {
					ui.Info("Current body:")
					fmt.Printf("    %s\n\n", req.Body)
				}
				body := ui.PromptMultiline("New body")
				fixed, modified := jsonutil.AutoFixJSON(body)
				if modified {
					ui.Success("Auto-fixed single quotes to double quotes for valid JSON")
					req.Body = fixed
				} else {
					req.Body = body
				}
			}
		}

		req.UpdatedAt = time.Now()

		if err := store.SaveRequest(group, req); err != nil {
			ui.Error(err.Error())
			return
		}

		fmt.Println()
		ui.Success(fmt.Sprintf("Updated \"%s\"", req.Name))
		ui.Info(fmt.Sprintf("%s %s", req.Method, req.URL))
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
