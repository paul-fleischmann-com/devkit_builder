// Package build implements `devkit build`: a devkit image from a commit of the devkits repository.
//
// The checkout must be clean; the build context is `git archive HEAD`, the Containerfile comes from
// the compose package, and podman runs as the factory worker.
package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/paul-fleischmann-com/devkit_builder/internal/cli"
	"github.com/paul-fleischmann-com/devkit_builder/internal/compose"
	"github.com/paul-fleischmann-com/devkit_builder/internal/config"
	"github.com/paul-fleischmann-com/devkit_builder/internal/podman"
)

// Run implements `devkit build <kit> [--src DIR] [--commit SHA] [--tmpdir DIR] [--print]`.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := cli.NewFlagSet("build", stderr)
	src := cli.SrcFlag(fs)
	wanted := fs.String("commit", "", "demand that HEAD of the checkout is exactly this commit")
	tmpdir := fs.String("tmpdir", "", "folder for the temporary layers of podman (default: config, $DEVKIT_TMPDIR)")
	print := fs.Bool("print", false, "only print the Containerfile (nothing is built, no worker needed)")
	pos, err := cli.Parse(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: devkit build <kit> [--src DIR] [--commit SHA] [--tmpdir DIR] [--print]")
		return 2
	}
	kit := pos[0]
	if err := run(context.Background(), stdout, stderr, *src, kit, *wanted, *tmpdir, *print); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		if e, ok := err.(failure); ok {
			return int(e)
		}
		return 2
	}
	return 0
}

// failure is an error with exit code 1 (the build ran and failed); everything else is a usage problem (2).
type failure int

func (failure) Error() string { return "failed" }

