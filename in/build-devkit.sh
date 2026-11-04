#!/usr/bin/env bash
# Builds a devkit image with the `devkit` tool from the releases of
# https://github.com/paul-fleischmann-com/devkit_builder/releases
# (a pinned version, the SHA-256 of the archive checked, cached per user; nothing is piped into a shell).
#
#   in/build-devkit.sh <kit> [--src DIR] [--commit SHA] [--tmpdir DIR]   devkit build <kit> ...
#   in/build-devkit.sh --print <kit>                                     devkit build <kit> --print   (the Containerfile only)
#   in/build-devkit.sh --list                                            devkit list kits
#   in/build-devkit.sh --images                                          devkit images
#   in/build-devkit.sh init|list|validate|compose|build|images|doctor|version ...   passed on to devkit as they are
#
# --src defaults to the folder of this script when it has kits/ (unless DEVKITS_DIR is set or --src is given).
# Settings: DEVKIT_CONFIG (else config/factory.json next to the repository, else factory.json next to this script, else what devkit looks up itself).
# For the fetch (all optional):
#   DEVKIT_VERSION       the release, default v0.2.1; another version is checked against the checksums.txt of the same release only (and says so)
#   DEVKIT_SHA256        the expected SHA-256 of the archive (instead of the pinned one / checksums.txt)
#   DEVKIT_BIN           use this binary, nothing is fetched
#   DEVKIT_CACHE_DIR     where the binaries are kept, default ${XDG_CACHE_HOME:-~/.cache}/devkit
#   DEVKIT_PLATFORM      <os>_<arch> (linux_amd64, linux_arm64, darwin_amd64, darwin_arm64), default from uname
#   DEVKIT_RELEASES_URL  default https://github.com/paul-fleischmann-com/devkit_builder/releases/download
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$here/.." && pwd)"

VERSION="${DEVKIT_VERSION:-v0.2.1}"
RELEASES="${DEVKIT_RELEASES_URL:-https://github.com/paul-fleischmann-com/devkit_builder/releases/download}"
CACHE="${DEVKIT_CACHE_DIR:-${XDG_CACHE_HOME:-$HOME/.cache}/devkit}"

die() { echo "error: $*" >&2; exit 2; }
usage() { awk 'NR > 1 { if (/^#/) { sub(/^# ?/, ""); print } else exit }' "$0"; }

# The SHA-256 of the archives of the pinned release v0.2.1 (checksums.txt of the release; the one of linux_amd64 was also computed from the downloaded file).
pinned_sha() {
    case "$VERSION/$1" in
        v0.2.1/darwin_amd64) echo 23d630c51fee43fec73a70ad0f0bca4f1f96875bd689e423e777136ae0cc340f ;;
        v0.2.1/darwin_arm64) echo 1e04512c18b685b7d73d08495db2cc2b672984973e0219f62ac21e789b7c86b9 ;;
        v0.2.1/linux_amd64) echo caf8ecc3dd077175e73a02eba237865b9ecbc08faf1c32163cdc44986c6f9c79 ;;
        v0.2.1/linux_arm64) echo c4fff8ec63239bed7a4a8cd894a202135584b007363f76d8b4e79394db43da01 ;;
        *) echo "" ;;
    esac
}

sha256() {
    if command -v sha256sum > /dev/null 2>&1; then sha256sum "$1" | cut -d' ' -f1
    elif command -v shasum > /dev/null 2>&1; then shasum -a 256 "$1" | cut -d' ' -f1
    else die "sha256sum or shasum is needed to check the download"; fi
}

platform() {
    if [ -n "${DEVKIT_PLATFORM:-}" ]; then echo "$DEVKIT_PLATFORM"; return; fi
    local os arch
    case "$(uname -s)" in Linux) os=linux ;; Darwin) os=darwin ;; *) die "unsupported system $(uname -s): use DEVKIT_BIN or a Linux or macOS machine" ;; esac
    case "$(uname -m)" in x86_64 | amd64) arch=amd64 ;; aarch64 | arm64) arch=arm64 ;; *) die "unsupported processor $(uname -m)" ;; esac
    echo "${os}_${arch}"
}

