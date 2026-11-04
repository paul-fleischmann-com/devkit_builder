package build

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paul-fleischmann-com/devkit_builder/internal/podman"
)

func TestFreeMB(t *testing.T) {
	w := podman.Worker{Exec: func(_ context.Context, args ...string) (string, error) {
		if args[0] != "df" {
			t.Errorf("unexpected command %v", args)
		}
		return "Avail\n  4096\n", nil
	}}
	n, err := FreeMB(context.Background(), w, "/tmp")
	if err != nil || n != 4096 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestFreeMBErrors(t *testing.T) {
	w := podman.Worker{Exec: func(context.Context, ...string) (string, error) { return "", errors.New("no df") }}
	if _, err := FreeMB(context.Background(), w, "/x"); err == nil {
		t.Error("want error from df")
	}
	w = podman.Worker{Exec: func(context.Context, ...string) (string, error) { return "Avail\nabc\n", nil }}
	if _, err := FreeMB(context.Background(), w, "/x"); err == nil {
		t.Error("want error for non-numeric output")
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func devkitsRepo(t *testing.T) string {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	files := map[string]string{
		"base.json":                 `{"image":"debian@sha256:` + strings.Repeat("b", 64) + `"}`,
		"kits/k.json":               `{"apt_snapshot":"20261019T000000Z","features":["git"]}`,
		"features/git/feature.json": `{"id":"git"}`,
		"features/git/install.sh":   "#!/bin/sh\n",
	}
	for rel, c := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-q", "-m", "init")
	return root
}

func runBuild(t *testing.T, args ...string) (int, string, string) {
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

func TestPrintContainerfile(t *testing.T) {
	src := devkitsRepo(t)
	code, out, errs := runBuild(t, "k", "--print", "--src", src)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	if !strings.Contains(out, "FROM debian@sha256:") {
		t.Errorf("no FROM line:\n%s", out)
	}
}

func TestBuildRefusals(t *testing.T) {
	src := devkitsRepo(t)

	if code, _, errs := runBuild(t, "--print", "--src", src); code != 2 || !strings.Contains(errs, "usage") {
		t.Errorf("no kit: exit %d, %q", code, errs)
	}
	if code, _, errs := runBuild(t, "nokit", "--print", "--src", src); code != 2 || !strings.Contains(errs, "no kit nokit") {
		t.Errorf("unknown kit: exit %d, %q", code, errs)
	}
	if code, _, errs := runBuild(t, "k", "--print", "--src", t.TempDir()); code != 2 || !strings.Contains(errs, "not a checkout") {
		t.Errorf("no checkout: exit %d, %q", code, errs)
	}
	if code, _, errs := runBuild(t, "k", "--print", "--commit", "deadbeef", "--src", src); code != 2 || !strings.Contains(errs, "not deadbeef") {
		t.Errorf("wrong commit: exit %d, %q", code, errs)
	}

	if err := os.WriteFile(filepath.Join(src, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := runBuild(t, "k", "--print", "--src", src); code != 2 || !strings.Contains(errs, "not committed") {
		t.Errorf("dirty checkout: exit %d, %q", code, errs)
	}
}
