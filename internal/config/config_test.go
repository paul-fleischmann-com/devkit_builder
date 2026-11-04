package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolate runs the test in an empty directory with no DEVKIT* variables.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
	for _, v := range []string{"DEVKIT_CONFIG", "DEVKITS_DIR", "DEVKIT_WORKER", "DEVKIT_TMPDIR", "DEVKIT_MIN_FREE_MB", "DEVKIT_IMAGE_PREFIX"} {
		t.Setenv(v, "")
	}
	t.Setenv("HOME", dir)
	return dir
}

func TestDefaults(t *testing.T) {
	isolate(t)
	c, err := Load("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Path != "" || c.Worker != "factory-worker" || c.MinFreeMB != 2048 || c.ImagePrefix != "localhost/factory-devkit-" || c.Tmpdir != "/server/factory-containers/tmp" {
		t.Errorf("unexpected defaults: %+v", c)
	}
}

func TestFileAndPrecedence(t *testing.T) {
	dir := isolate(t)
	body := `{"src":"~/kits","worker":{"name":"w"},"podman":{"tmpdir":"tmp","min_free_mb":5},"images":{"prefix":"p-"}}`
	if err := os.WriteFile("factory.json", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Worker != "w" || c.MinFreeMB != 5 || c.ImagePrefix != "p-" {
		t.Errorf("file values not used: %+v", c)
	}
	if c.Src != filepath.Join(dir, "kits") {
		t.Errorf("~ not expanded: %s", c.Src)
	}
	if c.Tmpdir != filepath.Join(dir, "tmp") {
		t.Errorf("relative tmpdir must be relative to the file: %s", c.Tmpdir)
	}

	t.Setenv("DEVKIT_WORKER", "env-worker")
	t.Setenv("DEVKIT_MIN_FREE_MB", "7")
	t.Setenv("DEVKITS_DIR", "/from/env")
	c, _ = Load("", nil)
	if c.Worker != "env-worker" || c.MinFreeMB != 7 || c.Src != "/from/env" {
		t.Errorf("environment must beat the file: %+v", c)
	}
	c, _ = Load("/from/flag", nil)
	if c.Src != "/from/flag" {
		t.Errorf("flag must beat the environment: %s", c.Src)
	}
}

func TestBadInput(t *testing.T) {
	isolate(t)
	os.WriteFile("factory.json", []byte("{"), 0o644)
	if _, err := Load("", nil); err == nil || !strings.Contains(err.Error(), "JSON") {
		t.Errorf("want JSON error, got %v", err)
	}
	os.Remove("factory.json")
	t.Setenv("DEVKIT_MIN_FREE_MB", "many")
	if _, err := Load("", nil); err == nil {
		t.Error("want error for non-numeric DEVKIT_MIN_FREE_MB")
	}
}

func TestPromptCreatesFile(t *testing.T) {
	dir := isolate(t)
	var out strings.Builder
	in := "~/mykits\n\nrelative\n/var/tmp/x\nabc\n1024\n\n"
	p := &Prompt{In: newReader(in), Out: &out}
	c, err := Load("", p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Src != filepath.Join(dir, "mykits") || c.Tmpdir != "/var/tmp/x" || c.MinFreeMB != 1024 || c.Worker != "factory-worker" {
		t.Errorf("unexpected config: %+v", c)
	}
	if !strings.Contains(out.String(), "absolute") || !strings.Contains(out.String(), "number") {
		t.Errorf("invalid answers were not rejected:\n%s", out.String())
	}
	// the file exists now, so the next load needs no prompt
	c2, err := Load("", nil)
	if err != nil || c2.Path == "" || c2.Tmpdir != "/var/tmp/x" {
		t.Errorf("created file not used: %+v, %v", c2, err)
	}
}

func TestPromptAbortWritesNothing(t *testing.T) {
	isolate(t)
	if _, err := Load("", &Prompt{In: newReader("only-src\n"), Out: &strings.Builder{}}); err == nil {
		t.Fatal("want error at end of input")
	}
	if _, err := os.Stat("factory.json"); err == nil {
		t.Error("factory.json must not be written after an aborted prompt")
	}
}
