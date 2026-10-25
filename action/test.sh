#!/usr/bin/env bash
# Offline tests for the composite action. The download path is exercised by
# .github/workflows/action.yml against a published release.
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "${script_dir}/.." && pwd)"
# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib.sh
source "${script_dir}/lib.sh"

cd "$root"

assert_eq() {
  local got="$1"
  local want="$2"
  local label="$3"
  if [ "$got" != "$want" ]; then
    printf 'FAIL %s\n got: [%s]\n want: [%s]\n' "$label" "$got" "$want" >&2
    exit 1
  fi
  printf 'ok %s\n' "$label"
}

assert_fail() {
  local label="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    printf 'FAIL %s: command succeeded\n' "$label" >&2
    exit 1
  fi
  printf 'ok %s\n' "$label"
}

assert_eq "$(trim '  v0.4.0  ')" "v0.4.0" "trim spaces"
assert_eq "$(trim $'v0.4.0\r')" "v0.4.0" "trim cr"

assert_eq "$(resolve_version "" "v0.4.0")" "v0.4.0" "version from tag ref"
assert_eq "$(resolve_version "" "refs/tags/v0.4.0")" "v0.4.0" "version from refs/tags"
assert_eq "$(resolve_version "v0.3.0" "v0.4.0")" "v0.3.0" "explicit version wins"
assert_eq "$(resolve_version "  v1.2.3-rc.1  " "")" "v1.2.3-rc.1" "prerelease version"
assert_fail "reject branch ref" resolve_version "" "main"
assert_fail "reject latest" resolve_version "latest" "v0.4.0"
assert_fail "reject path traversal version" resolve_version "v0.4.0-.." ""
assert_fail "reject newline version" resolve_version $'v0.4.0\nrm' ""
assert_fail "reject slash version" resolve_version "v0.4.0/evil" ""

release_asset "v0.4.0" "Linux" "X64"
assert_eq "$RELEASE_ASSET" "dashboard-linter_0.4.0_linux_amd64.tar.gz" "linux amd64 asset"
assert_eq "$RELEASE_BINARY" "dashboard-linter" "linux binary name"
release_asset "v0.4.0" "macOS" "ARM64"
assert_eq "$RELEASE_ASSET" "dashboard-linter_0.4.0_darwin_arm64.tar.gz" "darwin arm64 asset"
release_asset "v1.2.3-rc.1" "Linux" "ARM64"
assert_eq "$RELEASE_ASSET" "dashboard-linter_1.2.3-rc.1_linux_arm64.tar.gz" "prerelease asset"
release_asset "v0.4.0" "Windows" "X64"
assert_eq "$RELEASE_ASSET" "dashboard-linter_0.4.0_windows_amd64.zip" "windows asset"
assert_eq "$RELEASE_BINARY" "dashboard-linter.exe" "windows binary name"
assert_fail "reject windows arm64" release_asset "v0.4.0" "Windows" "ARM64"
assert_fail "reject unknown os" release_asset "v0.4.0" "FreeBSD" "X64"

if ! grep -F -q '{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}' "$root/.goreleaser.yaml"; then
  echo "FAIL goreleaser archive name template changed; update release_asset" >&2
  exit 1
fi
printf 'ok %s\n' "goreleaser name template"

assert_eq \
  "$(release_download_url "grafana/dashboard-linter" "v0.4.0" "checksums.txt")" \
  "https://github.com/grafana/dashboard-linter/releases/download/v0.4.0/checksums.txt" \
  "checksum url"
assert_eq \
  "$(release_download_url "grafana/dashboard-linter" "v0.4.0" "dashboard-linter_0.4.0_linux_amd64.tar.gz")" \
  "https://github.com/grafana/dashboard-linter/releases/download/v0.4.0/dashboard-linter_0.4.0_linux_amd64.tar.gz" \
  "archive url"
assert_fail "reject repo traversal" release_download_url "grafana/.." "v0.4.0" "checksums.txt"
assert_fail "reject asset traversal" release_download_url "grafana/dashboard-linter" "v0.4.0" "../checksums.txt"
assert_fail "reject extra repo slash" action_repository_value "grafana/dashboard-linter/evil"

got_os="$(unset RUNNER_OS; detect_runner_os)"
case "$(uname -s)" in
  Darwin) want_os="macOS" ;;
  Linux) want_os="Linux" ;;
  *) want_os="" ;;
esac
if [ -n "$want_os" ]; then
  assert_eq "$got_os" "$want_os" "detect os"
fi
assert_eq "$(export RUNNER_OS=Windows; detect_runner_os)" "Windows" "runner os override"

