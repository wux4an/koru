#!/usr/bin/env bash
# Build the release archives of koru for every supported platform.
#
#   scripts/release.sh v0.0.1 [output-dir]
#
# Writes koru_<version>_<os>_<arch>.tar.gz (.zip on Windows), a .sha256 file
# next to each archive and a SHA256SUMS file covering all of them. Builds are
# reproducible: same commit and Go version give the same bytes. Needs GNU tar,
# coreutils and zip (the release workflow runs it on Ubuntu).
set -euo pipefail

tag="${1:?usage: scripts/release.sh vX.Y.Z [output-dir]}"
out="${2:-dist}"
[[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || {
  echo "release: '$tag' is not a tag like v0.0.1" >&2
  exit 1
}
version="${tag#v}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

declared="$(sed -n 's/^\tversion = "\(.*\)"$/\1/p' cmd/koru/main.go)"
if [ "$declared" != "$version" ]; then
  echo "release: cmd/koru/main.go declares version $declared, tag is $tag" >&2
  exit 1
fi

GO="${GO:-go}"
commit="$(git rev-parse --short=12 HEAD)"
# Archive timestamps come from the commit so rebuilding gives identical files.
epoch="$(git log -1 --format=%ct HEAD)"
platforms="${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64}"

sha256() {
  if command -v sha256sum >/dev/null; then sha256sum "$@"; else shasum -a 256 "$@"; fi
}

rm -rf "$out"
mkdir -p "$out"
out="$(cd "$out" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

for platform in $platforms; do
  os="${platform%/*}"
  arch="${platform#*/}"
  name="koru_${version}_${os}_${arch}"
  dir="$work/$name"
  exe="koru"
  [ "$os" = windows ] && exe="koru.exe"
  mkdir -p "$dir"
  echo "building $name"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" "$GO" build -trimpath -buildvcs=false \
    -ldflags "-s -w -buildid= -X main.version=$version -X main.commit=$commit" \
    -o "$dir/$exe" ./cmd/koru
  cp LICENSE README.md "$dir/"
  [ -f CHANGELOG.md ] && cp CHANGELOG.md "$dir/"
  find "$dir" -exec touch -h -d "@$epoch" {} +
  if [ "$os" = windows ]; then
    (cd "$work" && find "$name" -print | LC_ALL=C sort | zip -X -q -@ "$out/$name.zip")
    archive="$name.zip"
  else
    tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$epoch" \
      -C "$work" -cf - "$name" | gzip -9n >"$out/$name.tar.gz"
    archive="$name.tar.gz"
  fi
  (cd "$out" && sha256 "$archive" >"$archive.sha256")
done

(cd "$out" && cat ./*.sha256 | LC_ALL=C sort -k2 >SHA256SUMS && sha256 -c --quiet SHA256SUMS)
echo "release $tag ($commit) written to $out:"
ls -l "$out"
