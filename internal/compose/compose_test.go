package compose

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

func fixture(t *testing.T, base string) string {
	root := t.TempDir()
	write(t, root, "base.json", `{"image":"`+base+`"}`)
	write(t, root, "kits/k.json", `{"description":"d","apt_snapshot":"20261019T000000Z","features":["git"]}`)
	write(t, root, "features/git/feature.json", `{"id":"git"}`)
	write(t, root, "features/git/install.sh", "#!/bin/sh\n")
	return root
}

func TestCompose(t *testing.T) {
	root := fixture(t, "debian@sha256:"+strings.Repeat("a", 64))
	out, err := Compose(root, "k", "abc")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"FROM debian@sha256:", "COPY features/git/ /tmp/devkit/git/", `devkit.commit="abc"`, "USER 1000:1000"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestComposeRejectsTagBase(t *testing.T) {
	root := fixture(t, "debian:trixie")
	if _, err := Compose(root, "k", ""); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("want digest error, got %v", err)
	}
}
