#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
cd "$repo_root"

status=0

if ! git diff --check; then
  status=1
fi

if ! git diff --cached --check; then
  status=1
fi

is_text_file() {
  local path="$1"

  if [[ ! -s "$path" ]]; then
    return 0
  fi

  grep -Iq . "$path"
}

check_untracked_file() {
  local path="$1"

  [[ -f "$path" ]] || return 0
  is_text_file "$path" || return 0

  if ! perl -ne '
    if (/[ \t]+$/) {
      print "$ARGV:$.: trailing whitespace\n";
      $bad = 1;
    }
    END {
      exit($bad ? 1 : 0);
    }
  ' "$path"; then
    status=1
  fi

  if [[ -s "$path" ]] && [[ "$(tail -c 1 "$path" | wc -l | tr -d ' ')" == "0" ]]; then
    echo "$path: missing newline at EOF" >&2
    status=1
  fi
}

while IFS= read -r -d '' path; do
  check_untracked_file "$path"
done < <(git ls-files --others --exclude-standard -z)

if [[ "$status" -ne 0 ]]; then
  exit "$status"
fi

echo "worktree whitespace check passed"
