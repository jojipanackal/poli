package cmd

import (
	"fmt"

	curlpkg "github.com/jojipanackal/poli/internal/curl"
	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var showCurl bool

var showCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a saved request",
	Long: `Display a saved request in a structured format.
Use --curl flag to output as a curl command.

Examples:
  poli show "Get Users"
  poli show "Get Users" --curl`,
	Args:    cobra.ExactArgs(1),
	GroupID: "management",
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		group := mustCurrentGroup()

		req, err := store.FindRequest(group, name)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		if showCurl {
			fmt.Println()
			fmt.Println("  " + curlpkg.Generate(req))
			fmt.Println()
		} else {
			ui.PrintRequest(req)
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVar(&showCurl, "curl", false, "output as curl command")
}
