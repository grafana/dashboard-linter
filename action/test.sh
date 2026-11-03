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
  unset ACTION_REPOSITORY GITHUB_ACTION_REPOSITORY
  unset GITHUB_ACTION_PATH
  action_repository
) >/dev/null 2>&1; then
  echo "FAIL Actions should not guess the action repository" >&2
  exit 1
fi
if (
  GITHUB_ACTIONS=true
  ACTION_REPOSITORY=
  GITHUB_ACTION_REPOSITORY=
  unset GITHUB_ACTION_PATH
  action_repository
) >/dev/null 2>&1; then
  echo "FAIL an empty action repository should not be guessed" >&2
  exit 1
fi
assert_eq "$(
  GITHUB_ACTIONS=true
  unset ACTION_REPOSITORY
  GITHUB_ACTION_REPOSITORY=
  unset GITHUB_ACTION_PATH
  export GITHUB_ACTION_REPOSITORY=grafana/dashboard-linter
  action_repository
)" "grafana/dashboard-linter" "repository assigned inside the step"
assert_eq "$(
  GITHUB_ACTIONS=true
  unset ACTION_REPOSITORY
  GITHUB_ACTION_REPOSITORY=actions/checkout
  GITHUB_ACTION_PATH=/home/runner/work/_actions/grafana/dashboard-linter/8ce5fef
  action_repository
)" "grafana/dashboard-linter" "action path wins over a stale GITHUB_ACTION_REPOSITORY"
assert_eq "$(
  GITHUB_ACTIONS=true
  unset GITHUB_ACTION_REPOSITORY
  GITHUB_ACTION_PATH=/home/runner/work/_actions/grafana/dashboard-linter/8ce5fef
  action_repository
)" "grafana/dashboard-linter" "repository from action path"
assert_eq "$(
  repository_from_action_path 'D:\a\_actions\grafana\dashboard-linter\2580d238c529bed3a2c1773ac98cba296813e977'
)" "grafana/dashboard-linter" "repository from a windows action path"
assert_eq "$(
  ref_from_action_path 'D:\a\_actions\grafana\dashboard-linter\2580d238c529bed3a2c1773ac98cba296813e977'
)" "2580d238c529bed3a2c1773ac98cba296813e977" "ref from a windows action path"
assert_eq "$(
  ref_from_action_path /home/runner/work/_actions/grafana/dashboard-linter/feature/foo
)" "feature/foo" "branch ref with a slash"
assert_fail "ref requires an actions directory" ref_from_action_path /opt/grafana/dashboard-linter/v0.4.0
assert_eq "$(
  GITHUB_ACTIONS=true
  ACTION_REPOSITORY=grafana/dashboard-linter
  GITHUB_ACTION_REPOSITORY=actions/checkout
  GITHUB_ACTION_PATH=/home/runner/work/_actions/actions/checkout/v4
  action_repository
)" "grafana/dashboard-linter" "ACTION_REPOSITORY wins over the action path"
assert_eq "$(
  unset GITHUB_ACTIONS ACTION_REPOSITORY GITHUB_ACTION_REPOSITORY GITHUB_ACTION_PATH
  action_repository
)" "grafana/dashboard-linter" "default repository outside Actions"

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
  # shellcheck disable=SC2030,SC2031
  PATH="$tmp/bin:${PATH}"
  export PATH CURL_LOG MOCK_CURL_REJECT_ANON
  MOCK_CURL_REJECT_ANON=1
  # shellcheck disable=SC2030
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

# action.yml must not assign the names the runner clears. The log still shows
# the step env, but the process receives them empty (actions/runner#2525).
if grep -E '^[[:space:]]+GITHUB_ACTION_REF:' "$root/action.yml" >/dev/null; then
  echo "FAIL action.yml assigns GITHUB_ACTION_REF; the runner clears that name" >&2
  exit 1
fi
if grep -E '^[[:space:]]+GITHUB_ACTION_REPOSITORY:' "$root/action.yml" >/dev/null; then
  echo "FAIL action.yml assigns GITHUB_ACTION_REPOSITORY; the runner clears that name" >&2
  exit 1
