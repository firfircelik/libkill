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

const talosFeedURL = "https://blog.talosintelligence.com/feed"

type TalosFeed struct {
	client *http.Client
}

type talosRSS struct {
	Channel struct {
		Items []talosItem `xml:"item"`
	} `xml:"channel"`
}

type talosItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func NewTalosFeed() *TalosFeed {
	return &TalosFeed{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (t *TalosFeed) Name() string { return "talos" }

func (t *TalosFeed) Fetch(ctx context.Context) ([]MalwareFamilyEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, talosFeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("talos feed: %w", err)
	}
	req.Header.Set("User-Agent", "LibKill/1.0 (+https://github.com/firfircelik/libkill)")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("talos feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("talos feed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("talos feed read: %w", err)
	}

	var rss talosRSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("talos parse: %w", err)
	}

	var entries []MalwareFamilyEntry
	for _, item := range rss.Channel.Items {
		name := extractMalwareName(item.Title + " " + item.Description)
		if name == "" {
			continue
		}

		entries = append(entries, MalwareFamilyEntry{
			Name:        name,
			Type:        inferMalwareType(item.Title + " " + item.Description),
			Platform:    inferPlatform(item.Title + " " + item.Description),
			Discovered:  parsePubDate(item.PubDate),
			Attribution: "See Cisco Talos report",
			Description: truncate(cleanDescription(item.Description), 500),
			SourceURL:   item.Link,
		})
	}

	return entries, nil
}

func parsePubDate(date string) string {
	layouts := []string{
		time.RFC1123Z, time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, date); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return date
}

func extractMalwareName(text string) string {
	indicators := []string{
		"malware", "backdoor", "RAT", "ransomware",
		"trojan", "stealer", "botnet", "rootkit",
		"wiper", "spyware", "keylogger", "dropper",
	}
	textLower := strings.ToLower(text)

	found := false
	for _, ind := range indicators {
		if strings.Contains(textLower, ind) {
			found = true
			break
		}
	}
	if !found {
		return ""
	}

	quoted := extractQuoted(text)
	if quoted != "" {
		return quoted
	}

	return "Talos Threat (" + time.Now().Format("2006-01-02") + ")"
}

func extractQuoted(text string) string {
	for _, q := range []string{`"`, `'`} {
		start := strings.Index(text, q)
		if start >= 0 {
			end := strings.Index(text[start+1:], q)
			if end > 0 {
				return text[start+1 : start+1+end]
			}
		}
	}

	words := strings.Fields(text)
	for i, w := range words {
		if i > 0 && strings.Contains(w, "malware") || strings.Contains(w, "backdoor") ||
			strings.Contains(w, "RAT") || strings.Contains(w, "ransomware") {
			if i >= 1 {
				return strings.TrimRight(words[i-1], ".:,!;")
			}
		}
	}
	return ""
}

func inferMalwareType(text string) string {
	textLower := strings.ToLower(text)
	types := map[string]string{
		"ransomware": "ransomware",
		"backdoor":   "backdoor",
		"rat":        "rat",
		"trojan":     "trojan",
		"stealer":    "stealer",
		"botnet":     "botnet",
		"rootkit":    "rootkit",
		"wiper":      "wiper",
		"worm":       "worm",
		"keylogger":  "keylogger",
		"dropper":    "dropper",
		"spyware":    "spyware",
	}
	var found []string
	for k, v := range types {
		if strings.Contains(textLower, k) {
			found = append(found, v)
		}
	}
	if len(found) > 0 {
		return strings.Join(found, ",")
	}
	return "unknown"
}

func inferPlatform(text string) string {
	textLower := strings.ToLower(text)
	platforms := map[string]string{
		"windows": "windows",
		"linux":   "linux",
		"macos":   "macos",
		"mac os":  "macos",
		"android": "android",
		"ios":     "ios",
		"esxi":    "esxi",
	}
	var found []string
	for k, v := range platforms {
		if strings.Contains(textLower, k) {
			found = append(found, v)
		}
	}
	if len(found) > 0 {
		return strings.Join(unique(found), ",")
	}
	return "windows"
}

func cleanDescription(desc string) string {
	for _, tag := range []string{"<p>", "</p>", "<br/>", "<br />", "<br>", "<strong>", "</strong>", "<em>", "</em>", "<code>", "</code>", "<ul>", "</ul>", "<li>", "</li>", "<ol>", "</ol>"} {
		desc = strings.ReplaceAll(desc, tag, " ")
	}
	desc = strings.Join(strings.Fields(desc), " ")
	return desc
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func unique(items []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
