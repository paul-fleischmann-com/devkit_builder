#!/usr/bin/env bash
# Feature cmake-ninja: CMake (with CTest) and Ninja
# Runs as root inside the build, once. Packages come from the dated Debian snapshot of the kit (see scripts/compose.py): the same kit gives the same image.
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends cmake ninja-build
dpkg-query -W -f='${Package} ${Version}\n' cmake ninja-build
