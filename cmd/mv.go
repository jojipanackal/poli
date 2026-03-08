package cmd

import (
	"fmt"

	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv [request] [target-group]",
	Short: "Move a request to another group",
	Long: `Move a request from the current group to a different group.

Supports index (r1, r2, ...) or name for the request.
Supports index (g1, g2, ...) or name for the target group.

Examples:
  poli mv "Get Users" "Other Group"
  poli mv r1 g2
  poli mv "Get Users" g2
  poli mv r1 "Other Group"`,
	Args:    cobra.ExactArgs(2),
	GroupID: "management",
	Run: func(cmd *cobra.Command, args []string) {
		reqArg := args[0]
		groupArg := args[1]

		sourceGroup := mustCurrentGroup()

		// Resolve request
		req, err := store.FindRequest(sourceGroup, reqArg)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		// Resolve target group
		targetGroup, err := store.FindGroup(groupArg)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		if targetGroup.Name == sourceGroup {
			ui.Warning("Request is already in that group")
			return
		}

		// Save to target group
		if err := store.SaveRequest(targetGroup.Name, req); err != nil {
			ui.Error(fmt.Sprintf("Failed to save to \"%s\": %s", targetGroup.Name, err))
			return
		}

		// Delete from source group
		if err := store.DeleteRequest(sourceGroup, req.Name); err != nil {
			ui.Error(fmt.Sprintf("Saved to \"%s\" but failed to remove from \"%s\": %s", targetGroup.Name, sourceGroup, err))
			return
		}

		fmt.Println()
		ui.Success(fmt.Sprintf("Moved \"%s\" → \"%s\"", req.Name, targetGroup.Name))
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
