package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func repo(t *testing.T) string {
	root := t.TempDir()
	write(t, root, "base.json", `{"image":"debian@sha256:`+strings.Repeat("b", 64)+`"}`)
	write(t, root, "kits/k.json", `{"apt_snapshot":"20261019T000000Z","features":["git"]}`)
	write(t, root, "features/git/feature.json", `{"id":"git"}`)
	write(t, root, "features/git/install.sh", "#!/bin/sh\n")
	return root
}

func TestCheckOK(t *testing.T) {
	if p := Check(repo(t)); len(p) != 0 {
		t.Fatalf("unexpected problems: %v", p)
	}
}

func TestCheckFindsProblems(t *testing.T) {
	root := repo(t)
	write(t, root, "features/broken/feature.json", `{"id":"broken"}`) // no install.sh, used by no kit
	write(t, root, "kits/bad.json", `{"apt_snapshot":"x","features":["git"]}`)
	p := strings.Join(Check(root), "\n")
	for _, want := range []string{"kit bad", "feature broken: no install.sh"} {
		if !strings.Contains(p, want) {
			t.Errorf("missing %q in\n%s", want, p)
		}
	}
}

func TestCheckNoKitsDir(t *testing.T) {
	if p := Check(t.TempDir()); len(p) == 0 {
		t.Fatal("want a problem for an empty directory")
	}
}
