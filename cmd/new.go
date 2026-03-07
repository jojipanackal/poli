/*
Copyright © 2026 Joji Panackal jojijospanackal@gmail.com
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	curlpkg "github.com/jojipanackal/poli/internal/curl"
	"github.com/jojipanackal/poli/internal/model"
	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var curlImport string

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new request or group",
	Long: `Create a new request in the current group.
Use 'poli new group "Name"' to create a new group instead.

Examples:
  poli new "Get Users"
  poli new "Get Users" --curl 'curl https://api.example.com/users'`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		name := args[0]
		group := mustCurrentGroup()

		var req model.Request

		if curlImport != "" {
			// Import from curl
			parsed, err := curlpkg.Parse(name, curlImport)
			if err != nil {
				ui.Error(fmt.Sprintf("Failed to parse curl: %s", err))
				return
			}
			req = parsed
			ui.Success(fmt.Sprintf("Parsed curl → %s %s", req.Method, req.URL))
		} else {
			// Interactive form
			fmt.Println()
			ui.Info("Creating request \"" + name + "\"")
			fmt.Println()

			method := ui.PromptMethod("")
			url := ui.PromptString("URL", "")

			if url == "" {
				ui.Error("URL is required")
				return
			}

			// Headers
			headerPairs := ui.PromptHeaders(nil)
			var headers []model.Header
			for _, h := range headerPairs {
				headers = append(headers, model.Header{Key: h.Key, Value: h.Value})
			}

			// Body (for methods that support it)
			var body string
			if method == "POST" || method == "PUT" || method == "PATCH" {
				body = ui.PromptMultiline("Body")
			}

			req = model.Request{
				Name:      name,
				Method:    method,
				URL:       url,
				Headers:   headers,
				Body:      body,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}

		if err := store.SaveRequest(group, req); err != nil {
			ui.Error(err.Error())
			return
		}

		fmt.Println()
		ui.Success(fmt.Sprintf("Saved \"%s\" in %s", name, group))
		ui.Info(fmt.Sprintf("%s %s", req.Method, req.URL))

		if len(req.Headers) > 0 {
			headerStrs := make([]string, len(req.Headers))
			for i, h := range req.Headers {
				headerStrs[i] = h.Key
			}
			ui.Info(fmt.Sprintf("Headers: %s", strings.Join(headerStrs, ", ")))
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVar(&curlImport, "curl", "", "import request from a curl command")
}
