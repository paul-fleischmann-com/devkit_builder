package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Prompt asks the user for the settings of a new factory.json.
type Prompt struct {
	In  *bufio.Reader
	Out io.Writer
}

// StdPrompt returns a prompt on stdin, writing to out (stderr, so that stdout stays clean for pipes).
// It returns nil if stdin is not a terminal: scripts and CI get the defaults instead of a hanging question.
func StdPrompt(out io.Writer) *Prompt {
	st, err := os.Stdin.Stat()
	if err != nil || st.Mode()&os.ModeCharDevice == 0 {
		return nil
	}
	if null, err := os.Stat(os.DevNull); err == nil && os.SameFile(st, null) {
		return nil
	}
	return &Prompt{In: bufio.NewReader(os.Stdin), Out: out}
}

// create asks for all settings and writes ./factory.json; it returns the settings and the path of the file.
func (p *Prompt) create(srcHint string) (file, string, error) {
	path, err := filepath.Abs(fileName)
	if err != nil {
		return file{}, "", err
	}
	fmt.Fprintf(p.Out, "No %s found. Creating %s; Enter takes the value in brackets.\n", fileName, path)

	var f file
	srcDefault := firstNonEmpty(srcHint, defaultSrc())
	if f.Src, err = p.ask("Checkout of the devkits repository (src)", srcDefault, nonEmpty); err != nil {
		return file{}, "", err
	}
	if f.Worker.Name, err = p.ask("User that runs podman (worker)", defaultWorker, nonEmpty); err != nil {
		return file{}, "", err
	}
	if f.Podman.Tmpdir, err = p.ask("Folder for the temporary layers of podman (tmpdir)", defaultTmpdir, func(v string) error {
		if !filepath.IsAbs(expand(v, ".")) {
			return errors.New("must be an absolute path")
		}
		return nil
	}); err != nil {
		return file{}, "", err
	}
	free, err := p.ask("Free space needed in tmpdir, MB (min_free_mb)", strconv.Itoa(defaultMinFreeMB), func(v string) error {
		if n, err := strconv.Atoi(v); err != nil || n < 0 {
			return errors.New("must be a number >= 0")
		}
		return nil
	})
	if err != nil {
		return file{}, "", err
	}
	n, _ := strconv.Atoi(free)
	f.Podman.MinFreeMB = &n
	if f.Images.Prefix, err = p.ask("Image name prefix (images.prefix)", defaultPrefix, nonEmpty); err != nil {
		return file{}, "", err
	}

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return file{}, "", err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return file{}, "", err
	}
	fmt.Fprintf(p.Out, "Wrote %s\n", path)
	return f, path, nil
}

func nonEmpty(v string) error {
	if v == "" {
		return errors.New("must not be empty")
	}
	return nil
}

// ask repeats the question until check accepts the answer; an empty answer means def.
func (p *Prompt) ask(question, def string, check func(string) error) (string, error) {
	for {
		fmt.Fprintf(p.Out, "%s [%s]: ", question, def)
		line, err := p.In.ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("no answer (%w); nothing written", err)
		}
		v := strings.TrimSpace(line)
		if v == "" {
			v = def
		}
		if cerr := check(v); cerr != nil {
			fmt.Fprintln(p.Out, "  ", cerr)
			continue
		}
		return v, nil
	}
}
