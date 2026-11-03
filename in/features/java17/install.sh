#!/usr/bin/env bash
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends openjdk-17-jre-headless
rm -rf /var/lib/apt/lists/*
