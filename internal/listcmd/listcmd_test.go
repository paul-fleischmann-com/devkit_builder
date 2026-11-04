package listcmd

import (
	"bytes"
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
	write(t, root, "kits/alpha.json", `{"description":"Alpha kit","features":["git"]}`)
	write(t, root, "kits/beta.json", `{"features":["git"]}`)
	write(t, root, "features/git/feature.json", `{"id":"git","description":"Git client"}`)
	write(t, root, "features/git/install.sh", "#!/bin/sh\n")
	return root
}

func run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "factory.json")
	if err := os.WriteFile(cfg, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEVKIT_CONFIG", cfg)
	var out, errb bytes.Buffer
	code := Run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestListAll(t *testing.T) {
	code, out, errs := run(t, "--src", repo(t))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, want := range []string{"Kits:", "  alpha", "Alpha kit", "  beta", "Features:", "  git", "Git client"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestListKitsOnly(t *testing.T) {
	code, out, _ := run(t, "kits", "--src", repo(t))
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out, "Kits:") || strings.Contains(out, "git") {
		t.Errorf("unexpected output:\n%s", out)
	}
	if !strings.HasPrefix(out, "alpha") {
		t.Errorf("want unindented kit lines:\n%s", out)
	}
}

func TestListFeaturesOnly(t *testing.T) {
	code, out, _ := run(t, "features", "--src", repo(t))
	if code != 0 || !strings.Contains(out, "git") || strings.Contains(out, "alpha") {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
}

func TestListBadArgs(t *testing.T) {
	src := repo(t)
	if code, _, errs := run(t, "bogus", "--src", src); code != 2 || !strings.Contains(errs, "unknown list") {
		t.Errorf("unknown list: exit %d, %q", code, errs)
	}
	if code, _, errs := run(t, "kits", "features", "--src", src); code != 2 || !strings.Contains(errs, "usage") {
		t.Errorf("too many args: exit %d, %q", code, errs)
	}
}

func TestDescribeFallback(t *testing.T) {
	if got := describe("x", filepath.Join(t.TempDir(), "missing.json")); got != "x" {
		t.Errorf("got %q", got)
	}
}
