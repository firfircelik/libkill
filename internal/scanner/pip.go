package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/firfircelik/libkill/internal/db"
)

type PIPScanner struct {
	store *db.Store
	home  string
}

func NewPIPScanner(store *db.Store) *PIPScanner {
	home, _ := os.UserHomeDir()
	return &PIPScanner{store: store, home: home}
}

func (p *PIPScanner) Name() string    { return "pip" }
func (p *PIPScanner) Ecosystem() string { return "pip" }

type pipPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (p *PIPScanner) Scan(ctx context.Context) ([]Result, int, error) {
	var all []Result

	global, _ := p.scanPipGlobal(ctx)
	all = append(all, global...)

	matches := p.matchThreats(ctx, all)
	return matches, len(all), nil
}

func (p *PIPScanner) scanPipGlobal(ctx context.Context) ([]Result, error) {
	for _, cmd := range []string{"pip3", "pip"} {
		if _, err := exec.LookPath(cmd); err != nil {
			continue
		}

		c := exec.CommandContext(ctx, cmd, "list", "--format", "json")
		c.Env = append(os.Environ(), "PIP_NO_COLOR=1")
		out, err := c.Output()
		if err != nil {
			continue
		}

		var pkgs []pipPackage
		if err := json.Unmarshal(out, &pkgs); err != nil {
			continue
		}

		var results []Result
		for _, pkg := range pkgs {
			results = append(results, Result{
				Package:   strings.ToLower(pkg.Name),
				Version:   pkg.Version,
				Ecosystem: "pip",
				Location:  fmt.Sprintf("pip global (%s)", cmd),
			})
		}
		return results, nil
	}
	return nil, nil
}

func (p *PIPScanner) scanVirtualEnvs(ctx context.Context) ([]Result, error) {
	var results []Result

	venvDirs := findVenvDirs(p.home, 5)
	for _, venvPath := range venvDirs {
		pipPath := filepath.Join(venvPath, "bin", "pip")
		if _, err := os.Stat(pipPath); err != nil {
			pipPath = filepath.Join(venvPath, "bin", "pip3")
		}
		if _, err := os.Stat(pipPath); os.IsNotExist(err) {
			continue
		}

		c := exec.CommandContext(ctx, pipPath, "list", "--format", "json")
		c.Env = append(os.Environ(), "PIP_NO_COLOR=1")
		out, err := c.Output()
		if err != nil {
			continue
		}

		var pkgs []pipPackage
		if err := json.Unmarshal(out, &pkgs); err != nil {
			continue
		}

		for _, pkg := range pkgs {
			results = append(results, Result{
				Package:   strings.ToLower(pkg.Name),
				Version:   pkg.Version,
				Ecosystem: "pip",
				Location:  fmt.Sprintf("venv: %s", venvPath),
			})
		}
	}

	return results, nil
}

func findVenvDirs(root string, maxDepth int) []string {
	var results []string
	depth := strings.Count(root, string(filepath.Separator)) + maxDepth

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		base := d.Name()
		if strings.HasPrefix(base, ".") && base != "." && base != ".." {
			return filepath.SkipDir
		}
		if base == "node_modules" || base == "vendor" || base == "__pycache__" {
			return filepath.SkipDir
		}
		sepCount := strings.Count(path, string(filepath.Separator))
		if sepCount > depth {
			return filepath.SkipDir
		}

		cfgPath := filepath.Join(path, "pyvenv.cfg")
		if _, err := os.Stat(cfgPath); err == nil {
			results = append(results, path)
			return filepath.SkipDir
		}
		return nil
	})

	return results
}

func (p *PIPScanner) matchThreats(ctx context.Context, results []Result) []Result {
	var matched []Result
	for _, r := range results {
		threats, err := p.store.Search(r.Package, r.Ecosystem)
		if err != nil || len(threats) == 0 {
			continue
		}
		for _, t := range threats {
			if t.Version == "" || t.Version == r.Version {
				r.Threat = t
				matched = append(matched, r)
				break
			}
		}
	}
	return matched
}
