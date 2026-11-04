// Package config reads the factory configuration (factory.json), applies the DEVKIT_* overrides and,
// if there is no file yet, asks the user for the settings and creates it.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config is the effective configuration after environment, file and defaults. Paths are expanded.
type Config struct {
	Path        string // file that was read or created; empty if defaults are used
	Src         string // checkout of the devkits repository
	Worker      string
	Tmpdir      string
	MinFreeMB   int
	ImagePrefix string
}

// file is factory.json. Paths are kept as written (`~`, `$VAR` and relative paths are expanded on load).
type file struct {
	Src    string `json:"src,omitempty"`
	Worker struct {
		Name string `json:"name,omitempty"`
	} `json:"worker"`
	Podman struct {
		Tmpdir    string `json:"tmpdir,omitempty"`
		MinFreeMB *int   `json:"min_free_mb,omitempty"`
	} `json:"podman"`
	Images struct {
		Prefix string `json:"prefix,omitempty"`
	} `json:"images"`
}

const (
	defaultWorker    = "factory-worker"
	defaultTmpdir    = "/server/factory-containers/tmp"
	defaultMinFreeMB = 2048
	defaultPrefix    = "localhost/factory-devkit-"
	fileName         = "factory.json"
)

// Locate returns the config file: $DEVKIT_CONFIG, else factory.json or config/factory.json in the current
// directory, else in srcHint (the checkout named by --src or $DEVKITS_DIR). It returns "" if there is none.
func Locate(srcHint string) string {
	if p := os.Getenv("DEVKIT_CONFIG"); p != "" {
		return p
	}
	dirs := []string{"."}
	if srcHint != "" {
		dirs = append(dirs, expand(srcHint, "."))
	}
	for _, dir := range dirs {
		for _, p := range []string{filepath.Join(dir, fileName), filepath.Join(dir, "config", fileName)} {
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}
	return ""
}

// Load builds the configuration. srcFlag is the value of --src ("" if not given).
// Precedence per value: flag/environment, file, default.
// If there is no config file and prompt is not nil, the user is asked for the settings and ./factory.json is created;
// otherwise defaults are used.
func Load(srcFlag string, prompt *Prompt) (Config, error) {
	hint := firstNonEmpty(srcFlag, os.Getenv("DEVKITS_DIR"))
	path := Locate(hint)

	var f file
	switch {
	case path != "":
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("config %s: %w", path, err)
		}
		if err := json.Unmarshal(data, &f); err != nil {
			return Config{}, fmt.Errorf("config %s: not readable as JSON (%w)", path, err)
		}
	case prompt != nil:
		var err error
		if f, path, err = prompt.create(hint); err != nil {
			return Config{}, err
		}
	}
	base := "."
	if path != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return Config{}, err
		}
		path, base = abs, filepath.Dir(abs)
	}

	c := Config{
		Path:        path,
		Worker:      firstNonEmpty(os.Getenv("DEVKIT_WORKER"), f.Worker.Name, defaultWorker),
		ImagePrefix: firstNonEmpty(os.Getenv("DEVKIT_IMAGE_PREFIX"), f.Images.Prefix, defaultPrefix),
		MinFreeMB:   defaultMinFreeMB,
	}
	// flag and environment are relative to the current directory, the file's values to the file
	if v := firstNonEmpty(srcFlag, os.Getenv("DEVKITS_DIR")); v != "" {
		c.Src = expand(v, ".")
	} else if f.Src != "" {
		c.Src = expand(f.Src, base)
	} else {
		c.Src = defaultSrc()
	}
	if v := os.Getenv("DEVKIT_TMPDIR"); v != "" {
		c.Tmpdir = expand(v, ".")
	} else {
		c.Tmpdir = expand(firstNonEmpty(f.Podman.Tmpdir, defaultTmpdir), base)
	}
	if f.Podman.MinFreeMB != nil {
		c.MinFreeMB = *f.Podman.MinFreeMB
	}
	if v := os.Getenv("DEVKIT_MIN_FREE_MB"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("DEVKIT_MIN_FREE_MB: not a number: %s", v)
		}
		c.MinFreeMB = n
	}
	return c, nil
}

// defaultSrc is the current directory if it has kits/, else ~/repos/devkits.
func defaultSrc() string {
	if st, err := os.Stat("kits"); err == nil && st.IsDir() {
		return "."
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "repos", "devkits")
	}
	return "."
}

// expand resolves `~`, `$VAR` and relative paths (against base).
func expand(p, base string) string {
	p = os.ExpandEnv(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(base, p)
	}
	return p
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
