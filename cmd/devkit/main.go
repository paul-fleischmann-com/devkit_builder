// Command devkit composes and builds devkit images: a base image plus features, in the order of a kit file.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/paul-fleischmann-com/devkit_builder/internal/build"
	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/internal/compose"
	"github.com/paul-fleischmann-com/devkit_builder/internal/config"
	"github.com/paul-fleischmann-com/devkit_builder/internal/doctor"
	"github.com/paul-fleischmann-com/devkit_builder/internal/images"
	"github.com/paul-fleischmann-com/devkit_builder/internal/initcmd"
	"github.com/paul-fleischmann-com/devkit_builder/internal/listcmd"
	"github.com/paul-fleischmann-com/devkit_builder/internal/validate"
)

// version is set by goreleaser (-X main.version=...).
var version = "v0.3.0"

const usage = `devkit <command> [flags]

Commands:
  init [dir]         create a devkits repository from the reference devkit
  list [kits|features]
                     list kits and features
  validate           check base image, kits and features
  compose <kit>      print the Containerfile of a kit (no git checks)
  build <kit>        build the image from the commit of the checkout (as the factory worker); --print only shows the Containerfile
  images             list the built devkit images with their FACTORY_TEST_PROFILES entry
  doctor             check this machine and the repository
  version            print the version

All commands except init and version take --src DIR, the checkout of the devkits repository
(default: $DEVKITS_DIR, "src" of factory.json, ./ if it has kits/, else ~/repos/devkits).

Settings come from factory.json in the current directory (or $DEVKIT_CONFIG); build, images and doctor
ask for them and create it if it is missing.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 1
	}
	rest := args[1:]
	switch args[0] {
	case "version", "--version":
		fmt.Fprintln(stdout, version)
		return 0
	case "init":
		return initcmd.Run(rest, stdout, stderr)
	case "list":
		return listcmd.Run(rest, stdout, stderr)
	case "validate":
		return validate.Run(rest, stdout, stderr)
	case "compose":
		return composeCmd(rest, stdout, stderr)
	case "build":
		return build.Run(rest, stdout, stderr)
	case "images":
		return images.Run(rest, stdout, stderr)
	case "doctor":
		return doctor.Run(rest, stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	}
	fmt.Fprintf(stderr, "devkit: unknown command %q\n\n%s", args[0], usage)
	return 2
}

func composeCmd(args []string, stdout, stderr io.Writer) int {
	fs := cli.NewFlagSet("compose", stderr)
	root := cli.SrcFlag(fs)
	commit := fs.String("commit", "", "commit of the devkits repository, recorded as a label")
	pos, err := cli.Parse(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: devkit compose <kit> [--src DIR] [--commit SHA]")
		return 2
	}
	cfg, err := config.Load(*root, nil)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	out, err := compose.Compose(cfg.Src, pos[0], *commit)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	fmt.Fprint(stdout, out)
	return 0
}