got_arch="$(unset RUNNER_ARCH; detect_runner_arch)"
case "$(uname -m)" in
  x86_64 | amd64) want_arch="X64" ;;
  aarch64 | arm64) want_arch="ARM64" ;;
  *) want_arch="" ;;
esac
if [ -n "$want_arch" ]; then
  assert_eq "$got_arch" "$want_arch" "detect arch"
fi

tmp="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp"
}
trap cleanup EXIT

printf 'hello' >"$tmp/hello"
assert_eq "$(file_sha256 "$tmp/hello")" "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" "sha256 hello"
verify_sha256 "$tmp/hello" "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
printf 'ok %s\n' "verify sha256"
assert_fail "reject checksum mismatch" verify_sha256 "$tmp/hello" "0000000000000000000000000000000000000000000000000000000000000000"

hash_b="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
hash_a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
printf '%s\n' \
  "${hash_a}  other.tar.gz" \
  "${hash_b}  dashboard-linter_0.4.0_linux_amd64.tar.gz" \
  >"$tmp/checksums.txt"
assert_eq "$(checksum_for "$tmp/checksums.txt" "dashboard-linter_0.4.0_linux_amd64.tar.gz")" "$hash_b" "checksum entry"
printf '%s\r\n' "${hash_b}  dashboard-linter_0.4.0_linux_amd64.tar.gz" >"$tmp/checksums-crlf.txt"
assert_eq "$(checksum_for "$tmp/checksums-crlf.txt" "dashboard-linter_0.4.0_linux_amd64.tar.gz")" "$hash_b" "checksum crlf"
printf '%s\n' "${hash_b}  dashboard-linter_0.4.0_linux_amd64.tar.gz" "${hash_a}  dashboard-linter_0.4.0_linux_amd64.tar.gz" >"$tmp/checksums-dup.txt"
assert_fail "reject duplicate checksum" checksum_for "$tmp/checksums-dup.txt" "dashboard-linter_0.4.0_linux_amd64.tar.gz"
assert_fail "reject missing checksum" checksum_for "$tmp/checksums.txt" "missing.tar.gz"

printf 'payload' >"$tmp/dashboard-linter"
tar -czf "$tmp/a.tar.gz" -C "$tmp" dashboard-linter
extract_binary "$tmp/a.tar.gz" "dashboard-linter" "$tmp/extracted"
assert_eq "$(cat "$tmp/extracted")" "payload" "extract tar member"
[ -x "$tmp/extracted" ]
printf 'ok %s\n' "extracted binary is executable"
assert_fail "reject missing tar member" extract_binary "$tmp/a.tar.gz" "missing" "$tmp/missing-out"

if command -v zip >/dev/null 2>&1; then
  printf 'zip-payload' >"$tmp/dashboard-linter.exe"
  (cd "$tmp" && zip -q "$tmp/a.zip" dashboard-linter.exe)
  extract_binary "$tmp/a.zip" "dashboard-linter.exe" "$tmp/extracted.exe"
  assert_eq "$(cat "$tmp/extracted.exe")" "zip-payload" "extract zip member"
fi

mkdir -p "$tmp/dash/node_modules" "$tmp/dir with space"
printf '{}\n' >"$tmp/dash/ok.json"
printf '{}\n' >"$tmp/dash/fail.json"
printf '{}\n' >"$tmp/dash/node_modules/skip.json"
printf '{}\n' >"$tmp/dir with space/a.json"
assert_eq "$(collect_files "$tmp/dash")" "$(printf '%s\n' "$tmp/dash/fail.json" "$tmp/dash/ok.json")" "directory walk skips node_modules"
assert_eq "$(collect_files "$tmp/dir with space")" "$tmp/dir with space/a.json" "path with spaces"
# Globs are expanded relative to the working directory.
mkdir -p "$tmp/glob"
printf '{}\n' >"$tmp/glob/one.json"
assert_eq "$(collect_files "$tmp/glob/*.json")" "$tmp/glob/one.json" "shell glob"
assert_eq "$(collect_files $'lint/testdata/dashboard.json\r\n\n')" "lint/testdata/dashboard.json" "crlf file list"
assert_fail "reject recursive glob" collect_files "lint/**/*.json"
assert_fail "reject missing file" collect_files "$tmp/nope.json"
empty="$(mktemp -d "$tmp/empty.XXXXXX")"
assert_eq "$(collect_files "$empty")" "" "empty directory"