fi
# shellcheck disable=SC2016
grep -F -q 'ACTION_REF: ${{ github.action_ref }}' "$root/action.yml" || {
  echo "FAIL action.yml must copy github.action_ref to ACTION_REF" >&2
  exit 1
}
# shellcheck disable=SC2016
grep -F -q 'ACTION_REPOSITORY: ${{ github.action_repository }}' "$root/action.yml" || {
  echo "FAIL action.yml must copy github.action_repository to ACTION_REPOSITORY" >&2
  exit 1
}
if grep -F 'GITHUB_ACTION_REF:-' "$root/action/install.sh" >/dev/null; then
  echo "FAIL install.sh reads GITHUB_ACTION_REF directly; use action_ref" >&2
  exit 1
fi
grep -F -q 'action_ref' "$root/action/install.sh" || {
  echo "FAIL install.sh must resolve the version from action_ref" >&2
  exit 1
}
printf 'ok %s\n' "action ref is not a runner-cleared name"

mkdir -p "$tmp/infer-bin"
cat >"$tmp/infer-bin/curl" <<'EOF'
#!/bin/sh
echo "offline curl refused: $*" >&2
exit 22
EOF
chmod +x "$tmp/infer-bin/curl"

# run_install_ref LABEL URL_NEEDLE then env assignments. Curl is offline.
# A resolved version reaches "Downloading <archive>" before curl fails.
run_install_ref() {
  local label="$1"
  local needle="$2"
  shift 2
  set +e
  # shellcheck disable=SC2031
  env -u GITHUB_TOKEN -u GH_TOKEN \
    GITHUB_ACTIONS=true \
    INPUT_VERSION= \
    PATH="$tmp/infer-bin:${PATH}" \
    "$@" \
    bash "$root/action/install.sh" >"$tmp/install-infer-out" 2>"$tmp/install-infer-err"
  status=$?
  set -e
  [ "$status" -ne 0 ] || {
    echo "FAIL ${label}: install should stop when the download is refused" >&2
    exit 1
  }
  if ! grep -F -q "$needle" "$tmp/install-infer-out"; then
    echo "FAIL ${label}: did not request ${needle}" >&2
    echo "--- stdout ---" >&2
    cat "$tmp/install-infer-out" >&2
    echo "--- stderr ---" >&2
    cat "$tmp/install-infer-err" >&2
    exit 1
  fi
  printf 'ok %s\n' "$label"
}

run_install_ref "install reads ACTION_REF when GITHUB_ACTION_REF is empty" \
  "releases/download/v0.4.0/dashboard-linter_" \
  GITHUB_ACTION_REF= \
  GITHUB_ACTION_REPOSITORY= \
  GITHUB_ACTION_PATH= \
  ACTION_REF=v0.4.0 \
  ACTION_REPOSITORY=grafana/dashboard-linter

run_install_ref "install reads the ref from the action directory" \
  "releases/download/v0.4.0/dashboard-linter_" \
  GITHUB_ACTION_REF= \
  GITHUB_ACTION_REPOSITORY=actions/checkout \
  ACTION_REF= \
  ACTION_REPOSITORY= \
  GITHUB_ACTION_PATH=/home/runner/work/_actions/grafana/dashboard-linter/v0.4.0

run_install_ref "install still accepts GITHUB_ACTION_REF outside the composite quirk" \
  "releases/download/v0.3.0/dashboard-linter_" \
  GITHUB_ACTION_REF=v0.3.0 \
  GITHUB_ACTION_REPOSITORY=grafana/dashboard-linter \
  GITHUB_ACTION_PATH= \
  ACTION_REF= \
  ACTION_REPOSITORY=

sha_v050="63340c76001dd04e98ec0c3e5ed9f927eb3b7b68"
tagobj_v050="a81931ee153a1e10a937ac06435a4b1e936eea2f"
other_sha="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
commit_v040="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
tags_v050="$(printf '%s\n' \
  "${tagobj_v050}"$'\t'"refs/tags/v0.5.0" \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0^{}" \
  "${commit_v040}"$'\t'"refs/tags/v0.4.0^{}" \
  "${other_sha}"$'\t'"refs/tags/v0.4.0")"
