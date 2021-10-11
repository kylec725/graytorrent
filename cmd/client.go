package cmd

import (
	"log"

	"github.com/kylec725/gray/internal/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "adds a new torrent to be managed",
		// Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.Add(); err != nil {
				log.Fatal(err)
			}
		},
	}
)
