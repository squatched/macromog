#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v wine >/dev/null 2>&1; then
  printf 'SKIP: wine not installed (optional contributor check)\n'
  exit 0
fi

BIN="$ROOT/dist/bin/macromog-windows-amd64.exe"
if [[ ! -f "$BIN" ]]; then
  echo "FAIL: missing $BIN (run make build-release-bins first)" >&2
  exit 1
fi

TMP="$(mktemp -d)"
trap 'wineserver -k 2>/dev/null || true; rm -rf "$TMP"' EXIT

export WINEPREFIX="$TMP/wine"
export WINEARCH=win64
export WINEDEBUG=-all
# Go binaries are native PE, not .NET — skip Wine's mono installer prompt.
export WINEDLLOVERRIDES='mscoree,mshtml=d'

wineboot --init >/dev/null 2>&1
wine "$BIN" --help >/dev/null

printf 'PASS: wine smoke (%s)\n' "$BIN"
