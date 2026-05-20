#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
cd "$repo_root"

mode="check-build"

usage() {
  cat >&2 <<'USAGE'
usage: scripts/harness/alpha_shell_smoke.sh [--check-build|--launch]

--check-build  Verify build artifacts and Alpha shell binding boundaries.
--launch       Also launch the macOS app bundle and confirm the process starts.
USAGE
}

while (($#)); do
  case "$1" in
    --check-build)
      mode="check-build"
      shift
      ;;
    --launch)
      mode="launch"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    echo "missing required file: $path" >&2
    exit 1
  fi
}

require_dir() {
  local path="$1"
  if [[ ! -d "$path" ]]; then
    echo "missing required directory: $path" >&2
    exit 1
  fi
}

require_pattern() {
  local pattern="$1"
  local path="$2"
  if ! rg -q "$pattern" "$path"; then
    echo "missing expected pattern '$pattern' in $path" >&2
    exit 1
  fi
}

collect_app_pids() {
  pgrep -x "tuyu-studio" || true
}

pid_in_text() {
  local needle="$1"

  printf '%s\n' "$2" | grep -Fxq "$needle"
}

require_file "build/bin/tuyu-studio.app/Contents/MacOS/tuyu-studio"
require_file "frontend/wailsjs/go/main/App.js"
require_file "frontend/wailsjs/go/main/App.d.ts"
require_file "frontend/wailsjs/go/models.ts"
require_file "frontend/package.json"
require_dir "frontend/src/api"

require_pattern "AppInfo" "frontend/wailsjs/go/main/App.js"
require_pattern "ShellHealth" "frontend/wailsjs/go/main/App.js"
require_pattern "WorkbenchProbe" "frontend/wailsjs/go/main/App.js"
require_pattern "AppError" "frontend/wailsjs/go/models.ts"
require_pattern "RuntimeEvent" "frontend/wailsjs/go/models.ts"
require_pattern "WorkbenchProbeResult" "frontend/wailsjs/go/models.ts"
require_pattern '"typecheck"' "frontend/package.json"
require_pattern '"test:unit"' "frontend/package.json"
require_pattern '"build"' "frontend/package.json"

if rg -n "wailsjs/go|wailsjs/runtime" frontend/src | grep -v '^frontend/src/api/' >&2; then
  echo "generated Wails binding import escaped frontend/src/api boundary" >&2
  exit 1
fi

if rg -n "<ul|<li" frontend/src >&2; then
  echo "raw list markup found in frontend/src; use Ant Design Vue list primitives for this shell" >&2
  exit 1
fi

if [[ "$mode" == "launch" ]]; then
  if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "--launch is only supported on macOS" >&2
    exit 1
  fi

  before_pids="$(collect_app_pids)"

  open -n "build/bin/tuyu-studio.app"

  started_pid=""
  started=0
  for _ in {1..30}; do
    after_pids="$(collect_app_pids)"
    for pid in $after_pids; do
      if ! pid_in_text "$pid" "$before_pids"; then
        started_pid="$pid"
        started=1
        break
      fi
    done

    if [[ "$started" -eq 1 ]]; then
      break
    fi

    sleep 1
  done

  if [[ "$started" -ne 1 ]]; then
    echo "new tuyu-studio process did not start after launching the app bundle" >&2
    exit 1
  fi

  if [[ "${TUYU_SMOKE_KEEP_APP:-0}" != "1" ]]; then
    kill "$started_pid" || true
  fi
fi

echo "alpha shell smoke passed (${mode})"
