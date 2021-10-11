package cmd

import (
	"fmt"

	"github.com/kylec725/graytorrent/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var (
	removeCmd = &cobra.Command{
		Use:   "rm",
		Short: "removes a managed torrent",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.Remove(args[0]); err != nil {
				fmt.Println(err)
			}
		},
	}
)
