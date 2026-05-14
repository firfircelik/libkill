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

const bleepingFeedURL = "https://www.bleepingcomputer.com/feed/"

type BleepingFeed struct {
	client *http.Client
}

type bleepingRSS struct {
	Channel struct {
		Items []bleepingItem `xml:"item"`
	} `xml:"channel"`
}

type bleepingItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Category    []string `xml:"category"`
}

func NewBleepingFeed() *BleepingFeed {
	return &BleepingFeed{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *BleepingFeed) Name() string { return "bleeping" }

func (b *BleepingFeed) Fetch(ctx context.Context) ([]MalwareFamilyEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bleepingFeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bleeping feed: %w", err)
	}
	req.Header.Set("User-Agent", "LibKill/1.0 (+https://github.com/firfircelik/libkill)")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bleeping feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bleeping feed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("bleeping feed read: %w", err)
	}

	var rss bleepingRSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("bleeping parse: %w", err)
	}

	var entries []MalwareFamilyEntry
	for _, item := range rss.Channel.Items {
		if !isSecurityCategory(item.Category) {
			continue
		}
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
			Attribution: "See BleepingComputer report",
			Description: truncate(cleanDescription(item.Description), 500),
			SourceURL:   item.Link,
		})
	}

	return entries, nil
}

func isSecurityCategory(cats []string) bool {
	for _, c := range cats {
		lower := strings.ToLower(c)
		if strings.Contains(lower, "malware") || strings.Contains(lower, "ransomware") ||
			strings.Contains(lower, "virus") || strings.Contains(lower, "security") ||
			strings.Contains(lower, "hack") || strings.Contains(lower, "cyber") ||
			strings.Contains(lower, "vulnerability") {
			return true
		}
	}
	return true
}
