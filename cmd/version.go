package cmd

import (
	"fmt"

	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of poli",
	Long:  `All software has versions. This is poli's`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintLogo()
		fmt.Println("poli version 0.0.1-alpha")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