build_lint_argv "dashboard-linter" "--strict   --config .lint" "dash.json"
assert_eq "${#LINT_ARGV[@]}" "6" "argv length"
assert_eq "${LINT_ARGV[0]}" "dashboard-linter" "argv bin"
assert_eq "${LINT_ARGV[1]}" "lint" "argv command"
assert_eq "${LINT_ARGV[2]}" "--strict" "argv strict"
assert_eq "${LINT_ARGV[3]}" "--config" "argv config flag"
assert_eq "${LINT_ARGV[4]}" ".lint" "argv config path"
assert_eq "${LINT_ARGV[5]}" "dash.json" "argv file"
build_lint_argv "dashboard-linter" "*" "dash.json"
assert_eq "${LINT_ARGV[2]}" "*" "argv does not glob"
build_lint_argv "dashboard-linter" "" "dash.json"
assert_eq "${#LINT_ARGV[@]}" "3" "argv without extra args"

out="$(mktemp "$tmp/output.XXXXXX")"
GITHUB_OUTPUT="$out" write_outputs "v0.4.0" "/opt/dashboard-linter"
assert_eq "$(cat "$out")" "$(printf '%s\n' "version=v0.4.0" "bin=/opt/dashboard-linter")" "action outputs"
path_file="$(mktemp "$tmp/path.XXXXXX")"
GITHUB_PATH="$path_file" append_github_path "/opt/bin" "Linux"
assert_eq "$(cat "$path_file")" "/opt/bin" "github path"
unset GITHUB_OUTPUT GITHUB_PATH

cat >"$tmp/dashboard-linter" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >> "$FAKE_LOG"
last=""
for last in "$@"; do
  :
done
case "$last" in
  */fail.json) exit 1 ;;
esac
exit 0
EOF
chmod +x "$tmp/dashboard-linter"

export FAKE_LOG="$tmp/fake.log"
: >"$FAKE_LOG"
set +e
DASHBOARD_LINTER_BIN="$tmp/dashboard-linter" \
  INPUT_FILES="$tmp/dash" \
  INPUT_ARGS="--strict" \
  bash "$root/action/lint.sh" >"$tmp/lint-out" 2>"$tmp/lint-err"
status=$?
set -e
[ "$status" -eq 1 ] || {
  echo "FAIL lint should fail when one dashboard fails (status ${status})" >&2
  cat "$tmp/lint-err" >&2
  exit 1
}
grep -F -q -- "lint --strict ${tmp}/dash/fail.json" "$FAKE_LOG" || {
  echo "FAIL lint did not receive --strict for fail.json" >&2
  cat "$FAKE_LOG" >&2
  exit 1
}
grep -F -q -- "lint --strict ${tmp}/dash/ok.json" "$FAKE_LOG" || {
  echo "FAIL lint stopped after the first failure" >&2
  cat "$FAKE_LOG" >&2
  exit 1
}
if grep -F -q "node_modules" "$FAKE_LOG"; then
  echo "FAIL linted a file under node_modules" >&2
  exit 1
fi
printf 'ok %s\n' "lint all files and fail"

: >"$FAKE_LOG"
DASHBOARD_LINTER_BIN="$tmp/dashboard-linter" \
  INPUT_FILES="$tmp/dash/ok.json" \
  INPUT_ARGS="" \
  bash "$root/action/lint.sh" >"$tmp/lint-ok" 2>"$tmp/lint-ok-err"
printf 'ok %s\n' "lint success"

mkdir -p "$tmp/weird"
rm -f "$root/pwned-by-lint"
# A filename that is a command substitution must stay literal. It cannot
# contain slashes, so the sentinel is created in the repo root if expanded.
# shellcheck disable=SC2016
injection='$(touch pwned-by-lint).json'
printf '{}\n' >"$tmp/weird/${injection}"
: >"$FAKE_LOG"
DASHBOARD_LINTER_BIN="$tmp/dashboard-linter" \
  INPUT_FILES="$tmp/weird/${injection}" \
  INPUT_ARGS="" \
  bash "$root/action/lint.sh" >"$tmp/lint-inject" 2>"$tmp/lint-inject-err"
if [ -e "$root/pwned-by-lint" ]; then
  rm -f "$root/pwned-by-lint"
  echo "FAIL lint expanded a command substitution in a filename" >&2
  exit 1
fi
# shellcheck disable=SC2016
grep -F -q -- '$(touch pwned-by-lint)' "$FAKE_LOG" || {
  echo "FAIL lint did not pass the filename through unchanged" >&2
  cat "$FAKE_LOG" >&2
  exit 1
}
printf 'ok %s\n' "filename is not evaluated"

set +e
INPUT_FILES="$tmp/nope.json" \
  DASHBOARD_LINTER_BIN="$tmp/dashboard-linter" \
  bash "$root/action/lint.sh" >"$tmp/lint-miss" 2>"$tmp/lint-miss-err"
status=$?
set -e
[ "$status" -eq 1 ] || {
  echo "FAIL missing dashboard should fail the lint step" >&2
  exit 1
}
printf 'ok %s\n' "lint missing file"

