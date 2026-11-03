#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
config="${DEVKIT_CONFIG:-$repo_root/config/factory.json}"

worker=""
worker_uid=""
tmpdir=""
min_free_mb=""
image_prefix=""

if [ -f "$config" ]; then
    command -v jq >/dev/null 2>&1 || {
        echo "error: jq required for $config" >&2
        exit 2
    }

    worker="$(jq -r '.worker.name // empty' "$config")"
    worker_uid="$(jq -r '.worker.uid // empty' "$config")"
    tmpdir="$(jq -r '.podman.tmpdir // empty' "$config")"
    min_free_mb="$(jq -r '.podman.min_free_mb // empty' "$config")"
    image_prefix="$(jq -r '.images.prefix // empty' "$config")"
fi

worker="${DEVKIT_WORKER:-${worker:-factory-worker}}"
worker_uid="${DEVKIT_WORKER_UID:-${worker_uid:-1000}}"
src="${DEVKITS_DIR:-$HOME/repos/devkits}"
tmpdir="${DEVKIT_TMPDIR:-${tmpdir:-/server/factory-containers/tmp}}"
min_free_mb="${DEVKIT_MIN_FREE_MB:-${min_free_mb:-2048}}"
image_prefix="${DEVKIT_IMAGE_PREFIX:-${image_prefix:-localhost/factory-devkit-}}"

echo "Configured worker=$worker uid=$worker_uid"
echo "Configured tmpdir=$tmpdir"
echo "Configured image_prefix=$image_prefix"
echo "Integrate remaining original build-devkit.sh logic below this block."
