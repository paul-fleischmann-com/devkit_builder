package podman

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestOutputUsesExec(t *testing.T) {
	var got []string
	w := Worker{Name: "nobody-such-user", Exec: func(_ context.Context, args ...string) (string, error) {
		got = args
		return "out", nil
	}}
	out, err := w.Output(context.Background(), "podman", "--version")
	if err != nil || out != "out" || strings.Join(got, " ") != "podman --version" {
		t.Fatalf("out=%q err=%v args=%v", out, err, got)
	}
}

func TestCommandUnknownWorker(t *testing.T) {
	_, err := Worker{Name: "no-such-user-xyz"}.Command(context.Background(), "true")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("err=%v", err)
	}
	if _, err := (Worker{Name: "no-such-user-xyz"}).Output(context.Background(), "true"); err == nil {
		t.Fatal("want error")
	}
}

func TestOutputExecError(t *testing.T) {
	want := errors.New("boom")
	w := Worker{Exec: func(context.Context, ...string) (string, error) { return "", want }}
	if _, err := w.Output(context.Background(), "x"); !errors.Is(err, want) {
		t.Fatalf("err=%v", err)
	}
}
