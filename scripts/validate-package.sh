#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

make package-plugin

ZIP="dist/macromog-$(awk -F"'" '/_addon.version/{print $2; exit}' macromog.lua).zip"
if [[ ! -f "$ZIP" ]]; then
  echo "FAIL: missing $ZIP" >&2
  exit 1
fi

list="$(unzip -l "$ZIP")"

require_entry() {
  local needle="$1"
  if ! grep -Fq "$needle" <<<"$list"; then
    echo "FAIL: zip missing $needle" >&2
    exit 1
  fi
}

forbid_entry() {
  local needle="$1"
  if grep -Fq "$needle" <<<"$list"; then
    echo "FAIL: zip must not contain $needle" >&2
    exit 1
  fi
}

require_entry 'Macromog/macromog.lua'
require_entry 'Macromog/lib/cli.lua'
require_entry 'Macromog/bin/macromog.exe'
forbid_entry 'macromog-linux'
forbid_entry 'macromog-windows'
require_entry 'Macromog/data/'

forbid_entry '.gitkeep'
forbid_entry 'example_macros.yml'

echo "PASS: package layout for $ZIP"
