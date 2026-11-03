#!/usr/bin/env bash
set -euo pipefail

apt-get update

apt-get install -y --no-install-recommends \
    gawk \
    diffstat \
    unzip \
    texinfo \
    chrpath \
    socat \
    cpio \
    file \
    xz-utils \
    zstd \
    lz4 \
    locales \
    iputils-ping

rm -rf /var/lib/apt/lists/*
