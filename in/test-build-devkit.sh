#!/usr/bin/env bash
# Tests of in/build-devkit.sh with a fake curl and a fake devkit binary; nothing is fetched from the network.
#   bash in/test-build-devkit.sh
set -uo pipefail

SCRIPT_SRC="$(cd "$(dirname "$0")" && pwd)/build-devkit.sh"
FAILED=0
ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT

pass() { echo "PASS $1"; }
fail() { echo "FAIL $1"; FAILED=1; }
check() { # name, condition result
    if [ "$2" = 0 ]; then pass "$1"; else fail "$1"; fi
}

# a repository like devkit_builder: <root>/in/{build-devkit.sh,kits/}, <root>/config/factory.json
REPO="$ROOT/repo"
mkdir -p "$REPO/in/kits" "$REPO/config" "$ROOT/bin" "$ROOT/fixtures" "$ROOT/cache"
cp "$SCRIPT_SRC" "$REPO/in/build-devkit.sh"
echo '{}' > "$REPO/config/factory.json"
SCRIPT="$REPO/in/build-devkit.sh"

# the fake devkit: writes down its arguments and DEVKIT_CONFIG
FAKE="$ROOT/fixtures/devkit"
cat > "$FAKE" << 'EOS'
#!/usr/bin/env bash
echo "$*" > "$FAKE_LOG"
echo "config=${DEVKIT_CONFIG:-}" >> "$FAKE_LOG"
EOS
chmod +x "$FAKE"
export FAKE_LOG="$ROOT/devkit.log"

# the archive as the release has it: README.md and devkit at the top
printf 'readme\n' > "$ROOT/fixtures/README.md"
(cd "$ROOT/fixtures" && tar -czf "devkit_0.2.1_linux_amd64.tar.gz" devkit README.md)
GOOD_SHA="$(sha256sum "$ROOT/fixtures/devkit_0.2.1_linux_amd64.tar.gz" | cut -d' ' -f1)"

# the fake curl: copies the fixture named like the URL and writes down every call
cat > "$ROOT/bin/curl" << 'EOS'
#!/usr/bin/env bash
echo "$*" >> "$CURL_LOG"
out=""; url=""
while [ $# -gt 0 ]; do
    case "$1" in -o) out="$2"; shift 2 ;; http*) url="$1"; shift ;; *) shift ;; esac
done
src="$FIXTURES/$(basename "$url")"
[ -f "$src" ] || exit 22
cp "$src" "$out"
EOS
chmod +x "$ROOT/bin/curl"
export CURL_LOG="$ROOT/curl.log" FIXTURES="$ROOT/fixtures"
export PATH="$ROOT/bin:$PATH"

run() { : > "$FAKE_LOG"; "$SCRIPT" "$@" 2> "$ROOT/err.txt"; }
last() { sed -n 1p "$FAKE_LOG"; }

echo "== passing the commands on"
export DEVKIT_BIN="$FAKE"
run c-cmake; check "a kit becomes: build <kit> --src <folder of the script>" "$([ "$(last)" = "build c-cmake --src $REPO/in" ]; echo $?)"
run --list; check "--list becomes: list kits" "$([ "$(last)" = "list kits --src $REPO/in" ]; echo $?)"
run --images; check "--images becomes: images" "$([ "$(last)" = "images --src $REPO/in" ]; echo $?)"
run --print python; check "--print <kit> becomes: build <kit> --print" "$([ "$(last)" = "build python --print --src $REPO/in" ]; echo $?)"
run c-cmake --commit abc123 --tmpdir /x; check "options of build are passed on" "$([ "$(last)" = "build c-cmake --commit abc123 --tmpdir /x --src $REPO/in" ]; echo $?)"
run doctor; check "doctor is passed on" "$([ "$(last)" = "doctor --src $REPO/in" ]; echo $?)"
run validate; check "validate is passed on" "$([ "$(last)" = "validate --src $REPO/in" ]; echo $?)"
run version; check "version gets no --src" "$([ "$(last)" = "version" ]; echo $?)"
run init /tmp/new; check "init gets no --src" "$([ "$(last)" = "init /tmp/new" ]; echo $?)"
run c-cmake --src /other; check "an explicit --src is not doubled" "$([ "$(last)" = "build c-cmake --src /other" ]; echo $?)"
run c-cmake --src=/other; check "--src=DIR is recognised" "$([ "$(last)" = "build c-cmake --src=/other" ]; echo $?)"
DEVKITS_DIR=/elsewhere run c-cmake; check "DEVKITS_DIR set: no --src is added" "$([ "$(last)" = "build c-cmake" ]; echo $?)"
rmdir "$REPO/in/kits"; run c-cmake; check "no kits/ next to the script: no --src is added" "$([ "$(last)" = "build c-cmake" ]; echo $?)"; mkdir "$REPO/in/kits"
run c-cmake; check "DEVKIT_CONFIG is taken from config/factory.json" "$([ "$(sed -n 2p "$FAKE_LOG")" = "config=$REPO/config/factory.json" ]; echo $?)"
DEVKIT_CONFIG=/mine.json run c-cmake; check "a DEVKIT_CONFIG of the caller is kept" "$([ "$(sed -n 2p "$FAKE_LOG")" = "config=/mine.json" ]; echo $?)"
run; check "no argument: usage and exit code 2" "$([ $? = 2 ] && grep -q 'devkit build' "$ROOT/err.txt"; echo $?)"
run --nonsense; check "an unknown option is refused" "$([ $? = 2 ] && grep -q 'unknown option' "$ROOT/err.txt"; echo $?)"
run --print; check "--print without a kit is refused" "$([ $? = 2 ] && grep -q 'needs a kit' "$ROOT/err.txt"; echo $?)"
DEVKIT_BIN=/no/such run c-cmake; check "a DEVKIT_BIN that is not executable is refused" "$([ $? = 2 ]; echo $?)"
[ ! -f "$CURL_LOG" ]; check "with DEVKIT_BIN nothing is fetched" "$?"
unset DEVKIT_BIN

