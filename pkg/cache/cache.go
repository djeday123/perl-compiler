// pkg/cache/cache.go
package cache

import (
	"os"
	"path/filepath"
	"time"
)

var cacheDir string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	cacheDir = filepath.Join(home, ".perlc", "cache")
}

type CachedModule struct {
	Name       string
	Version    string
	GoCode     string
	CompiledAt time.Time
}

func Get(module, version string) (*CachedModule, bool) {
	path := filepath.Join(cacheDir, module, version+".go")
	if _, err := os.Stat(path); err == nil {
		content, _ := os.ReadFile(path)
		return &CachedModule{
			Name:    module,
			Version: version,
			GoCode:  string(content),
		}, true
	}
	return nil, false
}

func Store(module, version, goCode string) {
	dir := filepath.Join(cacheDir, module)
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, version+".go")
	os.WriteFile(path, []byte(goCode), 0644)
}
