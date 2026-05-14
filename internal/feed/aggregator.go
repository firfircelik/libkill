package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/firfircelik/libkill/internal/db"
)

type Aggregator struct {
	store        *db.Store
	feeds        []Fetcher
	malwareFeeds []MalwareFetcher
}

func NewAggregator(store *db.Store) *Aggregator {
	a := &Aggregator{
		store: store,
		feeds: []Fetcher{
			NewSocketFeed(),
			NewOSVFeed(),
			NewGitHubFeed(),
		},
		malwareFeeds: []MalwareFetcher{
			NewTalosFeed(),
			NewUnit42Feed(),
			NewBleepingFeed(),
		},
	}

	if err := a.seedFromEmbedded(); err != nil {
		log.Printf("feed seed: %v", err)
	}

	if err := a.seedMalwareFromEmbedded(); err != nil {
		log.Printf("feed malware seed: %v", err)
	}

	return a
}

func (a *Aggregator) Update(ctx context.Context) (int, error) {
	total := 0

	for _, f := range a.feeds {
		lastSync, _ := a.store.LastSync(f.Name())

		entries, err := f.Fetch(ctx, lastSync)
		if err != nil {
			log.Printf("feed/%s: %v", f.Name(), err)
			continue
		}

		if len(entries) == 0 {
			continue
		}

		dbEntries := make([]db.ThreatEntry, len(entries))
		for i, e := range entries {
			dbEntries[i] = db.ThreatEntry{
				Package:   e.Package,
				Version:   e.Version,
				Ecosystem: e.Ecosystem,
				Feed:      e.Feed,
				Severity:  e.Severity,
				Reason:    e.Reason,
				Detected:  e.Detected,
				Platform:  e.Platform,
				CVE:       e.CVE,
				SourceURL: e.SourceURL,
			}
		}

		n, err := a.store.UpsertThreats(dbEntries)
		if err != nil {
			return total, fmt.Errorf("feed/%s store: %w", f.Name(), err)
		}
		total += n
	}

	return total, nil
}

func (a *Aggregator) UpdateMalware(ctx context.Context) (int, error) {
	total := 0

	for _, f := range a.malwareFeeds {
		entries, err := f.Fetch(ctx)
		if err != nil {
			log.Printf("malware-feed/%s: %v", f.Name(), err)
			continue
		}

		if len(entries) == 0 {
			continue
		}

		dbFamilies := make([]db.MalwareFamily, len(entries))
		for i, e := range entries {
			dbFamilies[i] = db.MalwareFamily{
				Name:        e.Name,
				Type:        e.Type,
				Platform:    e.Platform,
				Discovered:  e.Discovered,
				Attribution: e.Attribution,
				Description: e.Description,
				CVE:         e.CVE,
				SourceURL:   e.SourceURL,
			}
		}

		n, err := a.store.UpsertMalwareFamilies(dbFamilies)
		if err != nil {
			return total, fmt.Errorf("malware-feed/%s store: %w", f.Name(), err)
		}
		total += n
	}

	return total, nil
}

func (a *Aggregator) seedFromEmbedded() error {
	n, err := a.store.CountThreats()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	var entries []Entry
	if err := json.Unmarshal(seedData, &entries); err != nil {
		return fmt.Errorf("seed unmarshal: %w", err)
	}

	dbEntries := make([]db.ThreatEntry, len(entries))
	for i, e := range entries {
		dbEntries[i] = db.ThreatEntry{
			Package:   e.Package,
			Version:   e.Version,
			Ecosystem: e.Ecosystem,
			Feed:      e.Feed,
			Severity:  e.Severity,
			Reason:    e.Reason,
			Detected:  time.Now().UTC().Truncate(time.Second),
		}
	}

	inserted, err := a.store.UpsertThreats(dbEntries)
	if err != nil {
		return fmt.Errorf("seed insert: %w", err)
	}
	log.Printf("feed seed: inserted %d/%d compromised package artifacts", inserted, len(dbEntries))
	return nil
}

func (a *Aggregator) seedMalwareFromEmbedded() error {
	n, err := a.store.CountMalwareFamilies()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	var families []MalwareFamilyEntry
	if err := json.Unmarshal(seedMalwareData, &families); err != nil {
		return fmt.Errorf("seed malware unmarshal: %w", err)
	}

	dbFamilies := make([]db.MalwareFamily, len(families))
	for i, f := range families {
		dbFamilies[i] = db.MalwareFamily{
			Name:        f.Name,
			Type:        f.Type,
			Platform:    f.Platform,
			Discovered:  f.Discovered,
			Attribution: f.Attribution,
			Description: f.Description,
			CVE:         f.CVE,
			SourceURL:   f.SourceURL,
		}
	}

	inserted, err := a.store.UpsertMalwareFamilies(dbFamilies)
	if err != nil {
		return fmt.Errorf("seed malware insert: %w", err)
	}
	log.Printf("feed malware seed: inserted %d/%d malware families", inserted, len(dbFamilies))
	return nil
}
