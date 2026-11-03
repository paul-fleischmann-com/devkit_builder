#!/usr/bin/env bash
set -euo pipefail

apt-get update

apt-get install -y --no-install-recommends \
    clang-format

rm -rf /var/lib/apt/lists/*