assert_eq "$(version_from_tag_lines "$sha_v050" "$tags_v050")" "v0.5.0" "release tag for pinned sha"
assert_fail "tag object is not the release commit" version_from_tag_lines "$tagobj_v050" "$tags_v050"

light="$(printf '%s\n' "${sha_v050}"$'\t'"refs/tags/v1.2.3")"
assert_eq "$(version_from_tag_lines "$sha_v050" "$light")" "v1.2.3" "lightweight tag"

mixed="$(printf '%s\n' \
  "${sha_v050}"$'\t'"refs/tags/latest" \
  "${sha_v050}"$'\t'"refs/tags/../v0.9.0" \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0;rm -rf" \
  "not-a-line" \
  "${tagobj_v050}"$'\t'"refs/tags/v0.5.0" \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0^{}")"
assert_eq "$(version_from_tag_lines "$sha_v050" "$mixed")" "v0.5.0" "ignore non-release refs"

dup="$(printf '%s\n' \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0^{}" \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0^{}" \
  "${tagobj_v050}"$'\t'"refs/tags/v0.5.0")"
assert_eq "$(version_from_tag_lines "$sha_v050" "$dup")" "v0.5.0" "duplicate tag line"

rc="$(printf '%s\n' \
  "${sha_v050}"$'\t'"refs/tags/v1.2.3-rc.1^{}" \
  "${other_sha}"$'\t'"refs/tags/v1.2.3-rc.1")"
assert_eq "$(version_from_tag_lines "$sha_v050" "$rc")" "v1.2.3-rc.1" "prerelease tag on a commit"

ambiguous="$(printf '%s\n' \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0^{}" \
  "${tagobj_v050}"$'\t'"refs/tags/v0.5.0" \
  "${sha_v050}"$'\t'"refs/tags/v0.5.0-rc.1^{}" \
  "${other_sha}"$'\t'"refs/tags/v0.5.0-rc.1")"
status=0
ambig_out="$(version_from_tag_lines "$sha_v050" "$ambiguous" 2>"$tmp/ambig-err")" || status=$?
[ "$status" -ne 0 ] || {
  echo "FAIL ambiguous tags should not resolve" >&2
  exit 1
}
[ -z "$ambig_out" ] || {
  echo "FAIL ambiguous tags printed a version" >&2
  printf '%s\n' "$ambig_out" >&2
  exit 1
}
grep -F -q "v0.5.0, v0.5.0-rc.1" "$tmp/ambig-err" || {
  echo "FAIL ambiguous error did not list both release tags" >&2
  cat "$tmp/ambig-err" >&2
  exit 1
}
grep -F -q "set version" "$tmp/ambig-err" || {
  echo "FAIL ambiguous error did not ask for version" >&2
  cat "$tmp/ambig-err" >&2
  exit 1
}
printf 'ok %s\n' "reject two release tags on one commit"

big="$(printf '%300000s' x)"
status=0
big_out="$(version_from_tag_lines "$sha_v050" "$big" 2>"$tmp/big-err")" || status=$?
[ "$status" -ne 0 ] || {
  echo "FAIL oversized tag list should not resolve" >&2
  exit 1
}
[ -z "$big_out" ] || {
  echo "FAIL oversized tag list printed a version" >&2
  exit 1
}
grep -F -q "too large" "$tmp/big-err" || {
  echo "FAIL oversized tag list error was not explained" >&2
  cat "$tmp/big-err" >&2
  exit 1
}
printf 'ok %s\n' "reject oversized tag list"

got="$(
  unset GITHUB_ACTIONS GITHUB_ACTION_REPOSITORY GITHUB_ACTION_PATH ACTION_REPOSITORY ACTION_REF
  fetch_remote_tags() { printf '%s\n' "$tags_v050"; }
  resolve_version "" "$sha_v050" 2>/dev/null
)"
assert_eq "$got" "v0.5.0" "resolve_version maps a pinned sha"

