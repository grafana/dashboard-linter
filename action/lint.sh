#!/usr/bin/env bash
# Lint the dashboards named by INPUT_FILES. The binary comes from the install
# step (DASHBOARD_LINTER_BIN) or from PATH.
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib.sh
source "${script_dir}/lib.sh"

main() {
  local bin=""
  local files_input=""
  local files_list=""
  local file=""
  local failed=0
  local count=0

  bin="${DASHBOARD_LINTER_BIN:-dashboard-linter}"
  files_input="${INPUT_FILES:-}"
  if [ -z "$(trim "$files_input")" ]; then
    echo "no dashboard files were given" >&2
    exit 1
  fi
  # The lint step always passes a path. --stdin would ignore that path and
  # read the step's stdin instead, once, for every file.
  case " $(trim "${INPUT_ARGS:-}") " in
    *" --stdin "*)
      echo "do not pass --stdin when files is set; omit files and run: dashboard-linter lint --stdin" >&2
      exit 1
      ;;
  esac
  if [ ! -x "$bin" ] && ! command -v "$bin" >/dev/null 2>&1; then
    echo "dashboard-linter is not executable: ${bin}" >&2
    exit 1
  fi

  files_list="$(collect_files "$files_input")" || exit 1
  if [ -z "$files_list" ]; then
    echo "no dashboard JSON files matched" >&2
    exit 1
  fi

  # printf does not re-expand paths, so a filename cannot become a command.
  while IFS= read -r file || [ -n "$file" ]; do
    [ -n "$file" ] || continue
    count=$((count + 1))
    build_lint_argv "$bin" "${INPUT_ARGS:-}" "$file"
    echo "Linting ${file}"
    if ! "${LINT_ARGV[@]}"; then
      failed=$((failed + 1))
    fi
  done < <(printf '%s\n' "$files_list")

  if [ "$count" -eq 0 ]; then
    echo "no dashboard JSON files matched" >&2
    exit 1
  fi
  if [ "$failed" -ne 0 ]; then
    echo "${failed} of ${count} dashboards failed lint" >&2
    exit 1
  fi
  echo "Linted ${count} dashboards"
}

main "$@"
