package cmd

import (
	"fmt"
	"os"

	tuiapp "github.com/jojipanackal/poli/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCurlImport string

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Short:   "Open the interactive request workbench",
	GroupID: "request",
	Args:    cobra.NoArgs,
	Long: `Open a keyboard-first request workbench for the active group.

Use it to import cURL commands, edit requests, execute them, and inspect the
last response without leaving the terminal.

Examples:
  poli tui
  poli tui --curl 'curl https://api.example.com/users'`,
	Run: func(cmd *cobra.Command, args []string) {
		group := mustCurrentGroup()
		if err := tuiapp.Run(group, tuiCurlImport); err != nil {
			fmt.Fprintf(os.Stderr, "poli tui: %s\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().StringVar(&tuiCurlImport, "curl", "", "start with a cURL command")
}
