package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	currentGroup string
)

var rootCmd = &cobra.Command{
	Use:   "poli",
	Short: "Poli — fast terminal HTTP client",
	Long:  `Terminal-based HTTP client — collections, requests, curl import, zero lag.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load current group from config if not provided via flag
		if currentGroup == "" {
			currentGroup = viper.GetString("current_group")
		}

		// Assign automatic commands to utility group
		for _, c := range cmd.Root().Commands() {
			if c.Name() == "completion" || c.Name() == "help" {
				c.GroupID = "utility"
			}
		}
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddGroup(&cobra.Group{ID: "request", Title: "Request Operations:"})
	rootCmd.AddGroup(&cobra.Group{ID: "management", Title: "Collection Management:"})
	rootCmd.AddGroup(&cobra.Group{ID: "utility", Title: "Utility Commands:"})

	rootCmd.PersistentFlags().StringVarP(&currentGroup, "group", "g", "", "active group/collection (default from config)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.poli/config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".poli")
		os.MkdirAll(configDir, 0755)
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig() // ignore error — first run has no file
}
