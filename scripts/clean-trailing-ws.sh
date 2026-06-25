#!/usr/bin/env bash
# clean-trailing-ws.sh — remove trailing spaces/tabs from text files
#
# Strips trailing whitespace from lines in tracked and untracked text files
# (respecting .gitignore) and ensures
# a single trailing newline at EOF (no trailing blank lines).
#
# Usage:
#   scripts/clean-trailing-ws.sh [--staged | --changed] [--check] [--dry-run] [--verbose|-v]
#
# Options:
#   --staged    Only files staged for commit (A/M/C/R).
#   --changed   Only modified unstaged files.
#   --check     Validate only; exit 1 if issues found (implies --dry-run).
#   --dry-run   Report changes without editing.
#   --verbose   Print per-file progress.
#
# Exit codes: 0 ok | 1 validation failed or runtime error | 2 bad args
set -euo pipefail

MODE="all"
CHECK=0
DRY_RUN=0
VERBOSE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --staged) MODE="staged" ;;
    --changed) MODE="changed" ;;
    --check) CHECK=1; DRY_RUN=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --verbose|-v) VERBOSE=1 ;;
    --help|-h)
      sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *) echo "Unknown arg: $1" >&2; exit 2 ;;
  esac
  shift
done

log() { [[ $VERBOSE -eq 1 ]] && echo "$@" || true; }

declare -a FILES=()

case "$MODE" in
  all)
    while IFS= read -r -d '' f; do FILES+=("$f"); done < <(
      { git ls-files -z; git ls-files -z --others --exclude-standard; } | sort -zu
    )
    ;;
  staged)
    while IFS= read -r -d '' f; do FILES+=("$f"); done < <(git diff --cached --name-only -z --diff-filter=ACMR)
    ;;
  changed)
    while IFS= read -r -d '' f; do FILES+=("$f"); done < <(git diff --name-only -z --diff-filter=ACMR)
    ;;
esac

is_text_file() {
  local f="$1"
  [[ ! -s "$f" ]] && return 0
  grep -Iq . -- "$f"
}

declare -a TARGETS=()
for f in "${FILES[@]:-}"; do
  [[ -f "$f" ]] || continue
  [[ -L "$f" ]] && continue
  if is_text_file "$f"; then
    TARGETS+=("$f")
  fi
done

[[ ${#TARGETS[@]} -eq 0 ]] && exit 0

CHANGED=0

if [[ $DRY_RUN -eq 1 ]]; then
  declare -a ISSUES=()

  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    ISSUES+=("$f")
    log "would fix trailing whitespace: $f"
  done < <(printf '%s\0' "${TARGETS[@]}" | xargs -0 grep -l '[[:blank:]]$' 2>/dev/null || true)

  for f in "${TARGETS[@]}"; do
    [[ ! -s "$f" ]] && continue

    already=0
    for i in "${ISSUES[@]:-}"; do
      [[ "$i" == "$f" ]] && already=1 && break
    done
    [[ $already -eq 1 ]] && continue

    last2hex=$(tail -c2 "$f" | od -An -tx1 | tr -d ' \n')

    if [[ ! "$last2hex" =~ 0a$ ]]; then
      ISSUES+=("$f")
      log "would fix missing EOF newline: $f"
      continue
    fi

    if [[ "$last2hex" == "0a0a" ]]; then
      ISSUES+=("$f")
      log "would fix trailing blank lines: $f"
    fi
  done

  CHANGED=${#ISSUES[@]}
else
  for f in "${TARGETS[@]}"; do
    before_hash=$(sha256sum "$f" 2>/dev/null | awk '{print $1}')
    perl -0777 -pe 's/[ \t]+$//mg' -i -- "$f"
    perl -0777 -pe 's/\s*\z/\n/s if /\S/' -i -- "$f"
    after_hash=$(sha256sum "$f" 2>/dev/null | awk '{print $1}')
    if [[ "$before_hash" != "$after_hash" ]]; then
      ((CHANGED++)) || true
      log "fixed: $f"
    fi
  done
fi

if [[ $CHECK -eq 1 ]]; then
  if [[ $CHANGED -gt 0 ]]; then
    printf 'FAIL: trailing whitespace or EOF issues in %d file(s)\n' "$CHANGED" >&2
    exit 1
  fi
  printf 'PASS: no trailing whitespace or EOF issues\n'
  exit 0
fi

log "clean-trailing-ws: ${CHANGED} file(s) changed"
exit 0
