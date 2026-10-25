#!/usr/bin/env bash
# Download the dashboard-linter release for this runner, verify checksums.txt,
# and put the binary on PATH.
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib.sh
source "${script_dir}/lib.sh"

WORK_DIR=""
DEST_TMP=""

cleanup_download() {
  if [ -n "$WORK_DIR" ]; then
    rm -rf "$WORK_DIR"
  fi
  if [ -n "$DEST_TMP" ]; then
    rm -f "$DEST_TMP"
  fi
}
trap cleanup_download EXIT

install_release() {
  local repo="$1"
  local version="$2"
  local dest="$3"
  local checksum_url=""
  local archive_url=""
  local expected=""
  local archive=""

  WORK_DIR="$(mktemp -d)" || return 1
  DEST_TMP="${dest}.tmp"
  rm -f "$DEST_TMP"

  checksum_url="$(release_download_url "$repo" "$version" "checksums.txt")" || return 1
  archive_url="$(release_download_url "$repo" "$version" "$RELEASE_ASSET")" || return 1
  echo "Downloading ${archive_url}"
  download "$checksum_url" "${WORK_DIR}/checksums.txt" || return 1
  download "$archive_url" "${WORK_DIR}/${RELEASE_ASSET}" || return 1
  expected="$(checksum_for "${WORK_DIR}/checksums.txt" "$RELEASE_ASSET")" || return 1
  verify_sha256 "${WORK_DIR}/${RELEASE_ASSET}" "$expected" || return 1
  archive="${WORK_DIR}/${RELEASE_ASSET}"
  extract_binary "$archive" "$RELEASE_BINARY" "$DEST_TMP" || return 1
  # mv follows a symlink and would write outside this directory.
  if [ -L "$dest" ]; then
    rm -f "$dest"
  fi
  mv "$DEST_TMP" "$dest"
  DEST_TMP=""
  rm -rf "$WORK_DIR"
  WORK_DIR=""
}

main() {
  local input_version=""
  local version=""
  local os=""
  local arch=""
  local repo=""
  local cache=""
  local dest_dir=""
  local dest=""

  input_version="$(trim "${INPUT_VERSION:-}")"
  version="$(resolve_version "$input_version" "${GITHUB_ACTION_REF:-}")" || exit 1
  os="$(detect_runner_os)" || exit 1
  arch="$(detect_runner_arch)" || exit 1
  release_asset "$version" "$os" "$arch" || exit 1
  repo="$(action_repository)" || exit 1

  # RUNNER_TEMP is wiped at the start and end of the job. RUNNER_TOOL_CACHE
  # survives on self-hosted runners, so a later job could skip the checksum
  # by reusing a binary another job wrote there.
  cache="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
  dest_dir="${cache}/dashboard-linter/${version}/${os}-${arch}"
  mkdir -p "$dest_dir"
  dest_dir="$(cd "$dest_dir" && pwd)" || exit 1
  dest="${dest_dir}/${RELEASE_BINARY}"

  install_release "$repo" "$version" "$dest" || exit 1
  echo "Installed dashboard-linter ${version} (${RELEASE_ASSET}) to ${dest}"

  append_github_path "$dest_dir" "$os" || exit 1
  write_outputs "$version" "$dest" || exit 1
}

main "$@"
