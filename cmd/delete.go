package cmd

import (
	"fmt"

	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var deleteGroup bool

var deleteCmd = &cobra.Command{
	Use:   "delete [name|r1|g1]",
	Short: "Delete a request or group",
	Long: `Delete a saved request from the current group.
Use --group flag to delete a group instead.

Supports index (r1, r2, ...) or name for requests.
Supports index (g1, g2, ...) or name for groups.

Examples:
  poli delete "Get Users"
  poli delete r1
  poli delete --group "My API"
  poli delete --group g1`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if deleteGroup {
			// Delete a group
			g, err := store.FindGroup(name)
			if err != nil {
				ui.Error(err.Error())
				return
			}

			fmt.Println()
			ui.Warning(fmt.Sprintf("This will permanently delete group \"%s\" and all its requests.", g.Name))
			if !ui.PromptConfirm("Delete?") {
				ui.Info("Cancelled")
				return
			}

			if err := store.DeleteGroup(g.Name); err != nil {
				ui.Error(err.Error())
				return
			}

			// Clear current group if it was the deleted one
			if currentGroup == g.Name {
				setCurrentGroup("")
			}

			fmt.Println()
			ui.Success(fmt.Sprintf("Deleted group \"%s\"", g.Name))
			return
		}

		// Delete a request
		group := mustCurrentGroup()

		req, err := store.FindRequest(group, name)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		fmt.Println()
		ui.Warning(fmt.Sprintf("Delete \"%s\" (%s %s) from %s?", req.Name, req.Method, req.URL, group))
		if !ui.PromptConfirm("Delete?") {
			ui.Info("Cancelled")
			return
		}

		if err := store.DeleteRequest(group, req.Name); err != nil {
			ui.Error(err.Error())
			return
		}

		fmt.Println()
		ui.Success(fmt.Sprintf("Deleted \"%s\"", req.Name))
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVar(&deleteGroup, "group", false, "delete a group instead of a request")
}
