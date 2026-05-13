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
	store  *db.Store
	feeds  []Fetcher
}

func NewAggregator(store *db.Store) *Aggregator {
	a := &Aggregator{
		store: store,
		feeds: []Fetcher{
			NewSocketFeed(),
			NewOSVFeed(),
			NewGitHubFeed(),
		},
	}

	if err := a.seedFromEmbedded(); err != nil {
		log.Printf("feed seed: %v", err)
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