set +e
INPUT_FILES="$tmp/dash/ok.json" \
  INPUT_ARGS="--strict --stdin" \
  DASHBOARD_LINTER_BIN="$tmp/dashboard-linter" \
  bash "$root/action/lint.sh" >"$tmp/lint-stdin" 2>"$tmp/lint-stdin-err"
status=$?
set -e
[ "$status" -eq 1 ] || {
  echo "FAIL --stdin should be rejected when files is set" >&2
  exit 1
}
grep -F -q -- "--stdin" "$tmp/lint-stdin-err" || {
  echo "FAIL --stdin rejection did not explain itself" >&2
  cat "$tmp/lint-stdin-err" >&2
  exit 1
}
printf 'ok %s\n' "reject --stdin with files"

if (
  GITHUB_ACTIONS=true
  unset GITHUB_ACTION_REPOSITORY
  action_repository
) >/dev/null 2>&1; then
  echo "FAIL Actions should require GITHUB_ACTION_REPOSITORY" >&2
  exit 1
fi
assert_eq "$(unset GITHUB_ACTIONS; unset GITHUB_ACTION_REPOSITORY; action_repository)" "grafana/dashboard-linter" "default repository outside Actions"

mkdir -p "$tmp/bin"
cat >"$tmp/bin/curl" <<'EOF'
#!/bin/sh
printf 'CALL %s\n' "$*" >> "$CURL_LOG"
has_auth=0
case " $* " in
  *"Authorization:"*) has_auth=1 ;;
esac
if [ "${MOCK_CURL_REJECT_ANON:-}" = "1" ] && [ "$has_auth" -eq 0 ]; then
  exit 22
fi
output=""
previous=""
for arg in "$@"; do
  if [ "$previous" = "--output" ]; then
    output="$arg"
  fi
  previous="$arg"
done
if [ -z "$output" ]; then
  echo "mock curl: missing --output" >&2
  exit 2
fi
printf 'payload' >"$output"
exit 0
EOF
chmod +x "$tmp/bin/curl"
CURL_LOG="$tmp/curl.log"
: >"$CURL_LOG"
(
  # shellcheck disable=SC2030
  PATH="$tmp/bin:${PATH}"
  export PATH CURL_LOG
  GITHUB_TOKEN="example-token"
  curl_fetch "https://example.com/dashboard-linter.tar.gz" "$tmp/downloaded"
)
if grep -q 'Authorization:' "$CURL_LOG"; then
  echo "FAIL anonymous success still sent credentials" >&2
  cat "$CURL_LOG" >&2
  exit 1
fi
[ "$(grep -c '^CALL ' "$CURL_LOG")" -eq 1 ] || {
  echo "FAIL anonymous success should call curl once" >&2
  cat "$CURL_LOG" >&2
  exit 1
}
assert_eq "$(cat "$tmp/downloaded")" "payload" "anonymous download ignores token"

: >"$CURL_LOG"
(
  # shellcheck disable=SC2031
  PATH="$tmp/bin:${PATH}"
  export PATH CURL_LOG MOCK_CURL_REJECT_ANON
  MOCK_CURL_REJECT_ANON=1
  GITHUB_TOKEN="example-token"
  curl_fetch "https://example.com/dashboard-linter.tar.gz" "$tmp/downloaded-auth"
)
auth_calls="$(grep -c 'Authorization:' "$CURL_LOG" || true)"
[ "$auth_calls" -eq 1 ] || {
  echo "FAIL expected one authenticated retry, got ${auth_calls}" >&2
  cat "$CURL_LOG" >&2
  exit 1
}
[ "$(grep -c '^CALL ' "$CURL_LOG")" -eq 2 ] || {
  echo "FAIL expected anonymous attempt then authenticated retry" >&2
  cat "$CURL_LOG" >&2
  exit 1
}
if grep -F -q "example-token" "$tmp/downloaded-auth"; then
  echo "FAIL download wrote the token into the file" >&2
  exit 1
fi
assert_eq "$(cat "$tmp/downloaded-auth")" "payload" "authenticated retry after anonymous failure"

set +e
INPUT_VERSION="latest" bash "$root/action/install.sh" >"$tmp/install-out" 2>"$tmp/install-err"
status=$?
set -e
[ "$status" -ne 0 ] || {
  echo "FAIL install should reject version=latest" >&2
  exit 1
}
grep -F -q "invalid version" "$tmp/install-err" || {
  echo "FAIL install did not report an invalid version" >&2
  cat "$tmp/install-err" >&2
  exit 1
}
printf 'ok %s\n' "install rejects latest"

printf 'ok %s\n' "all action tests passed"
