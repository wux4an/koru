#!/usr/bin/env bash
# Every check CI runs, on one Linux machine. Go cross-compiles the other
# platforms, so no macOS or Windows runner is needed.
#
#   scripts/ci.sh                    # release archives go to a temporary folder
#   RELEASE_DIR=release scripts/ci.sh
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
GO="${GO:-go}"
export GO
step() { printf '\n==> %s\n' "$*"; }

step "gofmt"
mapfile -t sources < <(git ls-files '*.go')
unformatted="$(gofmt -l "${sources[@]}")"
if [ -n "$unformatted" ]; then
  echo "Run gofmt on:"
  echo "$unformatted"
  exit 1
fi

step "No external processes in the build"
# AGENTS.md: os/exec only in tests/external (build tag android_external).
if git grep -n '"os/exec"' -- '*.go' ':!tests/external/**'; then exit 1; fi

step "go vet for linux, darwin and windows"
for os in linux darwin windows; do
  GOOS="$os" "$GO" vet ./...
done

step "Tests compile for darwin and windows"
for os in darwin windows; do
  for pkg in $("$GO" list ./...); do
    GOOS="$os" "$GO" test -c -o /dev/null "$pkg"
  done
done

step "go test -race"
"$GO" test -race ./...

step "JavaScript module"
node --test --test-reporter=dot packages/javascript/core.test.cjs

step "Shell and PowerShell scripts"
shellcheck scripts/*.sh
if command -v pwsh >/dev/null; then
  # shellcheck disable=SC2016 # PowerShell code, expanded by pwsh
  pwsh -NoProfile -Command '
    $errors = $null
    [System.Management.Automation.Language.Parser]::ParseFile("scripts/install.ps1", [ref]$null, [ref]$errors) | Out-Null
    if ($errors) { $errors | ForEach-Object { Write-Error $_.ToString() }; exit 1 }'
else
  echo "pwsh not found: skipping the install.ps1 syntax check"
fi

version="$(sed -n 's/^\tversion = "\(.*\)"$/\1/p' cmd/koru/main.go)"
release="${RELEASE_DIR:-$(mktemp -d)/release}"
step "Release archives for every platform (v$version)"
scripts/release.sh "v$version" "$release"

step "Smoke test: create, build and verify an APK"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
tar -xzf "$release/koru_${version}_linux_$("$GO" env GOARCH).tar.gz" -C "$work"
koru="$work/koru_${version}_linux_$("$GO" env GOARCH)/koru"
"$koru" version
(cd "$work" && "$koru" create smoke-app --template basic --name Smoke --app-id com.example.smoke </dev/null >/dev/null &&
  "$koru" build smoke-app --output smoke.apk && "$koru" verify smoke.apk)

step "Homebrew formula and Scoop manifest"
pm="$work/pm"
mkdir -p "$pm/scripts"
cp scripts/package-managers.sh "$pm/scripts/"
"$pm/scripts/package-managers.sh" "v$version" "$release/SHA256SUMS" >/dev/null
ruby -c "$pm/Formula/koru.rb" >/dev/null
python3 -m json.tool "$pm/bucket/koru.json" >/dev/null
grep -q "\"version\": \"$version\"" "$pm/bucket/koru.json"
echo "ok - Formula/koru.rb and bucket/koru.json generate and parse"

step "Installer against the archives"
scripts/test-install.sh "$release"
