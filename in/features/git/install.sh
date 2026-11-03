#!/usr/bin/env bash
# Feature git: Git, for git archive / git apply of the test setup of the factory
# Runs as root inside the build, once. Packages come from the dated Debian snapshot of the kit (see scripts/compose.py): the same kit gives the same image.
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends git
dpkg-query -W -f='${Package} ${Version}\n' git
