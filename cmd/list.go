package cmd

import (
	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var listGroups bool

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List requests in the current group",
	Long:    `List all saved requests in the active group. Use --groups to list all groups instead.`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		if listGroups {
			groups, err := store.ListGroups()
			if err != nil {
				ui.Error(err.Error())
				return
			}
			ui.PrintGroupList(groups, currentGroup)
			return
		}

		group := mustCurrentGroup()

		reqs, err := store.ListRequests(group)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		ui.PrintRequestList(group, reqs)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listGroups, "groups", false, "list all groups instead of requests")
}
