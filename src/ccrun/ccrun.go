package ccrun

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ccrun",
	Short: "Coding Challenges Runtime",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
