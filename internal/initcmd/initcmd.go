// Package initcmd implements `devkit init`: a new devkits repository from the embedded reference devkit.
package initcmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/templates"
)

const templateRoot = "reference-devkit"

// Run implements `devkit init [dir] [--force]`.
func Run(args []string, stdout, stderr io.Writer) int {
	flags := cli.NewFlagSet("init", stderr)
	force := flags.Bool("force", false, "overwrite files that already exist")
	pos, err := cli.Parse(flags, args)
	if err != nil {
		return 2
	}
	if len(pos) > 1 {
		fmt.Fprintln(stderr, "usage: devkit init [dir] [--force]")
		return 2
	}
	target := "."
	if len(pos) == 1 {
		target = pos[0]
	}
	if err := Scaffold(target, *force); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	fmt.Fprintf(stdout, "devkit repository created in %s\n", target)
	fmt.Fprintln(stdout, "next: put the digest of your base image into base.json, then run `devkit validate`")
	return 0
}

// Scaffold copies the reference devkit into target. Without force it refuses to overwrite files.
func Scaffold(target string, force bool) error {
	var files []string
	err := fs.WalkDir(templates.FS, templateRoot, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			files = append(files, strings.TrimPrefix(p, templateRoot+"/"))
		}
		return err
	})
	if err != nil {
		return err
	}
	if !force {
		for _, rel := range files {
			if _, err := os.Lstat(filepath.Join(target, filepath.FromSlash(rel))); err == nil {
				return fmt.Errorf("%s exists; use --force to overwrite", filepath.Join(target, filepath.FromSlash(rel)))
			}
		}
	}
	for _, rel := range files {
		data, err := fs.ReadFile(templates.FS, path.Join(templateRoot, rel))
		if err != nil {
			return err
		}
		dst := filepath.Join(target, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(rel, ".sh") {
			mode = 0o755
		}
		if err := os.WriteFile(dst, data, mode); err != nil {
			return err
		}
	}
	return nil
}
