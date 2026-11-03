#!/usr/bin/env bash
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends jq
rm -rf /var/lib/apt/lists/*
