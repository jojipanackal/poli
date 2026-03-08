package cmd

import (
	"fmt"

	"github.com/jojipanackal/poli/internal/ui"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print the version number of poli",
	Long:    `All software has versions. This is poli's`,
	GroupID: "utility",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintLogo()
		fmt.Printf("poli v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