echo "== fetching the release"
export DEVKIT_CACHE_DIR="$ROOT/cache" DEVKIT_PLATFORM=linux_amd64 DEVKIT_SHA256="$GOOD_SHA"
run c-cmake; check "the first run fetches and runs the binary" "$([ "$(last)" = "build c-cmake --src $REPO/in" ]; echo $?)"
check "curl: https only, TLS 1.2, the archive of the version and platform" "$(grep -q -- "--proto =https --tlsv1.2" "$CURL_LOG" && grep -q "releases/download/v0.2.1/devkit_0.2.1_linux_amd64.tar.gz" "$CURL_LOG"; echo $?)"
check "the binary is in the cache, executable" "$([ -x "$ROOT/cache/v0.2.1/devkit" ]; echo $?)"
check "its SHA-256 is written next to it" "$([ "$(cat "$ROOT/cache/v0.2.1/devkit.sha256")" = "$(sha256sum "$ROOT/cache/v0.2.1/devkit" | cut -d' ' -f1)" ]; echo $?)"
: > "$CURL_LOG"; run doctor
check "the second run uses the cache and fetches nothing" "$([ ! -s "$CURL_LOG" ] && [ "$(last)" = "doctor --src $REPO/in" ]; echo $?)"
echo '#!/bin/sh' > "$ROOT/cache/v0.2.1/devkit"; echo 'echo tampered' >> "$ROOT/cache/v0.2.1/devkit"
: > "$CURL_LOG"; run c-cmake
check "a changed binary in the cache is replaced by a fresh checked one" "$([ -s "$CURL_LOG" ] && [ "$(last)" = "build c-cmake --src $REPO/in" ]; echo $?)"

rm -rf "$ROOT/cache"; mkdir "$ROOT/cache"; : > "$CURL_LOG"
DEVKIT_SHA256="$(printf 'x' | sha256sum | cut -d' ' -f1)" run c-cmake
check "a wrong SHA-256 stops the run" "$([ $? = 2 ] && grep -q 'does not match' "$ROOT/err.txt"; echo $?)"
check "and nothing is put into the cache" "$([ ! -e "$ROOT/cache/v0.2.1/devkit" ]; echo $?)"
check "and the downloaded file is not left behind" "$([ -z "$(find "${TMPDIR:-/tmp}" -maxdepth 1 -name 'tmp.*' -newer "$CURL_LOG" -exec grep -l devkit {} + 2> /dev/null)" ]; echo $?)"

DEVKIT_VERSION='../evil' run c-cmake; check "a version that is not vX.Y.Z is refused before anything is fetched" "$([ $? = 2 ] && grep -q 'must look like' "$ROOT/err.txt"; echo $?)"
DEVKIT_PLATFORM=windows_amd64 run c-cmake; check "an unsupported platform is refused" "$([ $? = 2 ] && grep -q 'unsupported platform' "$ROOT/err.txt"; echo $?)"

echo "== a version without a pinned hash"
unset DEVKIT_SHA256
cp "$ROOT/fixtures/devkit_0.2.1_linux_amd64.tar.gz" "$ROOT/fixtures/devkit_9.9.9_linux_amd64.tar.gz"
printf '%s  devkit_9.9.9_linux_amd64.tar.gz\n' "$GOOD_SHA" > "$ROOT/fixtures/checksums.txt"
: > "$CURL_LOG"; DEVKIT_VERSION=v9.9.9 run c-cmake
check "it is checked against checksums.txt of the release and says so" "$([ "$(last)" = "build c-cmake --src $REPO/in" ] && grep -q 'checksums.txt of the same release only' "$ROOT/err.txt" && grep -q 'v9.9.9/checksums.txt' "$CURL_LOG"; echo $?)"
printf '%s  devkit_9.9.9_linux_amd64.tar.gz\n' "$(printf 'y' | sha256sum | cut -d' ' -f1)" > "$ROOT/fixtures/checksums.txt"
rm -rf "$ROOT/cache"; mkdir "$ROOT/cache"
DEVKIT_VERSION=v9.9.9 run c-cmake
check "a wrong checksums.txt stops the run" "$([ $? = 2 ] && grep -q 'does not match' "$ROOT/err.txt"; echo $?)"
printf 'no matching line here\n' > "$ROOT/fixtures/checksums.txt"
DEVKIT_VERSION=v9.9.9 run c-cmake
check "no line for the archive in checksums.txt stops the run" "$([ $? = 2 ] && grep -q 'no usable SHA-256' "$ROOT/err.txt"; echo $?)"

echo "== the script itself"
check "four pinned hashes of 64 hex digits for v0.2.1" "$([ "$(grep -cE '^        v0\.2\.1/(darwin|linux)_(amd64|arm64)\) echo [0-9a-f]{64} ;;$' "$SCRIPT_SRC")" = 4 ]; echo $?)"
check "nothing is piped into a shell" "$(! grep -vE '^\s*#' "$SCRIPT_SRC" | grep -qE '\|\s*(ba|z)?sh\b'; echo $?)"
check "curl is called with --proto =https and --tlsv1.2 everywhere" "$([ "$(grep -c 'curl -fsSL' "$SCRIPT_SRC")" = "$(grep -c "curl -fsSL --proto '=https' --tlsv1.2" "$SCRIPT_SRC")" ]; echo $?)"

if [ "$FAILED" = 0 ]; then echo "test-build-devkit ok"; else echo "test-build-devkit FAILED" >&2; fi
exit "$FAILED"
