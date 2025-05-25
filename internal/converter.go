package internal

import (
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
)

type MarkdownConverter struct {
	converter *converter.Converter
}

// NewMarkdownConverter initializes a new MarkdownConverter with default options.
func NewMarkdownConverter() *MarkdownConverter {
	c := converter.NewConverter(converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
	))

	return &MarkdownConverter{
		converter: c,
	}
}

func (m *MarkdownConverter) ConvertToMarkdown(htmlContent string) (string, error) {
	return m.converter.ConvertString(htmlContent)
}
