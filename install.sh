#!/usr/bin/env sh
# hasdata-cli installer — downloads the latest release matching the host
# OS/arch from GitHub, verifies the checksum against the published
# checksums.txt, and installs to $PREFIX (default: /usr/local/bin, or
# $HOME/.local/bin if the former is not writable).
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/HasData/hasdata-cli/main/install.sh | sh
#   curl -sSL .../install.sh | HASDATA_VERSION=v1.2.3 sh
#   curl -sSL .../install.sh | PREFIX=$HOME/bin sh
set -eu

REPO="HasData/hasdata-cli"
BIN="hasdata"
VERSION="${HASDATA_VERSION:-latest}"

die() { printf 'install: %s\n' "$*" >&2; exit 1; }
log() { printf 'install: %s\n' "$*"; }

need() { command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"; }
need uname
need mkdir
need mv
need rm

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH_RAW=$(uname -m)
case "$OS" in
  linux|darwin) ;;
  mingw*|msys*|cygwin*) die "on Windows use Scoop or winget instead: scoop install hasdata" ;;
  *) die "unsupported OS: $OS" ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH=x86_64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) die "unsupported arch: $ARCH_RAW" ;;
esac

# Determine version.
if [ "$VERSION" = "latest" ]; then
  need curl
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)
  [ -n "$VERSION" ] || die "failed to resolve latest version"
fi

# Title-case OS matching goreleaser defaults (Linux/Darwin).
case "$OS" in
  linux) OS_PRETTY=Linux ;;
  darwin) OS_PRETTY=Darwin ;;
esac

VER_NO_V=${VERSION#v}
ASSET="${BIN}_${VER_NO_V}_${OS_PRETTY}_${ARCH}.tar.gz"
BASE="https://github.com/${REPO}/releases/download/${VERSION}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

log "downloading ${ASSET}"
need curl
curl -fsSL -o "$TMP/$ASSET" "$BASE/$ASSET" || die "download failed: $BASE/$ASSET"
curl -fsSL -o "$TMP/checksums.txt" "$BASE/checksums.txt" || die "failed to download checksums.txt"

# Verify checksum.
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$TMP" && grep " $ASSET$" checksums.txt | sha256sum -c -) || die "checksum mismatch"
elif command -v shasum >/dev/null 2>&1; then
  (cd "$TMP" && grep " $ASSET$" checksums.txt | shasum -a 256 -c -) || die "checksum mismatch"
else
  die "neither sha256sum nor shasum available to verify checksum"
fi
log "checksum ok"

# Extract.
need tar
tar -xzf "$TMP/$ASSET" -C "$TMP"
[ -x "$TMP/$BIN" ] || die "binary missing after extract"

# Choose install directory.
if [ -z "${PREFIX:-}" ]; then
  if [ -w "/usr/local/bin" ] || { [ -w "/usr/local" ] && mkdir -p /usr/local/bin 2>/dev/null; }; then
    PREFIX=/usr/local/bin
  else
    PREFIX="$HOME/.local/bin"
    mkdir -p "$PREFIX"
    case ":$PATH:" in
      *:"$PREFIX":*) ;;
      *) log "note: $PREFIX is not in PATH — add 'export PATH=\"$PREFIX:\$PATH\"' to your shell rc" ;;
    esac
  fi
fi

mv "$TMP/$BIN" "$PREFIX/$BIN"
chmod +x "$PREFIX/$BIN"
log "installed $BIN $VERSION to $PREFIX/$BIN"
"$PREFIX/$BIN" version || true
