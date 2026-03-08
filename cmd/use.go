package cmd

import (
	"fmt"
	"os"

	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:     "use [group]",
	Short:   "Switch to a different group/collection",
	Long:    `Set the active group so subsequent commands operate on it.`,
	Aliases: []string{"switch"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string

		if len(args) == 1 {
			name = args[0]
			
			// Resolve index if it's a number
			var idx int
			if _, err := fmt.Sscanf(name, "g%d", &idx); err == nil {
				groups, _ := store.ListGroups()
				if idx > 0 && idx <= len(groups) {
					name = groups[idx-1].Name
				}
			}
		} else {
			// Interactive selection
			groups, err := store.ListGroups()
			if err != nil || len(groups) == 0 {
				ui.Error("No groups found. Create one first:")
				fmt.Println("  poli new group \"Name\"")
				os.Exit(1)
			}

			options := make([]string, len(groups))
			for i, g := range groups {
				options[i] = g.Name
			}

			idx, _ := ui.PromptSelect("Select a group", options)
			name = groups[idx].Name
		}

		if !store.GroupExists(name) {
			ui.Error("Group \"" + name + "\" not found")
			return
		}

		setCurrentGroup(name)
		ui.Success("Switched to \"" + name + "\"")
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
