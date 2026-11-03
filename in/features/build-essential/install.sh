#!/usr/bin/env bash
# Feature build-essential: gcc, g++, make and pkg-config
# Runs as root inside the build, once. Packages come from the dated Debian snapshot of the kit (see scripts/compose.py): the same kit gives the same image.
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends build-essential pkg-config
dpkg-query -W -f='${Package} ${Version}\n' build-essential pkg-config
