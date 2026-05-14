package feed

import (
	"context"
	"time"
)

type Ecosystem string

const (
	EcoNPM Ecosystem = "npm"
	EcoPIP Ecosystem = "pip"
)

type Entry struct {
	Package   string    `json:"package"`
	Version   string    `json:"version"`
	Ecosystem string    `json:"ecosystem"`
	Feed      string    `json:"feed"`
	Severity  string    `json:"severity"`
	Reason    string    `json:"reason"`
	Detected  time.Time `json:"detected"`
	Platform  string    `json:"platform,omitempty"`
	CVE       string    `json:"cve,omitempty"`
	SourceURL string    `json:"source_url,omitempty"`
}

type Fetcher interface {
	Name() string
	Fetch(ctx context.Context, since time.Time) ([]Entry, error)
}

type MalwareFamilyEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Platform    string `json:"platform"`
	Discovered  string `json:"discovered"`
	Attribution string `json:"attribution"`
	Description string `json:"description"`
	CVE         string `json:"cve,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
}

type MalwareFetcher interface {
	Name() string
	Fetch(ctx context.Context) ([]MalwareFamilyEntry, error)
}
