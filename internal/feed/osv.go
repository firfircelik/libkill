package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const osvURL = "https://api.osv.dev/v1/querybatch"

type OSVFeed struct {
	client *http.Client
}

type osvQuery struct {
	Package  osvPackage `json:"package"`
	Version  string     `json:"version,omitempty"`
}

type osvPackage struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

type osvRequest struct {
	Queries []osvQuery `json:"queries"`
}

type osvResponse struct {
	Results []osvResult `json:"results"`
}

type osvResult struct {
	Vulns []osvVuln `json:"vulns"`
}

type osvVuln struct {
	ID       string   `json:"id"`
	Aliases  []string `json:"aliases"`
	Modified string   `json:"modified"`
}

func NewOSVFeed() *OSVFeed {
	return &OSVFeed{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (o *OSVFeed) Name() string { return "osv" }

func (o *OSVFeed) Fetch(ctx context.Context, since time.Time) ([]Entry, error) {
	return nil, fmt.Errorf("osv feed: requires package list for batch queries, use CheckPackage instead")
}

func (o *OSVFeed) CheckPackage(ctx context.Context, name, ecosystem string) ([]Entry, error) {
	eco := mapOSVEcosystem(ecosystem)
	if eco == "" {
		return nil, fmt.Errorf("osv feed: unsupported ecosystem %q", ecosystem)
	}

	reqBody := osvRequest{
		Queries: []osvQuery{
			{Package: osvPackage{Name: name, Ecosystem: eco}},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("osv marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, osvURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("osv: HTTP %d", resp.StatusCode)
	}

	var result osvResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("osv decode: %w", err)
	}

	var entries []Entry
	for _, r := range result.Results {
		for _, v := range r.Vulns {
			if !isMalwareVuln(v) {
				continue
			}
			entries = append(entries, Entry{
				Package:   name,
				Ecosystem: ecosystem,
				Feed:      "osv",
				Severity:  "critical",
				Reason:    fmt.Sprintf("OSV: %s", v.ID),
				Detected:  time.Now().UTC().Truncate(time.Second),
			})
		}
	}

	return entries, nil
}

func isMalwareVuln(v osvVuln) bool {
	for _, alias := range v.Aliases {
		if strings.HasPrefix(alias, "MAL-") {
			return true
		}
	}
	return false
}

func mapOSVEcosystem(eco string) string {
	switch eco {
	case "npm":
		return "npm"
	case "pip":
		return "PyPI"
	default:
		return ""
	}
}
