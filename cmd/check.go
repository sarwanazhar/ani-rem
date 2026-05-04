package cmd

import (
	"ani-rem/utils"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for airing anime and send notifications",
	Run: func(cmd *cobra.Command, args []string) {
		utils.CheckAiringAnime()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
