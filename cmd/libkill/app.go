package main

import (
	"context"
	"fmt"
	"os"

	"github.com/firfircelik/libkill/internal/config"
	"github.com/firfircelik/libkill/internal/db"
	"github.com/firfircelik/libkill/internal/feed"
)

type app struct {
	cfg   *config.Config
	store *db.Store
	feed  *feed.Aggregator
}

func newApp(cfg *config.Config) (*app, error) {
	store, err := db.New(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	f := feed.NewAggregator(store)

	return &app{
		cfg:   cfg,
		store: store,
		feed:  f,
	}, nil
}

func (a *app) close() {
	a.store.Close()
}

func (a *app) forceUpdate() error {
	fmt.Fprintf(os.Stderr, "Updating threat database...\n")
	n, err := a.feed.Update(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "  %d new threats added\n", n)
	return nil
}
