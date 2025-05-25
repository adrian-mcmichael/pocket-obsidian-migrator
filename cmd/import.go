/*
Copyright © 2025 NAME HERE adrian.mcmichael@gmail.com
*/
package cmd

import (
	"fmt"
	"github.com/adrian-mcmichael/pocket-obsidian-migrator/internal"
	"github.com/adrian-mcmichael/pocket-obsidian-migrator/internal/logger"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports a Pocket export file into Obsidian markdown files",
	Long: `Given a Pocket export file, this command will convert the articles into markdown files
into a specified directory for use in Obsidian. It will apply any tags on the articles as Obsidian tags and will
attempt to scrape and include the article's content in the markdown file.`,
	Run: func(cmd *cobra.Command, args []string) {
		importFile := cmd.Flag("file").Value.String()
		outputDir := cmd.Flag("output").Value.String()
		if importFile == "" {
			fmt.Println("Error: The --file flag is required")
			return
		}

		verbose, err := cmd.Flags().GetBool("verbose")

		var logLevel string
		if verbose {
			logLevel = "debug"
		} else {
			logLevel = "fatal"
		}

		l := logger.Get(logLevel)
		ctx := logger.Attach(cmd.Context(), l)

		crawler, err := internal.NewPocketCrawler(outputDir)
		if err != nil {
			fmt.Printf("Error creating Pocket crawler: %v\n", err)
			return
		}

		fmt.Println(fmt.Sprintf("Importing links from Pocket export file %s...", importFile))

		results, err := crawler.ImportLinks(ctx, importFile)
		if err != nil {
			fmt.Printf("Error importing links: %v\n", err)
			return
		}

		resultsWriter, err := internal.NewResultsWriter(fmt.Sprintf("%s/failed.csv", outputDir))
		if err != nil {
			fmt.Printf("Error initializing results writer: %v\n", err)
		}

		if err := resultsWriter.WriteResults(results); err != nil {
			fmt.Printf("Error writing results: %v\n", err)
		}

		fmt.Println(fmt.Sprintf("All links visited and markdown files created at %s", outputDir))
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("file", "f", "", "Path to the Pocket export file (required)")
	err := importCmd.MarkFlagRequired("file")
	if err != nil {
		fmt.Println(err)
	}

	importCmd.Flags().StringP("output", "o", "./exported/", "Directory to save the Obsidian markdown files")
	importCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}
