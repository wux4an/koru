#!/bin/sh
# Install the koru CLI on Linux or macOS.
#
#   curl -fsSL https://koru.js.org/install.sh | sh
#   curl -fsSL https://koru.js.org/install.sh | sh -s -- --version v0.0.1
#
# Options (or environment variables):
#   --version vX.Y.Z   KORU_VERSION       release to install (default: latest)
#   --dir DIR          KORU_INSTALL_DIR   where to put koru (default: ~/.local/bin)
#
# The archive is checked against the SHA-256 published with the release before
# anything is installed. Source: https://github.com/wux4an/koru
set -eu

repo="wux4an/koru"
base="${KORU_DOWNLOAD_URL:-https://github.com/$repo/releases}"
# Only HTTPS, unless a mirror for testing is given explicitly.
proto="=https"
[ -n "${KORU_DOWNLOAD_URL:-}" ] && proto="=https,http,file"

say() { printf '%s\n' "$*"; }
fail() {
  printf 'koru install: %s\n' "$*" >&2
  exit 1
}

fetch() { # fetch URL FILE
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --proto "$proto" --retry 3 -o "$2" "$1"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$2" "$1"
  else
    fail "curl or wget is required"
  fi
}

latest() {
  # /releases/latest redirects to /releases/tag/vX.Y.Z; reading the redirect
  # avoids the GitHub API and its rate limit.
  if command -v curl >/dev/null 2>&1; then
    url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$base/latest") || return 1
  else
    url=$(wget -q -S --spider "$base/latest" 2>&1 | sed -n 's/^ *[Ll]ocation: *//p' | tail -n 1 | tr -d '\r') || return 1
  fi
  tag=${url##*/}
  case $tag in v[0-9]*) printf '%s\n' "$tag" ;; *) return 1 ;; esac
}

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d ' ' -f 1
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | cut -d ' ' -f 1
  else
    fail "sha256sum or shasum is required to verify the download"
  fi
}

main() {
  version="${KORU_VERSION:-}"
  dir="${KORU_INSTALL_DIR:-$HOME/.local/bin}"
  while [ $# -gt 0 ]; do
    case $1 in
      --version) [ $# -ge 2 ] || fail "--version needs a value"; version=$2; shift 2 ;;
      --dir) [ $# -ge 2 ] || fail "--dir needs a value"; dir=$2; shift 2 ;;
      -h | --help)
        say "usage: install.sh [--version vX.Y.Z] [--dir DIR]"
        say "  KORU_VERSION and KORU_INSTALL_DIR set the same options."
        exit 0
        ;;
      *) fail "unknown option $1 (see --help)" ;;
    esac
  done

  case $(uname -s) in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    MINGW* | MSYS* | CYGWIN*) fail "on Windows run in PowerShell: irm https://koru.js.org/install.ps1 | iex" ;;
    *) fail "unsupported system $(uname -s); download a build from https://github.com/$repo/releases" ;;
  esac
  case $(uname -m) in
    x86_64 | amd64) arch=amd64 ;;
    arm64 | aarch64) arch=arm64 ;;
    *) fail "unsupported architecture $(uname -m); build from source: https://github.com/$repo" ;;
  esac

  if [ -z "$version" ]; then
    version=$(latest) || fail "could not find the latest release at $base/latest"
  fi
  case $version in v*) ;; *) version="v$version" ;; esac
  name="koru_${version#v}_${os}_${arch}"
  url="$base/download/$version/$name.tar.gz"

  tmp=$(mktemp -d 2>/dev/null || mktemp -d -t koru)
  trap 'rm -rf "$tmp"' EXIT INT TERM
  say "Downloading koru $version for $os/$arch"
  fetch "$url" "$tmp/$name.tar.gz" || fail "download failed: $url"
  fetch "$url.sha256" "$tmp/$name.tar.gz.sha256" || fail "checksum download failed: $url.sha256"

  want=$(cut -d ' ' -f 1 <"$tmp/$name.tar.gz.sha256")
  got=$(sha256_of "$tmp/$name.tar.gz")
  [ -n "$want" ] && [ "$want" = "$got" ] || fail "checksum mismatch for $name.tar.gz (expected $want, got $got)"
  say "Checksum verified ($got)"

  tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
  [ -f "$tmp/$name/koru" ] || fail "archive has no koru binary"
  mkdir -p "$dir"
  # Install through a temporary name so a running koru is never half-written.
  cp "$tmp/$name/koru" "$dir/.koru.new"
  chmod 755 "$dir/.koru.new"
  mv -f "$dir/.koru.new" "$dir/koru"
  say "Installed $("$dir/koru" version 2>/dev/null || echo "koru $version") to $dir/koru"

  case ":$PATH:" in
    *":$dir:"*) say "Run: koru create my-app" ;;
    *)
      say ""
      say "$dir is not in your PATH. Add it, for example:"
      say "  echo 'export PATH=\"$dir:\$PATH\"' >> ~/.profile && . ~/.profile"
      ;;
  esac
}

# Everything runs from main, so a download cut short never runs half a script.
main "$@"
