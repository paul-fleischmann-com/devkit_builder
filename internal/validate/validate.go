// Package validate checks a devkits repository: base image, every kit and every feature.
package validate

import (
	"fmt"
	"io"

	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/internal/compose"
	"github.com/paul-fleischmann-com/devkit_builder/internal/config"
)

// Check returns one message per problem found below root.
func Check(root string) []string {
	var problems []string
	kits, err := compose.Kits(root)
	if err != nil {
		return []string{err.Error()}
	}
	if len(kits) == 0 {
		problems = append(problems, "no kits found in kits/")
	}
	for _, kit := range kits {
		if _, err := compose.Compose(root, kit, ""); err != nil {
			problems = append(problems, fmt.Sprintf("kit %s: %v", kit, err))
		}
	}
	features, err := compose.Features(root)
	if err != nil {
		return append(problems, err.Error())
	}
	for _, id := range features {
		if _, err := compose.CheckFeature(root, id); err != nil {
			problems = append(problems, err.Error())
		}
	}
	return problems
}

// Run implements `devkit validate`.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := cli.NewFlagSet("validate", stderr)
	root := cli.SrcFlag(fs)
	if _, err := cli.Parse(fs, args); err != nil {
		return 2
	}
	cfg, err := config.Load(*root, nil)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	problems := Check(cfg.Src)
	for _, p := range problems {
		fmt.Fprintln(stderr, "error:", p)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Fprintln(stdout, "ok")
	return 0
}
