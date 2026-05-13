package daemon

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/firfircelik/libkill/internal/db"
	"github.com/firfircelik/libkill/internal/feed"
	"github.com/firfircelik/libkill/internal/notify"
	"github.com/firfircelik/libkill/internal/scanner"
)

type Daemon struct {
	store    *db.Store
	feed     *feed.Aggregator
	notifier *notify.Notifier
	interval time.Duration
}

func New(store *db.Store, f *feed.Aggregator, interval time.Duration) *Daemon {
	return &Daemon{
		store:    store,
		feed:     f,
		notifier: notify.New(),
		interval: interval,
	}
}

func (d *Daemon) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("LibKill daemon started (interval: %s)", d.interval)

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	if err := d.scanAndNotify(ctx); err != nil {
		log.Printf("initial scan: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("LibKill daemon shutting down")
			return nil
		case <-ticker.C:
			if err := d.scanAndNotify(ctx); err != nil {
				log.Printf("scan error: %v", err)
			}
		}
	}
}

func (d *Daemon) scanAndNotify(ctx context.Context) error {
	log.Println("Starting periodic scan...")

	if _, err := d.feed.Update(ctx); err != nil {
		log.Printf("feed update: %v", err)
	}

	scanners := []scanner.Scanner{
		scanner.NewNPMScanner(d.store),
		scanner.NewPIPScanner(d.store),
	}

	var threats []scanner.Result
	for _, s := range scanners {
		matches, _, err := s.Scan(ctx)
		if err != nil {
			log.Printf("%s scan error: %v", s.Name(), err)
			continue
		}
		for _, r := range matches {
			if r.Threat.Package != "" {
				threats = append(threats, r)
			}
		}
	}

	if len(threats) > 0 {
		msg := fmt.Sprintf("%d compromised package(s) found", len(threats))
		for _, t := range threats {
			msg += fmt.Sprintf("\n  %s@%s (%s)", t.Package, t.Version, t.Location)
		}
		log.Printf("THREATS FOUND: %s", msg)

		if err := d.notifier.Send(
			"LibKill Alert",
			fmt.Sprintf("%d compromised package(s) detected on your system. Run 'libkill tui' to review.", len(threats)),
		); err != nil {
			log.Printf("notification error: %v", err)
		}
	} else {
		log.Println("Scan complete: no threats found")
	}

	return nil
}
