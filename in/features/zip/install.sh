#!/usr/bin/env bash
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends zip unzip
rm -rf /var/lib/apt/lists/*
