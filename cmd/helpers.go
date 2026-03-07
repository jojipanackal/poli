package cmd

import (
	"fmt"
	"os"

	"github.com/jojipanackal/poli/internal/store"
	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/viper"
)

func poliHome() string {
	return store.PoliHome()
}

func setCurrentGroup(name string) {
	viper.Set("current_group", name)

	// Ensure config file exists before writing
	cfgFile := viper.ConfigFileUsed()
	if cfgFile == "" {
		cfgFile = poliHome() + "/config.yaml"
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		viper.SafeWriteConfig()
	}
	currentGroup = name
}

func mustCurrentGroup() string {
	if currentGroup == "" {
		// Try to let user select a group
		groups, err := store.ListGroups()
		if err != nil || len(groups) == 0 {
			ui.Error("No group selected. Create one first:")
			fmt.Println("  poli new group \"Name\"")
			os.Exit(1)
		}

		options := make([]string, len(groups))
		for i, g := range groups {
			options[i] = g.Name
		}

		idx, _ := ui.PromptSelect("Select a group", options)
		setCurrentGroup(groups[idx].Name)
		return groups[idx].Name
	}
	return currentGroup
}
