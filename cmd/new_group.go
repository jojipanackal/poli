/*
Copyright © 2026 Joji Panackal jojijospanackal@gmail.com
*/
package cmd

import (
	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group [name]",
	Short: "Create a new collection/group",
	Long:  `Create a new group to organize your API requests, like folders in Postman.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := store.CreateGroup(name); err != nil {
			ui.Error(err.Error())
			return
		}

		setCurrentGroup(name)
		ui.Success("Created group \"" + name + "\"")
		ui.Info("Switched to \"" + name + "\"")
	},
}

func init() {
	newCmd.AddCommand(groupCmd)
}
