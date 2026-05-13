package main

import (
	"fmt"
	"os"

	"github.com/firfircelik/libkill/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "libkill: loading config: %v\n", err)
		os.Exit(1)
	}

	app, err := newApp(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "libkill: initializing: %v\n", err)
		os.Exit(1)
	}
	defer app.close()

	root := newRootCmd(app)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "libkill: %v\n", err)
		os.Exit(1)
	}
}
