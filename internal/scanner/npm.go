package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/firfircelik/libkill/internal/db"
)

type NPMSscanner struct {
	store *db.Store
	home  string
}

func NewNPMScanner(store *db.Store) *NPMSscanner {
	home, _ := os.UserHomeDir()
	return &NPMSscanner{store: store, home: home}
}

func (n *NPMSscanner) Name() string    { return "npm" }
func (n *NPMSscanner) Ecosystem() string { return "npm" }

func (n *NPMSscanner) Scan(ctx context.Context) ([]Result, int, error) {
	var all []Result

	npmGlobal, _ := n.scanNPMGlobal(ctx)
	all = append(all, npmGlobal...)

	bunGlobal, _ := n.scanBunGlobal(ctx)
	all = append(all, bunGlobal...)

	matches := n.matchThreats(ctx, all)
	return matches, len(all), nil
}

type npmList struct {
	Dependencies map[string]npmDep `json:"dependencies"`
}

type npmDep struct {
	Version string `json:"version"`
}

func (n *NPMSscanner) scanNPMGlobal(ctx context.Context) ([]Result, error) {
	if _, err := exec.LookPath("npm"); err != nil {
		return nil, nil
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, "npm", "ls", "-g", "--json", "--depth=0")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	_ = cmd.Run()

	out := stdout.Bytes()
	if len(out) == 0 {
		return nil, nil
	}

	var list npmList
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, nil
	}

	var results []Result
	for name, dep := range list.Dependencies {
		results = append(results, Result{
			Package:   name,
			Version:   dep.Version,
			Ecosystem: "npm",
			Location:  "npm global",
		})
	}
	return results, nil
}

func (n *NPMSscanner) scanBunGlobal(ctx context.Context) ([]Result, error) {
	bunCache := filepath.Join(n.home, ".bun", "install", "cache")
	if _, err := os.Stat(bunCache); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(bunCache)
	if err != nil {
		return nil, nil
	}

	var results []Result
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		pkgFile := filepath.Join(bunCache, name, "package.json")
		data, err := os.ReadFile(pkgFile)
		if err != nil {
			continue
		}

		var pkg struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			continue
		}
		if pkg.Name == "" || pkg.Version == "" {
			continue
		}

		results = append(results, Result{
			Package:   pkg.Name,
			Version:   pkg.Version,
			Ecosystem: "npm",
			Location:  fmt.Sprintf("bun cache: %s", name),
		})
	}
	return results, nil
}

func (n *NPMSscanner) scanNodeModules(ctx context.Context) ([]Result, error) {
	var results []Result
	seen := map[string]bool{}

	searchDirs := []string{
		n.home,
	}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		pkgLockFiles := findLockFiles(dir, 4, "package.json", "bun.lockb", "pnpm-lock.yaml", "yarn.lock")
		for _, lockPath := range pkgLockFiles {
			projectDir := filepath.Dir(lockPath)
			nodeModulesDir := filepath.Join(projectDir, "node_modules")
			if _, err := os.Stat(nodeModulesDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(nodeModulesDir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				if strings.HasPrefix(name, "@") {
					subEntries, _ := os.ReadDir(filepath.Join(nodeModulesDir, name))
					for _, se := range subEntries {
						fullName := name + "/" + se.Name()
						if seen[fullName] {
							continue
						}
						seen[fullName] = true
						r := n.readPackageJSON(filepath.Join(nodeModulesDir, name, se.Name(), "package.json"), fullName, projectDir)
						if r != nil {
							results = append(results, *r)
						}
					}
				} else {
					if seen[name] {
						continue
					}
					seen[name] = true
					r := n.readPackageJSON(filepath.Join(nodeModulesDir, name, "package.json"), name, projectDir)
					if r != nil {
						results = append(results, *r)
					}
				}
			}
		}
	}

	return results, nil
}

func findLockFiles(root string, maxDepth int, names ...string) []string {
	var results []string
	depth := strings.Count(root, string(filepath.Separator)) + maxDepth

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
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
			return nil
		}
		for _, name := range names {
			if d.Name() == name {
				results = append(results, path)
				break
			}
		}
		return nil
	})

	return results
}

func (n *NPMSscanner) readPackageJSON(pkgPath, name, projectDir string) *Result {
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	if pkg.Version == "" {
		return nil
	}
	return &Result{
		Package:   name,
		Version:   pkg.Version,
		Ecosystem: "npm",
		Location:  fmt.Sprintf("node_modules in %s", projectDir),
	}
}

func (n *NPMSscanner) matchThreats(ctx context.Context, results []Result) []Result {
	var matched []Result
	for _, r := range results {
		threats, err := n.store.Search(r.Package, r.Ecosystem)
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
