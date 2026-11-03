#!/usr/bin/env bash
# Feature python3: Python 3 and pytest from the distribution
# Runs as root inside the build, once. Packages come from the dated Debian snapshot of the kit (see scripts/compose.py): the same kit gives the same image.
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends python3 python3-pytest
dpkg-query -W -f='${Package} ${Version}\n' python3 python3-pytest
