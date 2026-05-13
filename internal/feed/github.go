package feed

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ghsaURL = "https://api.github.com/advisories"

type GitHubFeed struct {
	client *http.Client
}

func NewGitHubFeed() *GitHubFeed {
	return &GitHubFeed{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GitHubFeed) Name() string { return "github" }

func (g *GitHubFeed) Fetch(ctx context.Context, since time.Time) ([]Entry, error) {
	return nil, fmt.Errorf("github feed: requires API token or uses OSV for GHSA data")
}

func mapGitHubEcosystem(eco string) string {
	switch eco {
	case "npm":
		return "npm"
	case "pip":
		return "pip"
	default:
		return ""
	}
}

var _ = strings.TrimSpace
var _ = http.MethodGet