got="$(
  ACTION_REF=
  GITHUB_ACTION_REF=v2
  GITHUB_ACTION_PATH="/home/runner/work/_actions/grafana/dashboard-linter/${sha_v050}"
  action_ref
)"
assert_eq "$got" "$sha_v050" "action directory ref wins over a stale GITHUB_ACTION_REF"

: >"$tmp/fetch-count"
got="$(
  ACTION_REF="$sha_v050"
  GITHUB_ACTION_REF=v0.4.0
  GITHUB_ACTION_PATH=/home/runner/work/_actions/grafana/dashboard-linter/v0.4.0
  fetch_remote_tags() {
    echo fetched >>"$tmp/fetch-count"
    printf '%s\n' "$tags_v050"
  }
  resolve_version "" "$(action_ref)" 2>/dev/null
)"
assert_eq "$got" "v0.5.0" "ACTION_REF wins over the action directory"
if [ ! -s "$tmp/fetch-count" ]; then
  echo "FAIL ACTION_REF did not look up the pinned sha" >&2
  exit 1
fi

got="$(
  unset ACTION_REF
  GITHUB_ACTION_REF=
  GITHUB_ACTION_PATH="/home/runner/work/_actions/grafana/dashboard-linter/${sha_v050}"
  fetch_remote_tags() { printf '%s\n' "$tags_v050"; }
  resolve_version "" "$(action_ref)" 2>/dev/null
)"
assert_eq "$got" "v0.5.0" "cleared GITHUB_ACTION_REF still resolves from the action directory"

got="$(
  fetch_remote_tags() { echo FETCHED >&2; return 1; }
  resolve_version "v0.3.0" "$sha_v050"
)"
assert_eq "$got" "v0.3.0" "explicit version skips sha lookup"

lookup_err="$(
  fetch_remote_tags() { echo FETCHED >&2; return 1; }
  resolve_version "" "main" 2>&1 || true
)"
case "$lookup_err" in
  *FETCHED*)
    echo "FAIL branch ref listed tags" >&2
    printf '%s\n' "$lookup_err" >&2
    exit 1
    ;;
esac
case "$lookup_err" in
  *"version is required"*) ;;
  *)
    echo "FAIL branch ref did not require version" >&2
    printf '%s\n' "$lookup_err" >&2
    exit 1
    ;;
esac
printf 'ok %s\n' "branch ref does not list tags"

lookup_err="$(
  GITHUB_ACTIONS=true
  unset ACTION_REPOSITORY GITHUB_ACTION_REPOSITORY GITHUB_ACTION_PATH
  fetch_remote_tags() { echo FETCHED >&2; return 1; }
  resolve_version "" "$sha_v050" 2>&1 || true
)"
case "$lookup_err" in
  *FETCHED*)
    echo "FAIL sha lookup ran without a repository" >&2
    printf '%s\n' "$lookup_err" >&2
    exit 1
    ;;
esac
printf 'ok %s\n' "sha lookup requires the action repository"

mkdir -p "$tmp/bin"
cat >"$tmp/bin/git" <<'EOF'
#!/bin/sh
printf 'ARGV %s\n' "$*" >> "$GIT_LOG"
if [ -n "${GITHUB_TOKEN+x}" ]; then
  echo "TOKEN_SET" >> "$GIT_LOG"
else
  echo "TOKEN_UNSET" >> "$GIT_LOG"
fi
if [ -n "${GIT_CONFIG_COUNT+x}" ]; then
  echo "CONFIG_COUNT_SET" >> "$GIT_LOG"
else
  echo "CONFIG_COUNT_UNSET" >> "$GIT_LOG"
fi
printf 'PWD %s\n' "$PWD" >> "$GIT_LOG"
printf 'SSL [%s]\n' "${GIT_SSL_NO_VERIFY-unset}" >> "$GIT_LOG"
printf 'VERBOSE [%s]\n' "${GIT_CURL_VERBOSE-unset}" >> "$GIT_LOG"
printf 'PRELOAD [%s]\n' "${LD_PRELOAD-unset}" >> "$GIT_LOG"
printf 'BASHENV [%s]\n' "${BASH_ENV-unset}" >> "$GIT_LOG"
printf 'PROTO [%s]\n' "${GIT_ALLOW_PROTOCOL-unset}" >> "$GIT_LOG"
printf 'HOME [%s]\n' "${HOME-unset}" >> "$GIT_LOG"
case "${GIT_ASKPASS-}" in
  "" | true)
    exit 1
    ;;
