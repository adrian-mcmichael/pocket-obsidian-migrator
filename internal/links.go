package internal

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/adrian-mcmichael/pocket-obsidian-migrator/internal/logger"
	"github.com/gocarina/gocsv"
	"github.com/gocolly/colly"
	"go.uber.org/zap"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Links struct {
	Links []Link `json:"links,omitempty"`
}

type Link struct {
	Title     string            `json:"title,omitempty" csv:"title"`
	URL       string            `json:"url,omitempty" csv:"url"`
	TimeAdded time.Time         `json:"time_added,omitempty" csv:"time_added"`
	Tags      []string          `json:"tags,omitempty" csv:"tags"`
	Status    string            `json:"status,omitempty" csv:"status"`
	Meta      map[string]string `json:"meta,omitempty" csv:"meta"`
}

func (l *Link) String() string {
	return fmt.Sprintf("Title: %s, URL: %s, TimeAdded: %s, Tags: %v, Status: %s", l.Title, l.URL, l.TimeAdded.Format(time.RFC3339), l.Tags, l.Status)
}

func (l *Link) TitleValue() string {
	if l.Meta != nil {
		if title, ok := l.Meta["og:title"]; ok && title != "" {
			return html.EscapeString(title)
		}
		if title, ok := l.Meta["title"]; ok && title != "" {
			return html.EscapeString(title)
		}
	}
	return html.EscapeString(l.Title)
}

func (l *Link) Author() string {
	if l.Meta != nil {
		if author, ok := l.Meta["article:author"]; ok && author != "" {
			return author
		}
	}
	return "unknown"
}

func (l *Link) Description() string {
	if l.Meta != nil {
		if description, ok := l.Meta["description"]; ok && description != "" {
			return description
		}
		if description, ok := l.Meta["og:description"]; ok && description != "" {
			return description
		}
	}
	return "none"
}

func (l *Link) PublishedTime() time.Time {
	if l.Meta != nil {
		if published, ok := l.Meta["article:published_time"]; ok && published != "" {
			t, err := time.Parse(time.RFC3339, published)
			if err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func (l *Link) ProcessMetaTags(e *colly.HTMLElement) {
	meta := make(map[string]string)
	e.DOM.Find("meta").Each(func(i int, s *goquery.Selection) {
		meta[s.AttrOr("name", s.AttrOr("property", ""))] = s.AttrOr("content", "")
	})

	title := e.DOM.Find("title").Text()
	if title != "" && !IsURL(title) {
		meta["title"] = title
	} else {
		e.DOM.Find("h1").Each(func(i int, s *goquery.Selection) {
			if title := s.Text(); title != "" {
				meta["title"] = title
				return // Stop after finding the first h1 tag
			}
		})

		if meta["title"] == "" {
			meta["title"] = html.EscapeString(l.Title) // Fallback to link title if no title tag found
		}
	}

	l.Meta = meta
}

func (l *Link) ToRawLink() RawLink {
	tags := ""
	if len(l.Tags) > 0 {
		tags = strings.Join(l.Tags, "|")
	}

	return RawLink{
		Title:     l.TitleValue(),
		URL:       l.URL,
		TimeAdded: l.TimeAdded.Unix(),
		Tags:      tags,
		Status:    l.Status,
	}
}

type RawLink struct {
	Title     string `json:"title,omitempty" csv:"title"`
	URL       string `json:"url,omitempty" csv:"url"`
	TimeAdded int64  `json:"time_added,omitempty" csv:"time_added"`
	Tags      string `json:"tags,omitempty" csv:"tags"`
	Status    string `json:"status,omitempty" csv:"status"`
}

func (r *RawLink) ToLink() Link {
	timeAdded := time.Unix(r.TimeAdded, 0)

	link := Link{
		Title:     r.Title,
		URL:       r.URL,
		TimeAdded: timeAdded,
		Status:    r.Status,
	}

	if r.Tags != "" {
		link.Tags = []string{}
		tags := strings.Split(r.Tags, "|")
		link.Tags = append(link.Tags, tags...)
	}

	return link
}

func (l *Links) ImportFrom(ctx context.Context, path string) error {
	log := logger.Logger(ctx)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error getting absolute path for %s: %w", path, err)
	}

	log.Debug("Importing links from file", zap.String("path", absPath))

	linksFile, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", absPath, err)
	}

	defer func(linksFile *os.File) {
		_ = linksFile.Close()
	}(linksFile)

	rawLinks := make([]RawLink, 0)
	if err := gocsv.UnmarshalFile(linksFile, &rawLinks); err != nil {
		return fmt.Errorf("error unmarshalling CSV file %s: %w", absPath, err)
	}

	for _, raw := range rawLinks {
		l.Links = append(l.Links, raw.ToLink())
	}

	return nil
}
