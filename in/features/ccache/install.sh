#!/usr/bin/env bash
set -euo pipefail

apt-get update

apt-get install -y --no-install-recommends \
    ccache

rm -rf /var/lib/apt/lists/*
