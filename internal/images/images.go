// Package images implements `devkit images`: the devkit images of the worker, with the ready line for FACTORY_TEST_PROFILES.
package images

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/internal/config"
	"github.com/paul-fleischmann-com/devkit_builder/internal/podman"
)

// Run implements `devkit images`: read only, newest first.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := cli.NewFlagSet("images", stderr)
	src := cli.SrcFlag(fs)
	if _, err := cli.Parse(fs, args); err != nil {
		return 2
	}
	cfg, err := config.Load(*src, config.StdPrompt(stderr))
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	w := podman.Worker{Name: cfg.Worker}
	listing, err := w.Output(context.Background(), "podman", "images", "--no-trunc",
		"--format", "{{.Repository}}|{{.Tag}}|{{.ID}}|{{.CreatedAt}}|{{.Size}}")
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	found := false
	for _, line := range strings.Split(listing, "\n") {
		f := strings.SplitN(line, "|", 5)
		if len(f) != 5 || !strings.HasPrefix(f[0], cfg.ImagePrefix) {
			continue
		}
		found = true
		kit := strings.TrimPrefix(f[0], cfg.ImagePrefix)
		fmt.Fprintf(stdout, "kit=%s commit=%s created=%s size=%s\n", kit, f[1], f[3], f[4])
		fmt.Fprintf(stdout, "  FACTORY_TEST_PROFILES entry: %s=%s\n", kit, strings.TrimPrefix(f[2], "sha256:"))
	}
	if !found {
		fmt.Fprintln(stdout, "no devkit image yet: build one with  devkit build <kit>")
	}
	return 0
}
