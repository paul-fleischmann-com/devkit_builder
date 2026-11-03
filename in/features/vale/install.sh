#!/usr/bin/env bash
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends vale || true
rm -rf /var/lib/apt/lists/*