func run(ctx context.Context, stdout, stderr io.Writer, src, kit, wanted, tmpdir string, printOnly bool) error {
	cfg, err := config.Load(src, config.StdPrompt(stderr))
	if err != nil {
		return err
	}
	src = cfg.Src
	if tmpdir == "" {
		tmpdir = cfg.Tmpdir
	}

	if st, err := os.Stat(filepath.Join(src, ".git")); err != nil || !st.IsDir() {
		return fmt.Errorf("%s is not a checkout of the devkits repository (set --src or DEVKITS_DIR)", src)
	}
	if _, err := os.Stat(filepath.Join(src, "kits", kit+".json")); err != nil {
		if kits, kerr := compose.Kits(src); kerr == nil {
			return fmt.Errorf("no kit %s in %s/kits (have: %s)", kit, src, strings.Join(kits, ", "))
		}
		return fmt.Errorf("no kit %s in %s/kits (devkit list shows them)", kit, src)
	}

	if status, err := git(src, "status", "--porcelain"); err != nil {
		return err
	} else if status != "" {
		return fmt.Errorf("the checkout %s has changes that are not committed: a devkit is built from a commit", src)
	}
	commit, err := git(src, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if wanted != "" && !strings.HasPrefix(commit, wanted) {
		return fmt.Errorf("HEAD of %s is %s, not %s", src, commit[:12], wanted)
	}
	short := commit[:12]
	tag := cfg.ImagePrefix + kit + ":" + short

	containerfile, err := compose.Compose(src, kit, commit)
	if err != nil {
		return fmt.Errorf("compose refused the kit %s: %w", kit, err)
	}
	if printOnly {
		fmt.Fprint(stdout, containerfile)
		return nil
	}

	if !filepath.IsAbs(tmpdir) {
		return fmt.Errorf("--tmpdir must be an absolute path")
	}
	w := podman.Worker{Name: cfg.Worker, TmpDir: tmpdir}
	if _, err := w.Output(ctx, "test", "-d", tmpdir, "-a", "-w", tmpdir); err != nil {
		return fmt.Errorf("%s does not exist or is not writable for %s: run  sudo infra/factory-orch/podman-tmp.sh set", tmpdir, cfg.Worker)
	}
	freeMB, err := FreeMB(ctx, w, tmpdir)
	if err != nil || freeMB < cfg.MinFreeMB {
		got := "?"
		if err == nil {
			got = strconv.Itoa(freeMB)
		}
		return fmt.Errorf("only %s MB free in %s (need %d): free space or choose another folder with --tmpdir", got, tmpdir, cfg.MinFreeMB)
	}

	ctxDir, err := w.Output(ctx, "mktemp", "-d", "/tmp/factory-devkit.XXXXXX")
	if err != nil {
		return err
	}
	ctxDir = strings.TrimSpace(ctxDir)
	defer w.Output(context.Background(), "rm", "-rf", ctxDir)

	if err := extractArchive(ctx, w, src, ctxDir); err != nil {
		return err
	}
	tee, err := w.Command(ctx, "tee", ctxDir+"/Containerfile")
	if err != nil {
		return err
	}
	tee.Stdin = strings.NewReader(containerfile)
	if err := tee.Run(); err != nil {
		return fmt.Errorf("writing the Containerfile: %w", err)
	}

	fmt.Fprintf(stdout, "building %s from commit %s of %s (the build needs the network once, for the Debian snapshot)\n", tag, short, src)
	if err := stream(ctx, w, stdout, stderr, "podman", "build", "--pull=missing", "-t", tag, "-f", ctxDir+"/Containerfile", ctxDir); err != nil {
		return failure(1)
	}

	// a short check of the result: the user is not root, there is no claude and no token, the git of the test setup is there
	smoke := `test "$(id -u)" = 1000 && ! command -v claude > /dev/null && git --version`
	if err := stream(ctx, w, io.Discard, stderr, "podman", "run", "--rm", "--network=none", tag, "bash", "-c", smoke); err != nil {
		fmt.Fprintf(stderr, "error: the image %s failed its smoke test\n", tag)
		return failure(1)
	}

	id, err := w.Output(ctx, "podman", "image", "inspect", tag, "--format", "{{.Id}}")
	if err != nil {
		return failure(1)
	}
	id = strings.TrimSpace(id)
	info, err := w.Output(ctx, "podman", "image", "inspect", tag, "--format", "image-id={{.Id}} size={{.Size}} created={{.Created}}")
	if err != nil {
		return failure(1)
	}
	fmt.Fprint(stdout, info)
	fmt.Fprintf(stdout, "FACTORY_TEST_PROFILES entry: %s=%s   (commit %s)\n", kit, strings.TrimPrefix(id, "sha256:"), short)
	return nil
}

// FreeMB returns the space available in dir, measured as the worker (df).
func FreeMB(ctx context.Context, w podman.Worker, dir string) (int, error) {
	out, err := w.Output(ctx, "df", "--output=avail", "-B1M", dir)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return strconv.Atoi(strings.TrimSpace(lines[len(lines)-1]))
}

// extractArchive pipes `git archive HEAD` of src into `tar -x` of the worker.
func extractArchive(ctx context.Context, w podman.Worker, src, dir string) error {
	archive := exec.CommandContext(ctx, "git", "-C", src, "archive", "HEAD")
	tar, err := w.Command(ctx, "tar", "-x", "-C", dir)
	if err != nil {
		return err
	}
	pipe, err := archive.StdoutPipe()
	if err != nil {
		return err
	}
	tar.Stdin = pipe
	archive.Stderr, tar.Stderr = os.Stderr, os.Stderr
	if err := tar.Start(); err != nil {
		return err
	}
	if err := archive.Run(); err != nil {
		_ = tar.Process.Kill()
		_ = tar.Wait()
		return fmt.Errorf("git archive: %w", err)
	}
	if err := tar.Wait(); err != nil {
		return fmt.Errorf("tar: %w", err)
	}
	return nil
}

func stream(ctx context.Context, w podman.Worker, stdout, stderr io.Writer, args ...string) error {
	cmd, err := w.Command(ctx, args...)
	if err != nil {
		return err
	}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return cmd.Run()
}

func git(src string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", src}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}
