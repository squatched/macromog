#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

BIN="$ROOT/dist/bin/macromog-windows-386.exe"
if [[ ! -f "$BIN" ]]; then
  echo "FAIL: missing $BIN (run make build-release-bins first)" >&2
  exit 1
fi

# Native Windows (Git Bash / MSYS2): run the .exe directly.
if [[ "${OS:-}" == "Windows_NT" ]]; then
  "$BIN" --help >/dev/null
  printf 'PASS: native Windows smoke (%s)\n' "$BIN"
  exit 0
fi

# WSL2: Windows interop executes .exe files as real Windows processes — no Wine needed.
if [[ -n "${WSL_DISTRO_NAME:-}" ]]; then
  "$BIN" --help >/dev/null
  printf 'PASS: WSL2 native Windows smoke (%s)\n' "$BIN"
  exit 0
fi

# Linux: require Wine.
if ! command -v wine >/dev/null 2>&1; then
  printf 'SKIP: wine not installed; on Windows run dist\\bin\\macromog.exe natively\n'
  exit 0
fi

TMP="$(mktemp -d)"
trap 'wineserver -k 2>/dev/null || true; rm -rf "$TMP"' EXIT

export WINEPREFIX="$TMP/wine"
export WINEDEBUG=-all
# Go binaries are native PE, not .NET — skip Wine's mono installer prompt.
export WINEDLLOVERRIDES='mscoree,mshtml=d'

wineboot --init >/dev/null 2>&1
wine "$BIN" --help >/dev/null

printf 'PASS: wine smoke (%s)\n' "$BIN"
