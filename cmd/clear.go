/*
Copyright © 2025 NAME HERE adrian.mcmichael@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clears the named import folder",
	Long:  `Clears the specified import folder of all files and directories.`,
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, err := cmd.Flags().GetString("output")

		if err != nil {
			fmt.Printf("Error getting output directory: %v\n", err)
			return
		}

		if outputDir == "" {
			fmt.Println("Error: The --output flag is required")
			return
		}

		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			fmt.Printf("Error getting absolute path for %s: %v\n", outputDir, err)
			return
		}

		err = os.RemoveAll(absPath)
		if err != nil {
			fmt.Printf("Error clearing directory %s: %v\n", absPath, err)
			return
		}
		fmt.Printf("Successfully cleared directory: %s\n", absPath)
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)

	clearCmd.Flags().StringP("output", "o", "./exported/", "The output directory to clear")
}
