package internal

import (
	"context"
	"errors"
	"github.com/adrian-mcmichael/pocket-obsidian-migrator/internal/logger"
	"github.com/gocolly/colly"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/url"
	"os"
	"time"
)

type CrawlResult struct {
	RawLink
	Success bool   `csv:"success"`
	Error   string `csv:"error,omitempty"`
}

type PocketCrawler struct {
	convertor    *MarkdownConverter
	writer       *MarkdownWriter
	links        *Links
	crawlResults []CrawlResult
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"

// NewPocketCrawler initializes a new PocketCrawler.
func NewPocketCrawler(baseFolder string) (*PocketCrawler, error) {
	writer, err := NewMarkdownWriter(baseFolder)
	if err != nil {
		return nil, err
	}

	return &PocketCrawler{
		convertor:    NewMarkdownConverter(),
		writer:       writer,
		crawlResults: []CrawlResult{},
	}, nil
}

func (c *PocketCrawler) ImportLinks(ctx context.Context, linksFile string) ([]CrawlResult, error) {
	log := logger.Logger(ctx)

	links := &Links{}
	if err := links.ImportFrom(ctx, linksFile); err != nil {
		return nil, err
	}
	c.links = links

	log.Debug("Found links", zap.Int("count", len(c.links.Links)))

	if c.links == nil || len(c.links.Links) == 0 {
		log.Warn("No links to visit")
		return nil, nil // No links to visit
	}

	g, groupCtx := errgroup.WithContext(context.Background())

	for _, link := range c.links.Links {
		g.Go(func() error {
			return c.handleLink(groupCtx, link)
		})
	}

	if err := g.Wait(); err != nil {
		log.Error("Error visiting links", zap.Error(err))
		return nil, err
	}

	log.Debug("Finished visiting links", zap.Int("count", len(c.crawlResults)))

	return c.crawlResults, nil
}

func (c *PocketCrawler) handleLink(ctx context.Context, link Link) error {
	log := logger.Logger(ctx)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	log.Debug("Visiting link", zap.String("url", link.URL))

	err := c.visitPage(ctx, link)
	errValue := ""
	if err != nil {
		errValue = err.Error()
	}
	c.crawlResults = append(c.crawlResults, CrawlResult{
		RawLink: link.ToRawLink(),
		Success: err == nil,
		Error:   errValue,
	})

	return nil
}

func (c *PocketCrawler) visitPage(ctx context.Context, link Link) error {
	collector := colly.NewCollector(
		colly.UserAgent(userAgent),
	)
	collector.SetRequestTimeout(30 * time.Second)

	collector.OnError(func(r *colly.Response, err error) {
		c.handleError(ctx, err, link)
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		link.ProcessMetaTags(e)
		c.writeToFile(ctx, e, link)
	})

	err := collector.Visit(link.URL)
	if err != nil && !IsTimeoutError(err) {
		return err
	}

	return nil
}

func (c *PocketCrawler) handleError(ctx context.Context, err error, link Link) {
	log := logger.Logger(ctx)
	if IsTimeoutError(err) {
		log.Warn("Timeout error while visiting link", zap.String("url", link.URL), zap.Error(err))
	} else {
		log.Error("Error while visiting link", zap.String("url", link.URL), zap.Error(err))
	}
}

func (c *PocketCrawler) writeToFile(ctx context.Context, e *colly.HTMLElement, link Link) {
	log := logger.Logger(ctx)

	htmlContent, err := e.DOM.Html()
	if err != nil {
		log.Error("Error getting HTML content", zap.Error(err))
		return
	}
	markdownContent, err := c.convertor.ConvertToMarkdown(htmlContent)
	if err != nil {
		log.Error("Error converting HTML to Markdown", zap.Error(err))
		return
	}

	fileName, err := c.writer.WriteMarkdownFile(link, markdownContent)
	if err != nil {
		log.Error("Error writing Markdown file", zap.Error(err), zap.String("url", link.URL), zap.String("fileName", fileName))
		return
	} else {
		log.Debug("Successfully wrote Markdown file", zap.String("url", link.URL), zap.String("fileName", fileName))
	}
}

func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func IsTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(errors.Unwrap(err), context.DeadlineExceeded) {
		return true
	}
	if os.IsTimeout(err) {
		return true
	}
	if os.IsTimeout(errors.Unwrap(err)) {
		return true
	}
	return false
}
