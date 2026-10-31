#!/usr/bin/env bash
# End-to-end test of scripts/install.sh against a local copy of a release:
# latest-version lookup, pinned versions, checksum verification and errors.
#
#   scripts/test-install.sh [release-dir]   # default: builds one with release.sh
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$(mktemp -d)"
trap 'kill "${server:-0}" 2>/dev/null || true; rm -rf "$work"' EXIT
version="$(sed -n 's/^\tversion = "\(.*\)"$/\1/p' "$root/cmd/koru/main.go")"
tag="v$version"

release="${1:-}"
if [ -z "$release" ]; then
  release="$work/release"
  PLATFORMS="$(go env GOOS)/$(go env GOARCH)" "$root/scripts/release.sh" "$tag" "$release" >/dev/null
fi
mkdir -p "$work/mirror/download/$tag"
cp "$release"/* "$work/mirror/download/$tag/"

# GitHub answers /releases/latest with a redirect to /releases/tag/<tag>.
cat >"$work/server.py" <<PY
import http.server, sys
class H(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *a, **k): super().__init__(*a, directory=sys.argv[2], **k)
    def do_HEAD(self):
        if self.path == "/latest":
            self.send_response(302); self.send_header("Location", "/tag/$tag"); self.end_headers(); return
        if self.path == "/tag/$tag":
            self.send_response(200); self.end_headers(); return
        super().do_HEAD()
    def do_GET(self):
        if self.path in ("/latest", "/tag/$tag"): return self.do_HEAD()
        super().do_GET()
    def log_message(self, *a): pass
http.server.ThreadingHTTPServer(("127.0.0.1", int(sys.argv[1])), H).serve_forever()
PY
port=$((20000 + RANDOM % 20000))
python3 "$work/server.py" "$port" "$work/mirror" &
server=$!
for _ in $(seq 50); do curl -fsI "http://127.0.0.1:$port/latest" >/dev/null 2>&1 && break; sleep 0.1; done
export KORU_DOWNLOAD_URL="http://127.0.0.1:$port"

check() { echo "ok - $*"; }

sh "$root/scripts/install.sh" --dir "$work/bin" >"$work/log"
"$work/bin/koru" version | grep -q "koru $version" && check "installs the latest release"

KORU_VERSION="$version" KORU_INSTALL_DIR="$work/pinned" sh <"$root/scripts/install.sh" >/dev/null
"$work/pinned/koru" version | grep -q "koru $version" && check "installs a pinned version when piped"

if sh "$root/scripts/install.sh" --version v0.0.0-missing --dir "$work/none" >/dev/null 2>"$work/err"; then exit 1; fi
grep -q "download failed" "$work/err" && check "fails on a missing release"

for archive in "$work/mirror/download/$tag"/*.tar.gz; do
  echo "0000000000000000000000000000000000000000000000000000000000000000  ${archive##*/}" >"$archive.sha256"
done
if sh "$root/scripts/install.sh" --dir "$work/bad" >/dev/null 2>"$work/err"; then exit 1; fi
grep -q "checksum mismatch" "$work/err" && [ ! -e "$work/bad/koru" ] && check "refuses a checksum mismatch"

if sh "$root/scripts/install.sh" --bogus 2>/dev/null; then exit 1; fi
check "rejects unknown options"
