package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const unit42FeedURL = "https://unit42.paloaltonetworks.com/feed/"

type Unit42Feed struct {
	client *http.Client
}

type unit42RSS struct {
	Channel struct {
		Items []unit42Item `xml:"item"`
	} `xml:"channel"`
}

type unit42Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Category    []string `xml:"category"`
}

func NewUnit42Feed() *Unit42Feed {
	return &Unit42Feed{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (u *Unit42Feed) Name() string { return "unit42" }

func (u *Unit42Feed) Fetch(ctx context.Context) ([]MalwareFamilyEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, unit42FeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unit42 feed: %w", err)
	}
	req.Header.Set("User-Agent", "LibKill/1.0 (+https://github.com/firfircelik/libkill)")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unit42 feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unit42 feed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("unit42 feed read: %w", err)
	}

	var rss unit42RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("unit42 parse: %w", err)
	}

	var entries []MalwareFamilyEntry
	for _, item := range rss.Channel.Items {
		text := item.Title + " " + item.Description
		name := extractMalwareName(text)
		if name == "" {
			continue
		}

		entries = append(entries, MalwareFamilyEntry{
			Name:        name,
			Type:        inferMalwareType(text),
			Platform:    inferPlatform(text),
			Discovered:  parsePubDate(item.PubDate),
			Attribution: attributionFromCats(item.Category),
			Description: truncate(cleanDescription(item.Description), 500),
			SourceURL:   item.Link,
		})
	}

	return entries, nil
}

func attributionFromCats(cats []string) string {
	for _, c := range cats {
		lower := strings.ToLower(c)
		if strings.Contains(lower, "apt") || strings.Contains(lower, "china") ||
			strings.Contains(lower, "iran") || strings.Contains(lower, "russia") ||
			strings.Contains(lower, "north korea") || strings.Contains(lower, "belarus") ||
			strings.Contains(lower, "threat group") || strings.Contains(lower, "group") {
			return c
		}
	}
	return "See Unit 42 report"
}
