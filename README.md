# devkit

Command line tool that composes and builds devkit images: a base image (by digest) plus features, in the order of a kit file.
Port of `compose.py` and `infra/factory-test/build-devkit.sh`; standard library only.

```
devkit init [dir]            new devkits repository from the reference devkit
devkit list [kits|features]
devkit validate              base image, all kits, all features
devkit compose <kit>         print the Containerfile (--commit SHA adds a label)
devkit build <kit>           image <prefix><kit>:<commit> from the clean checkout (git archive HEAD), podman as factory-worker via sudo,
                             smoke test, FACTORY_TEST_PROFILES line; --commit SHA, --tmpdir DIR, --print (Containerfile only)
devkit images                built devkit images + FACTORY_TEST_PROFILES entries
devkit doctor                git, podman as worker, tmpdir space, repository
devkit version
```

All commands except `init` and `version` take `--src DIR` (alias `--root`): the checkout of the devkits repository
(default: `$DEVKITS_DIR`, `./` if it has `kits/`, else `~/repos/devkits`).

## Installation

Download the archive for your platform from the [releases](https://github.com/paul-fleischmann-com/devkit_builder/releases)
(`devkit_<version>_<os>_<arch>.tar.gz`, `.zip` on Windows; linux/darwin/windows, amd64/arm64), verify it against `checksums.txt`
and put the `devkit` binary on your `PATH`. Or build it yourself: `go install ./cmd/devkit`.

## Repository layout

`base.json`, `kits/<name>.json`, `features/<id>/{feature.json,install.sh}`, optional `factory.json` (or `config/factory.json`).

## Configuration

`factory.json` is looked up in `$DEVKIT_CONFIG`, then in the current directory (`factory.json`, `config/factory.json`), then in the checkout.
If there is none, `build`, `images` and `doctor` ask on the terminal for `src`, `worker.name`, `podman.tmpdir`, `podman.min_free_mb`
and `images.prefix` and create `./factory.json`. Without a terminal (scripts, CI) they use the defaults and create nothing.

```json
{
  "src": "~/repos/devkits",
  "worker": { "name": "factory-worker" },
  "podman": { "tmpdir": "/server/factory-containers/tmp", "min_free_mb": 2048 },
  "images": { "prefix": "localhost/factory-devkit-" }
}
```

`~`, `$VAR` and paths relative to the file are expanded. Precedence: flag, environment, file, default.
Environment: `DEVKITS_DIR`, `DEVKIT_WORKER`, `DEVKIT_TMPDIR`, `DEVKIT_MIN_FREE_MB`, `DEVKIT_IMAGE_PREFIX`.
