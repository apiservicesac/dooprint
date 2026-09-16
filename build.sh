#!/usr/bin/env bash
# Builds dooprint inside Docker: neither Node nor Go needs to be installed.
# Output: dist/dooprint (Linux x86_64) and dist/dooprint.exe (Windows x86_64), with the web
# interface embedded and USB support in both.
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p dist

echo "== web interface"
docker run --rm -v "$PWD":/src -w /src/frontend node:22-bookworm bash -c "
    npm install --no-audit --no-fund --silent
    npm run build
    rm -rf /src/internal/webui/dist && cp -r dist /src/internal/webui/dist
    chown -R $(id -u):$(id -g) node_modules dist /src/internal/webui/dist
"

echo "== binaries"
# For Windows, libusb is cross-compiled with mingw and linked statically, so the .exe needs no DLL.
docker run --rm -v "$PWD":/src -w /src golang:1.25-bookworm bash -c "
    set -e
    apt-get update -qq
    apt-get install -y -qq libusb-1.0-0-dev pkg-config gcc-mingw-w64-x86-64 curl make bzip2 >/dev/null 2>&1

    echo '-- linux'
    CGO_ENABLED=1 go build -buildvcs=false -trimpath -ldflags '-s -w' -o dist/dooprint ./cmd/dooprint

    echo '-- libusb for windows'
    curl -sL https://github.com/libusb/libusb/releases/download/v1.0.27/libusb-1.0.27.tar.bz2 -o /tmp/libusb.tar.bz2
    mkdir -p /tmp/libusb && tar -xjf /tmp/libusb.tar.bz2 -C /tmp/libusb --strip-components=1
    cd /tmp/libusb
    ./configure --host=x86_64-w64-mingw32 --prefix=/opt/libusb-win --enable-static --disable-shared >/dev/null
    make -j4 >/dev/null && make install >/dev/null
    cd /src

    echo '-- windows'
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \\
        PKG_CONFIG_PATH=/opt/libusb-win/lib/pkgconfig \\
        go build -buildvcs=false -trimpath -ldflags '-s -w' -o dist/dooprint.exe ./cmd/dooprint

    chown $(id -u):$(id -g) dist/dooprint dist/dooprint.exe
"
ls -lh dist/dooprint dist/dooprint.exe
