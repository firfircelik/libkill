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
}

type Fetcher interface {
	Name() string
	Fetch(ctx context.Context, since time.Time) ([]Entry, error)
}