esac
if grep -F -q "$EXPECTED_TOKEN" "$GIT_ASKPASS"; then
  echo "TOKEN_IN_ASKPASS" >> "$GIT_LOG"
else
  echo "ASKPASS_CLEAN" >> "$GIT_LOG"
fi
user="$("$GIT_ASKPASS" "Username for 'https://github.com':")"
pw="$("$GIT_ASKPASS" "Password:")"
printf 'USER %s\n' "$user" >> "$GIT_LOG"
if [ "$pw" = "${EXPECTED_TOKEN}" ]; then
  echo "ASKPASS_MATCH" >> "$GIT_LOG"
else
  echo "ASKPASS_MISMATCH" >> "$GIT_LOG"
fi
printf '%s\t%s\n' "63340c76001dd04e98ec0c3e5ed9f927eb3b7b68" "refs/tags/v0.5.0^{}"
exit 0
EOF
chmod +x "$tmp/bin/git"
GIT_LOG="$tmp/git.log"
: >"$GIT_LOG"
# $GIT_LOG must be expanded by BASH_ENV, not while writing this file.
# shellcheck disable=SC2016
echo 'echo BASH_ENV_RAN >> "$GIT_LOG"' >"$tmp/bash-env"
(
  # shellcheck disable=SC2030,SC2031
  PATH="$tmp/bin:${PATH}"
  export PATH GIT_LOG
  # CI sets GITHUB_ACTIONS, which refuses a git binary under TMPDIR.
  unset GITHUB_ACTIONS
  # shellcheck disable=SC2030,SC2031
  export GITHUB_TOKEN="example_token"
  export EXPECTED_TOKEN="example_token"
  export GIT_CONFIG_COUNT=1
  export GIT_CONFIG_KEY_0="url.https://example.invalid/.insteadOf"
  export GIT_CONFIG_VALUE_0="https://github.com/"
  export GIT_SSL_NO_VERIFY=1
  export GIT_SSL_CAINFO="$tmp/missing-ca.pem"
  export SSL_CERT_FILE="$tmp/missing-ca.pem"
  export GIT_CURL_VERBOSE=1
  export LD_PRELOAD="$tmp/libevil.so"
  export BASH_ENV="$tmp/bash-env"
  fetch_remote_tags "grafana/dashboard-linter" >"$tmp/git-tags" 2>"$tmp/git-err"
)
assert_eq "$(cat "$tmp/git-tags")" "$(printf '%s\t%s\n' "$sha_v050" "refs/tags/v0.5.0^{}")" "tag lookup stdout"
grep -F -q "retrying with credentials" "$tmp/git-err" || {
  echo "FAIL anonymous failure did not retry with credentials" >&2
  cat "$tmp/git-err" >&2
  exit 1
}
assert_eq "$(grep -E '^TOKEN_' "$GIT_LOG")" "$(printf '%s\n' TOKEN_UNSET TOKEN_UNSET)" "git does not see the token"
[ "$(grep -c '^ARGV ' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL expected anonymous attempt then authenticated retry" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q -- "credential.helper=" "$GIT_LOG" || {
  echo "FAIL git was invoked without clearing credential helpers" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q -- "https://github.com/grafana/dashboard-linter.git" "$GIT_LOG" || {
  echo "FAIL git URL was not the action repository" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q -- "ls-remote --tags --" "$GIT_LOG" || {
  echo "FAIL git was not an ls-remote of tags" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
if grep -F -q "example_token" "$GIT_LOG"; then
  echo "FAIL git log contains the token" >&2
  exit 1
fi
grep -F -q "USER x-access-token" "$GIT_LOG" || {
  echo "FAIL authenticated retry did not ask for the x-access-token user" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q "ASKPASS_MATCH" "$GIT_LOG" || {
  echo "FAIL askpass did not return the token" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q "ASKPASS_CLEAN" "$GIT_LOG" || {
  echo "FAIL askpass script contains the token" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
if grep -F -q "TOKEN_IN_ASKPASS" "$GIT_LOG"; then
  echo "FAIL askpass script contains the token" >&2
  exit 1
fi
[ "$(grep -c '^CONFIG_COUNT_UNSET$' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL git inherited GIT_CONFIG_COUNT" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
if grep -F -q "PWD $(pwd)" "$GIT_LOG"; then
  echo "FAIL git ran inside the caller repository" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
[ "$(grep -c '^SSL \[unset\]$' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL git inherited GIT_SSL_NO_VERIFY" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
[ "$(grep -c '^VERBOSE \[unset\]$' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL git inherited GIT_CURL_VERBOSE" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
[ "$(grep -c '^PRELOAD \[unset\]$' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL git inherited LD_PRELOAD" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
[ "$(grep -c '^BASHENV \[unset\]$' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL git inherited BASH_ENV" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
[ "$(grep -c '^PROTO \[https\]$' "$GIT_LOG")" -eq 2 ] || {
  echo "FAIL git was not limited to the https protocol" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
# shellcheck disable=SC2031
if grep -F -q "HOME [${HOME}]" "$GIT_LOG"; then
  echo "FAIL git used the caller home directory" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
if grep -F -q "BASH_ENV_RAN" "$GIT_LOG"; then
  echo "FAIL BASH_ENV was sourced" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
grep -F -q -- "http.sslVerify=true" "$GIT_LOG" || {
  echo "FAIL git was invoked without forcing TLS verification" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q -- "http.followRedirects=false" "$GIT_LOG" || {
  echo "FAIL git was allowed to follow redirects" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q -- "protocol.file.allow=never" "$GIT_LOG" || {
  echo "FAIL git was invoked without blocking file://" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
printf 'ok %s\n' "tag lookup retries with askpass"

: >"$GIT_LOG"
if (
  # shellcheck disable=SC2030,SC2031
  PATH="$tmp/bin:${PATH}"
  export PATH GIT_LOG
  unset GITHUB_ACTIONS
  # shellcheck disable=SC2031
  export GITHUB_TOKEN='example_token;id'
  fetch_remote_tags "grafana/dashboard-linter" >"$tmp/bad-token-out" 2>"$tmp/bad-token-err"
); then
  echo "FAIL a token with shell metacharacters should be refused" >&2
  exit 1
fi
[ "$(grep -c '^ARGV ' "$GIT_LOG")" -eq 1 ] || {
  echo "FAIL rejected token was still passed to an authenticated git" >&2
  cat "$GIT_LOG" >&2
  exit 1
}
grep -F -q "unexpected characters" "$tmp/bad-token-err" || {
  echo "FAIL rejected token error was not explained" >&2
  cat "$tmp/bad-token-err" >&2
  exit 1
}
if grep -F -q 'example_token;id' "$GIT_LOG"; then
  echo "FAIL rejected token was written to the git log" >&2
  exit 1
fi
printf 'ok %s\n' "tag lookup refuses an unsafe token"

: >"$GIT_LOG"
if (
  # shellcheck disable=SC2030,SC2031
  PATH="$tmp/bin:${PATH}"
  export PATH GIT_LOG
  git_ls_remote_tags "https://evil.example/grafana/dashboard-linter.git"
) >/dev/null 2>"$tmp/git-url-err"; then
  echo "FAIL unexpected git URL should be rejected" >&2
  exit 1
fi
if [ -s "$GIT_LOG" ]; then
  echo "FAIL unexpected git URL invoked git" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
grep -F -q "unexpected URL" "$tmp/git-url-err" || {
  echo "FAIL unexpected URL error was not explained" >&2
  cat "$tmp/git-url-err" >&2
  exit 1
}
printf 'ok %s\n' "tag lookup refuses a non-github URL"

: >"$GIT_LOG"
if (
  # shellcheck disable=SC2031
  PATH="$tmp/bin:${PATH}"
  export PATH GIT_LOG
  fetch_remote_tags "grafana/../evil" >"$tmp/bad-repo" 2>"$tmp/bad-repo-err"
); then
  echo "FAIL invalid repository should not list tags" >&2
  exit 1
fi
if [ -s "$GIT_LOG" ]; then
  echo "FAIL invalid repository invoked git" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
printf 'ok %s\n' "tag lookup rejects a bad repository"

: >"$GIT_LOG"
# A comment ends a continued for-list, so the credential-shaped URL lives in
# an array where "# trufflehog:ignore" can share its line.
bad_urls=(
  "https://user:token@github.com/grafana/dashboard-linter.git" # trufflehog:ignore
  "https://github.com/grafana/foo/bar.git"
  "HTTPS://github.com/grafana/dashboard-linter.git"
)
for bad_url in "${bad_urls[@]}"; do
  if git_ls_remote_tags "$bad_url" >"$tmp/bad-url-out" 2>"$tmp/bad-url-err"; then
    echo "FAIL URL should be rejected: ${bad_url}" >&2
    exit 1
  fi
done
if [ -s "$GIT_LOG" ]; then
  echo "FAIL rejected URL invoked git" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
printf 'ok %s\n' "tag lookup refuses userinfo, extra path, and uppercase host"

: >"$GIT_LOG"
if (
  shopt -s nocasematch
  git_ls_remote_tags "HTTPS://github.com/grafana/dashboard-linter.git" >"$tmp/nocase-out" 2>"$tmp/nocase-err"
); then
  echo "FAIL nocasematch accepted a non-lowercase URL" >&2
  exit 1
fi
if [ -s "$GIT_LOG" ]; then
  echo "FAIL nocasematch URL check invoked git" >&2
  cat "$GIT_LOG" >&2
  exit 1
fi
printf 'ok %s\n' "tag lookup ignores nocasematch"

status=0
rel_err="$(
  cd "$tmp"
  # shellcheck disable=SC2123
  PATH="bin"
  unset GITHUB_ACTIONS
  git_executable 2>&1
)" || status=$?
[ "$status" -ne 0 ] || {
  echo "FAIL relative git path should be refused" >&2
  printf '%s\n' "$rel_err" >&2
  exit 1
}
grep -F -q "relative path" "$tmp/rel-err" 2>/dev/null || grep -F -q "relative path" <<<"$rel_err" || {
  echo "FAIL relative git error was not explained" >&2
  printf '%s\n' "$rel_err" >&2
  exit 1
}
printf 'ok %s\n' "refuse a relative git path"

status=0
plant_err="$(
  GITHUB_ACTIONS=true
  GITHUB_WORKSPACE="$tmp"
  RUNNER_TEMP="$tmp/runner"
  TMPDIR="$tmp"
  PATH="$tmp/bin:/usr/bin:/bin"
  git_executable 2>&1
)" || status=$?
[ "$status" -ne 0 ] || {
  echo "FAIL git under the workspace should be refused in Actions" >&2
  printf '%s\n' "$plant_err" >&2
  exit 1
}
grep -F -q "workspace or a temp directory" <<<"$plant_err" || {
  echo "FAIL planted git error was not explained" >&2
  printf '%s\n' "$plant_err" >&2
  exit 1
}
printf 'ok %s\n' "refuse git planted in the workspace"

got="$(
  GITHUB_ACTIONS=true
  GITHUB_WORKSPACE="$tmp/not-the-workspace"
  RUNNER_TEMP="$tmp/not-the-runner-temp"
  TMPDIR="$tmp"
  PATH="/usr/bin:/bin:/opt/homebrew/bin:/usr/local/bin"
  git_executable
)"
case "$got" in
  /*) ;;
  *)
    echo "FAIL system git path is not absolute: ${got}" >&2
    exit 1
    ;;
esac
case "$got" in
  "$tmp"/*)
    echo "FAIL system git resolved under temp: ${got}" >&2
    exit 1
    ;;
esac
printf 'ok %s\n' "use an absolute system git in Actions"

printf 'ok %s\n' "all action tests passed"
