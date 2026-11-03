#!/usr/bin/env bash
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf binutils-arm-linux-gnueabihf
rm -rf /var/lib/apt/lists/*
