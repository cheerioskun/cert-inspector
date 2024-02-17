package main

import (
	"io/fs"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A brief description of your application",
	Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This is where you'd put the main logic for your app
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "A short description of graph",
	Long:  `A longer description of graph`,
	Run: func(cmd *cobra.Command, args []string) {
		// This is where you'd put the logic for the graph command
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(graphCmd)
	os.Mkdir("/tmp/.cert-inspector", fs.ModeDir|0766)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger := log15.New()
		logger.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}