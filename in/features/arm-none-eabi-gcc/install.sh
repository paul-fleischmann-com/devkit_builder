#!/usr/bin/env bash
# Feature arm-none-eabi-gcc: Cross compiler for ARM Cortex-M (bare metal, newlib)
# Runs as root inside the build, once. Packages come from the dated Debian snapshot of the kit (see scripts/compose.py): the same kit gives the same image.
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends gcc-arm-none-eabi binutils-arm-none-eabi libnewlib-arm-none-eabi libstdc++-arm-none-eabi-newlib
dpkg-query -W -f='${Package} ${Version}\n' gcc-arm-none-eabi binutils-arm-none-eabi libnewlib-arm-none-eabi libstdc++-arm-none-eabi-newlib
