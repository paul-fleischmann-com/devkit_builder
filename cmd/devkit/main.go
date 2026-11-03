package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paul-fleischmann-com/devkit_builder/internal/compose"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "compose" {
		fmt.Fprintln(os.Stderr, "usage: devkit compose <kit> [--commit SHA]")
		os.Exit(1)
	}
	fs := flag.NewFlagSet("compose", flag.ExitOnError)
	commit := fs.String("commit", "", "commit label")
	_ = fs.Parse(os.Args[3:])
	kit := os.Args[2]
	root, _ := filepath.Abs(".")
	out, err := compose.Compose(root, kit, *commit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	fmt.Print(out)
}