# prints the path of the devkit binary: DEVKIT_BIN, or the checked one in the cache, or fetches it
ensure_devkit() {
    if [ -n "${DEVKIT_BIN:-}" ]; then
        [ -x "$DEVKIT_BIN" ] || die "DEVKIT_BIN=$DEVKIT_BIN is not an executable file"
        echo "$DEVKIT_BIN"
        return
    fi
    [[ $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "DEVKIT_VERSION must look like v0.2.1, not '$VERSION'"
    local plat plain bin archive url expected actual tmp
    plat="$(platform)"
    case "$plat" in linux_amd64 | linux_arm64 | darwin_amd64 | darwin_arm64) ;; *) die "unsupported platform '$plat'" ;; esac
    bin="$CACHE/$VERSION/devkit"
    if [ -x "$bin" ] && [ -f "$bin.sha256" ] && [ "$(sha256 "$bin")" = "$(cat "$bin.sha256")" ]; then
        echo "$bin"
        return
    fi
    command -v curl > /dev/null 2>&1 || die "curl is needed to fetch devkit $VERSION (or set DEVKIT_BIN)"
    command -v tar > /dev/null 2>&1 || die "tar is needed to unpack devkit $VERSION"
    plain="${VERSION#v}"
    archive="devkit_${plain}_${plat}.tar.gz"
    url="$RELEASES/$VERSION"
    tmp="$(mktemp -d)"
    echo "fetching devkit $VERSION ($plat) from $url ..." >&2
    curl -fsSL --proto '=https' --tlsv1.2 --max-time 300 -o "$tmp/$archive" "$url/$archive" || { rm -rf "$tmp"; die "could not fetch $url/$archive"; }
    expected="${DEVKIT_SHA256:-$(pinned_sha "$plat")}"
    if [ -z "$expected" ]; then
        curl -fsSL --proto '=https' --tlsv1.2 --max-time 60 -o "$tmp/checksums.txt" "$url/checksums.txt" || { rm -rf "$tmp"; die "could not fetch $url/checksums.txt"; }
        expected="$(awk -v f="$archive" '$2 == f {print $1}' "$tmp/checksums.txt" | head -n 1)"
        echo "note: no pinned SHA-256 for $VERSION; checked against checksums.txt of the same release only" >&2
    fi
    [[ $expected =~ ^[0-9a-f]{64}$ ]] || { rm -rf "$tmp"; die "no usable SHA-256 for $archive"; }
    actual="$(sha256 "$tmp/$archive")"
    if [ "$actual" != "$expected" ]; then
        rm -rf "$tmp"
        die "the SHA-256 of $archive does not match (expected ${expected:0:16}..., got ${actual:0:16}...): the file is NOT used"
    fi
    tar -xzf "$tmp/$archive" -C "$tmp" devkit || { rm -rf "$tmp"; die "$archive does not contain 'devkit'"; }
    { [ -f "$tmp/devkit" ] && [ ! -L "$tmp/devkit" ]; } || { rm -rf "$tmp"; die "'devkit' in $archive is not a plain file"; }
    mkdir -p "$CACHE/$VERSION"
    install -m 0755 "$tmp/devkit" "$bin"
    sha256 "$bin" > "$bin.sha256"
    rm -rf "$tmp"
    echo "$bin"
}

[ $# -gt 0 ] || { usage >&2; exit 2; }

cmd=()
case "$1" in
    -h | --help) usage; exit 0 ;;
    --list) shift; cmd=(list kits "$@") ;;
    --images) shift; cmd=(images "$@") ;;
    --print)
        shift
        [ $# -gt 0 ] || die "--print needs a kit"
        kit="$1"
        shift
        cmd=(build "$kit" --print "$@")
        ;;
    init | list | validate | compose | build | images | doctor | version | help) cmd=("$@") ;;
    -*) die "unknown option $1 (see --help)" ;;
    *) cmd=(build "$@") ;;
esac

# --src: the folder of this script if it is a devkits checkout, unless the caller said otherwise
case "${cmd[0]}" in
    init | version | help) ;;
    *)
        has_src=0
        for arg in "${cmd[@]}"; do
            case "$arg" in --src | --src=* | --root | --root=*) has_src=1 ;; esac
        done
        if [ "$has_src" = 0 ] && [ -z "${DEVKITS_DIR:-}" ] && [ -d "$here/kits" ]; then cmd+=(--src "$here"); fi
        ;;
esac

if [ -z "${DEVKIT_CONFIG:-}" ]; then
    if [ -f "$repo_root/config/factory.json" ]; then export DEVKIT_CONFIG="$repo_root/config/factory.json"
    elif [ -f "$here/factory.json" ]; then export DEVKIT_CONFIG="$here/factory.json"; fi
fi

bin="$(ensure_devkit)"
exec "$bin" "${cmd[@]}"
