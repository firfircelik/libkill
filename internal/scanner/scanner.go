package scanner

import (
	"context"

	"github.com/firfircelik/libkill/internal/db"
)

type Result struct {
	Package   string
	Version   string
	Ecosystem string
	Location  string
	Threat    db.ThreatEntry
}

type Scanner interface {
	Name() string
	Ecosystem() string
	Scan(ctx context.Context) (matches []Result, total int, err error)
}
