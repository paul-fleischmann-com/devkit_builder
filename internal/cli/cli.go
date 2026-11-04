// Package cli holds the small helpers shared by all subcommands.
package cli

import (
	"flag"
	"io"
)

// NewFlagSet returns a FlagSet that reports errors to stderr and never exits.
func NewFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("devkit "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

// Parse parses flags and positional arguments in any order
// (`compose kit --commit X` as well as `compose --commit X kit`).
func Parse(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return positional, nil
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

// SrcFlag registers --src (and its alias --src): the checkout of the devkits repository.
func SrcFlag(fs *flag.FlagSet) *string {
	src := new(string)
	fs.StringVar(src, "src", "", "checkout of the devkits repository (default: $DEVKITS_DIR, \"src\" of factory.json, ./ or ~/repos/devkits)")
	fs.StringVar(src, "root", "", "alias of --src")
	return src
}
