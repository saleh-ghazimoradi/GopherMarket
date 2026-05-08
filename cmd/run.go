package cmd

import (
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")

		sLogger := logger.NewSlogLogger()

		cfg, err := config.GetConfigInstance()
		if err != nil {
			sLogger.Error("Failed to get config instance", "err", err)
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
