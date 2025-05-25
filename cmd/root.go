/*
Copyright © 2025 NAME HERE adrian.mcmichael@gmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pocket-obsidian-migrator",
	Short: "A application to migrate Pocket articles to Obsidian",
	Long: `A application to migrate Pocket articles to Obsidian.
This application takes a Pocket export file and converts the articles into markdown files for use in Obsidian.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
