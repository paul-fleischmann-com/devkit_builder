#!/usr/bin/env bash
# Feature cppcheck: cppcheck, for the MISRA check and static analysis (#390)
# Runs as root inside the build, once. Packages come from the dated Debian snapshot of the kit (see scripts/compose.py): the same kit gives the same image.
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends cppcheck
dpkg-query -W -f='${Package} ${Version}\n' cppcheck
