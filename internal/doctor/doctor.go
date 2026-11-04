// Package doctor implements `devkit doctor`: checks whether this machine can build devkits.
package doctor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/paul-fleischmann-com/devkit_builder/internal/build"
	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/internal/config"
	"github.com/paul-fleischmann-com/devkit_builder/internal/podman"
	"github.com/paul-fleischmann-com/devkit_builder/internal/validate"
)

// Run implements `devkit doctor`. It exits 1 if a check fails.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := cli.NewFlagSet("doctor", stderr)
	src := cli.SrcFlag(fs)
	if _, err := cli.Parse(fs, args); err != nil {
		return 2
	}
	failed := 0
	report := func(ok bool, format string, a ...any) {
		mark := "ok  "
		if !ok {
			mark = "FAIL"
			failed++
		}
		fmt.Fprintf(stdout, "[%s] %s\n", mark, fmt.Sprintf(format, a...))
	}
	ctx := context.Background()

	cfg, err := config.Load(*src, config.StdPrompt(stderr))
	if err != nil {
		report(false, "%v", err)
		return 1
	}
	if cfg.Path == "" {
		fmt.Fprintln(stdout, "[ -- ] no factory.json found, using defaults")
	} else {
		report(true, "config %s", cfg.Path)
	}
	fmt.Fprintf(stdout, "[ -- ] worker=%s image_prefix=%s tmpdir=%s\n", cfg.Worker, cfg.ImagePrefix, cfg.Tmpdir)

	_, err = exec.LookPath("git")
	report(err == nil, "git in PATH")
	st, err := os.Stat(filepath.Join(cfg.Src, ".git"))
	report(err == nil && st.IsDir(), "%s is a git checkout", cfg.Src)

	w := podman.Worker{Name: cfg.Worker, TmpDir: cfg.Tmpdir}
	if _, err := w.Output(ctx, "podman", "--version"); err != nil {
		report(false, "podman as %s: %v", cfg.Worker, err)
	} else {
		report(true, "podman as %s (sudo -n)", cfg.Worker)
		if _, err := w.Output(ctx, "test", "-d", cfg.Tmpdir, "-a", "-w", cfg.Tmpdir); err != nil {
			report(false, "tmpdir %s does not exist or is not writable for %s", cfg.Tmpdir, cfg.Worker)
		} else {
			free, err := build.FreeMB(ctx, w, cfg.Tmpdir)
			report(err == nil && free >= cfg.MinFreeMB, "tmpdir %s: %d MB free (need %d)", cfg.Tmpdir, free, cfg.MinFreeMB)
		}
	}

	problems := validate.Check(cfg.Src)
	report(len(problems) == 0, "devkits repository %s: %d problem(s)", cfg.Src, len(problems))
	for _, p := range problems {
		fmt.Fprintln(stdout, "       -", p)
	}

	if failed > 0 {
		return 1
	}
	return 0
}
