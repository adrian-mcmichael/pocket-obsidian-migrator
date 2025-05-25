package internal

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/gocolly/colly"
	"os"
	"path/filepath"
)

type MarkdownWriter struct {
	baseFolder string
}

// NewMarkdownWriter initializes a new MarkdownWriter with the given file path.
func NewMarkdownWriter(baseFolder string) (*MarkdownWriter, error) {
	// If no base folder is provided, use the default "./imported"
	if baseFolder == "" {
		baseFolder = "./imported"
	}

	// Strip trailing slash if present
	if baseFolder[len(baseFolder)-1] == '/' {
		baseFolder = baseFolder[:len(baseFolder)-1]
	}

	absPath, err := filepath.Abs(baseFolder)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path for %s: %w", baseFolder, err)
	}

	// Ensure the clippings folder exists
	clippingsPath := fmt.Sprintf("%s/clippings", absPath)
	if err := os.MkdirAll(clippingsPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating base folder %s: %v\n", baseFolder, err)
	}

	return &MarkdownWriter{
		baseFolder: absPath,
	}, nil
}

// WriteMarkdownFile writes a new file based on the given Link and its content.
func (w *MarkdownWriter) WriteMarkdownFile(link Link, content string) (string, error) {
	markdownFileName := colly.SanitizeFileName(fmt.Sprintf("%s.md", link.TitleValue()))
	fileName := fmt.Sprintf("%s/clippings/%s", w.baseFolder, markdownFileName)

	file, err := os.Create(fileName)
	if err != nil {
		return fileName, fmt.Errorf("error creating file %s: %w", w.baseFolder, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Write file header
	err = w.writeFileHeader(link, file)
	if err != nil {
		return fileName, err
	}

	// Write the content to the file
	_, err = file.WriteString(content)
	if err != nil {
		return fileName, fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}

	return fileName, nil
}

func (w *MarkdownWriter) writeFileHeader(link Link, file *os.File) error {
	_, err := file.WriteString("---\n")
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}

	_, err = file.WriteString(fmt.Sprintf("title: \"%s\"\n", link.TitleValue()))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	_, err = file.WriteString(fmt.Sprintf("source: \"%s\"\n", link.URL))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	_, err = file.WriteString(fmt.Sprintf("author: \n  - \"%s\"\n", link.Author()))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	if !link.PublishedTime().IsZero() {
		_, err = file.WriteString(fmt.Sprintf("published: %s\n", link.PublishedTime().Format("2006-01-02")))
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
		}
	}
	_, err = file.WriteString(fmt.Sprintf("created: %s\n", link.TimeAdded.Format("2006-01-02")))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	if link.Description() != "" {
		_, err = file.WriteString(fmt.Sprintf("description: \"%s\"\n", link.Description()))
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
		}
	}
	_, err = file.WriteString("tags:\n")
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	_, err = file.WriteString(fmt.Sprintf("  - \"clippings\"\n"))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	_, err = file.WriteString(fmt.Sprintf("  - \"pocket\"\n"))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	for _, tag := range link.Tags {
		_, err = file.WriteString(fmt.Sprintf("  - \"%s\"\n", tag))
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
		}
	}
	_, err = file.WriteString("---\n\n")
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", w.baseFolder, err)
	}
	return nil
}

type ResultsWriter struct {
	outputPath string
}

// NewResultsWriter initializes a new ResultsWriter with the given output path.
func NewResultsWriter(outputPath string) (*ResultsWriter, error) {
	if outputPath == "" {
		outputPath = "./failed.csv"
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating directory %s: %v\n", dir, err)
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path for %s: %w", outputPath, err)
	}

	return &ResultsWriter{
		outputPath: absPath,
	}, nil
}

// WriteResults writes the given results to the output file.
func (w *ResultsWriter) WriteResults(results []CrawlResult) error {
	file, err := os.Create(w.outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", w.outputPath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	failureCount := 0
	for _, result := range results {
		if !result.Success {
			failureCount++
		}
	}
	successCount := len(results) - failureCount

	fmt.Println("Total URLs Crawled:", len(results))
	fmt.Println("Successful URLs:", successCount)
	fmt.Println("Failed URLs:", failureCount)

	failed := make([]CrawlResult, 0, failureCount)
	for _, result := range results {
		if !result.Success {
			failed = append(failed, result)
		}
	}

	if failureCount > 0 {
		err = gocsv.MarshalFile(failed, file)
		if err != nil {
			return fmt.Errorf("error writing results to file %s: %w", w.outputPath, err)
		}
	}

	return nil
}
