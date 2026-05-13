package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const socketURL = "https://socket.dev/supply-chain-attacks/mini-shai-hulud"

type SocketFeed struct {
	client *http.Client
}

func NewSocketFeed() *SocketFeed {
	return &SocketFeed{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SocketFeed) Name() string { return "socket" }

func (s *SocketFeed) Fetch(ctx context.Context, since time.Time) ([]Entry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, socketURL, nil)
	if err != nil {
		return nil, fmt.Errorf("socket feed: %w", err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "LibKill/1.0 (+https://github.com/firfircelik/libkill)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("socket feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("socket feed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("socket feed read: %w", err)
	}

	return parseSocketHTML(string(body), since), nil
}

func parseSocketHTML(html string, since time.Time) []Entry {
	var entries []Entry
	seen := map[string]bool{}

	lines := strings.Split(html, "\n")
	for _, line := range lines {
		for _, eco := range []string{"npm", "pypi"} {
			prefix := eco + " "
			idx := strings.Index(line, prefix)
			if idx == -1 {
				continue
			}

			rest := strings.TrimSpace(line[idx+len(prefix):])
			parts := strings.Fields(rest)
			if len(parts) < 2 {
				continue
			}
			pkg := parts[0]
			ver := parts[1]

			if !strings.Contains(ver, ".") {
				continue
			}
			ver = strings.TrimRight(ver, ")]")

			key := eco + ":" + pkg + ":" + ver
			if seen[key] {
				continue
			}
			seen[key] = true

			ecosystem := "npm"
			if eco == "pypi" {
				ecosystem = "pip"
			}

			entries = append(entries, Entry{
				Package:   pkg,
				Version:   ver,
				Ecosystem: ecosystem,
				Feed:      "socket",
				Severity:  "critical",
				Reason:    "Mini Shai-Hulud supply-chain attack",
				Detected:  time.Now().UTC().Truncate(time.Second),
			})
		}
	}

	return entries
}
