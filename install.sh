#!/bin/sh
# Download the latest prebuilt herdr-jump binary from GitHub Releases.
#
# Used by scripts/build.sh when no Go toolchain is present, so
# `herdr plugin install agustinvalencia/herdr-jump` works without Go. You can
# also run it directly to drop the binary on your PATH:
#
#   INSTALL_DIR=/usr/local/bin sh install.sh
#
# Release archives are named without a version so the "latest" redirect resolves
# them directly (see .goreleaser.yml).
set -eu

REPO="agustinvalencia/herdr-jump"
INSTALL_DIR="${INSTALL_DIR:-./bin}"

os="$(uname -s)"
case "$os" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) echo "herdr-jump: unsupported OS: $os" >&2; exit 1 ;;
esac

arch="$(uname -m)"
case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	arm64 | aarch64) arch="arm64" ;;
	*) echo "herdr-jump: unsupported arch: $arch" >&2; exit 1 ;;
esac

url="https://github.com/${REPO}/releases/latest/download/herdr-jump_${os}_${arch}.tar.gz"

mkdir -p "$INSTALL_DIR"
echo "herdr-jump: downloading $url" >&2
curl -fsSL "$url" | tar -xzf - -C "$INSTALL_DIR" herdr-jump
chmod +x "$INSTALL_DIR/herdr-jump"
echo "herdr-jump: installed to $INSTALL_DIR/herdr-jump" >&2
