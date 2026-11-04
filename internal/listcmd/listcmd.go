// Package listcmd implements `devkit list`: the kits and features of a devkits repository.
package listcmd

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/internal/compose"
	"github.com/paul-fleischmann-com/devkit_builder/internal/config"
)

// Run implements `devkit list [kits|features]`.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := cli.NewFlagSet("list", stderr)
	root := cli.SrcFlag(fs)
	pos, err := cli.Parse(fs, args)
	if err != nil {
		return 2
	}
	cfg, err := config.Load(*root, nil)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	src := cfg.Src
	what := "all"
	if len(pos) > 1 {
		fmt.Fprintln(stderr, "usage: devkit list [kits|features] [--src DIR]")
		return 2
	}
	if len(pos) == 1 {
		what = pos[0]
	}
	if what != "all" && what != "kits" && what != "features" {
		fmt.Fprintf(stderr, "error: unknown list %q (kits or features)\n", what)
		return 2
	}

	if what != "features" {
		kits, err := compose.Kits(src)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		if what == "all" {
			fmt.Fprintln(stdout, "Kits:")
		}
		for _, kit := range kits {
			fmt.Fprintf(stdout, "%s%s\n", indent(what), describe(kit, filepath.Join(src, "kits", kit+".json")))
		}
	}
	if what != "kits" {
		features, err := compose.Features(src)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		if what == "all" {
			fmt.Fprintln(stdout, "Features:")
		}
		for _, id := range features {
			fmt.Fprintf(stdout, "%s%s\n", indent(what), describe(id, filepath.Join(src, "features", id, "feature.json")))
		}
	}
	return 0
}

func indent(what string) string {
	if what == "all" {
		return "  "
	}
	return ""
}

// describe returns "name  description" using the description of the JSON file, if readable.
func describe(name, path string) string {
	meta, err := loadMeta(path)
	if err != nil || meta == "" {
		return name
	}
	return fmt.Sprintf("%-20s %s", name, meta)
}
