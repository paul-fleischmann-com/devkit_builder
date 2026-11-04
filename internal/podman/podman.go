// Package podman runs commands as the factory worker, the way build-devkit.sh does:
// `sudo -n -u <worker> -H env XDG_RUNTIME_DIR=... DBUS_SESSION_BUS_ADDRESS=... [TMPDIR=...] cmd`.
package podman

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"os/user"
)

// Worker is the user that owns the rootless podman of the factory.
type Worker struct {
	Name   string
	TmpDir string // TMPDIR for the commands; empty: not set
}

// Command prepares args to run as the worker (in /). The worker is looked up here, not earlier,
// so that commands that do not need it work on a machine without it.
func (w Worker) Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	u, err := user.Lookup(w.Name)
	if err != nil {
		return nil, fmt.Errorf("the user %s does not exist", w.Name)
	}
	argv := []string{"-n", "-u", w.Name, "-H", "env",
		"XDG_RUNTIME_DIR=/run/user/" + u.Uid,
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/" + u.Uid + "/bus"}
	if w.TmpDir != "" {
		argv = append(argv, "TMPDIR="+w.TmpDir)
	}
	cmd := exec.CommandContext(ctx, "sudo", append(argv, args...)...)
	cmd.Dir = "/"
	return cmd, nil
}

// Output runs args as the worker and returns stdout; stderr goes into the error.
func (w Worker) Output(ctx context.Context, args ...string) (string, error) {
	cmd, err := w.Command(ctx, args...)
	if err != nil {
		return "", err
	}
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	if err := cmd.Run(); err != nil {
		if msg := bytes.TrimSpace(errBuf.Bytes()); len(msg) > 0 {
			return out.String(), fmt.Errorf("%s: %w: %s", args[0], err, msg)
		}
		return out.String(), fmt.Errorf("%s: %w", args[0], err)
	}
	return out.String(), nil
}
