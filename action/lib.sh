#!/usr/bin/env bash
# Shared helpers for the composite action. Asset names match .goreleaser.yaml:
#   {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
# with tar.gz everywhere except windows, which is a zip. .Version has no "v".

# trim removes leading and trailing whitespace, including CR.
trim() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
}

# version_is_valid accepts release tags this repo publishes, such as v0.4.0
# and v1.2.3-rc.1. It rejects anything that could change the download path.
version_is_valid() {
  local version="$1"
  # Bash 3.2 (macOS /bin/bash) treats a quoted =~ pattern as literal text.
  # shellcheck disable=SC2076
  local version_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$'
  case "$version" in
    *$'\n'* | *$'\r'* | *..* | *. | *- | */*) return 1 ;;
  esac
  # shellcheck disable=SC2076
  [[ "$version" =~ $version_re ]] || return 1
  [ "${#version}" -le 64 ]
}

# resolve_version INPUT REF prints the release tag to install.
# An explicit input wins. Otherwise REF must already be a release tag
# (GITHUB_ACTION_REF is "v0.4.0" when the action is pinned to that tag).
resolve_version() {
  local input="$1"
  local ref="$2"
  local version=""
  input="$(trim "$input")"
  if [ -n "$input" ]; then
    version="$input"
  else
    version="${ref#refs/tags/}"
    version="$(trim "$version")"
  fi
  if ! version_is_valid "$version"; then
    if [ -n "$input" ]; then
      echo "invalid version: ${input}" >&2
    else
      echo "version is required when the action ref is not a release tag (ref: ${ref:-unset})" >&2
    fi
    return 1
  fi
  printf '%s' "$version"
}

# release_asset VERSION OS ARCH sets RELEASE_ASSET and RELEASE_BINARY.
# OS and ARCH use GitHub's RUNNER_OS / RUNNER_ARCH values.
release_asset() {
  local version="$1"
  local os="$2"
  local arch="$3"
  local goos=""
  local goarch=""
  local ext=""
  local nov=""
  if ! version_is_valid "$version"; then
    echo "invalid version: ${version}" >&2
    return 1
  fi
  case "$os" in
    Linux) goos="linux" ;;
    macOS) goos="darwin" ;;
    Windows) goos="windows" ;;
    *)
      echo "unsupported runner OS: ${os}" >&2
      return 1
      ;;
  esac
  case "$arch" in
    X64) goarch="amd64" ;;
    ARM64) goarch="arm64" ;;
    *)
      echo "unsupported runner architecture: ${arch}" >&2
      return 1
      ;;
  esac
  if [ "${goos}/${goarch}" = "windows/arm64" ]; then
    echo "no dashboard-linter release is published for windows/arm64" >&2
    return 1
  fi
  if [ "$goos" = "windows" ]; then
    ext="zip"
    # Read by install.sh after release_asset returns.
    # shellcheck disable=SC2034
    RELEASE_BINARY="dashboard-linter.exe"
  else
    ext="tar.gz"
    # shellcheck disable=SC2034
    RELEASE_BINARY="dashboard-linter"
  fi
  nov="${version#v}"
  # shellcheck disable=SC2034
  RELEASE_ASSET="dashboard-linter_${nov}_${goos}_${goarch}.${ext}"
}

detect_runner_os() {
  if [ -n "${RUNNER_OS:-}" ]; then
    printf '%s' "$RUNNER_OS"
    return 0
  fi
  local uname_s=""
  uname_s="$(uname -s)"
  case "$uname_s" in
    Linux) printf 'Linux' ;;
    Darwin) printf 'macOS' ;;
    MINGW* | MSYS* | CYGWIN*) printf 'Windows' ;;
    *)
      echo "unsupported operating system: ${uname_s}" >&2
      return 1
      ;;
  esac
}

detect_runner_arch() {
  if [ -n "${RUNNER_ARCH:-}" ]; then
    printf '%s' "$RUNNER_ARCH"
    return 0
  fi
  local uname_m=""
  uname_m="$(uname -m)"
  case "$uname_m" in
    x86_64 | amd64) printf 'X64' ;;
    aarch64 | arm64) printf 'ARM64' ;;
    *)
      echo "unsupported architecture: ${uname_m}" >&2
      return 1
      ;;
  esac
}

action_repository() {
  local repo="${GITHUB_ACTION_REPOSITORY:-}"
  if [ -z "$repo" ]; then
    # Inside Actions the repository must come from the runner. Falling back
    # to the upstream repo would let a fork action download upstream assets
    # without that being visible in the uses: line.
    if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
      echo "GITHUB_ACTION_REPOSITORY is not set" >&2
      return 1
    fi
    repo="grafana/dashboard-linter"
  fi
  # shellcheck disable=SC2076
  local repo_re='^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$'
  case "$repo" in
    *$'\n'* | *$'\r'* | *..* | */*/*) 
      echo "invalid action repository: ${repo}" >&2
      return 1
      ;;
  esac
  # shellcheck disable=SC2076
  if [[ ! "$repo" =~ $repo_re ]]; then
    echo "invalid action repository: ${repo}" >&2
    return 1
  fi
  printf '%s' "$repo"
}

# release_download_url REPO VERSION ASSET prints an https URL under
# /releases/download/. Every component is validated so the path cannot escape
# that prefix.
release_download_url() {
  local repo="$1"
  local version="$2"
  local asset="$3"
  # shellcheck disable=SC2076
  local asset_re='^[A-Za-z0-9._-]+$'
  if ! version_is_valid "$version"; then
    echo "invalid version: ${version}" >&2
    return 1
  fi
  repo="$(action_repository_value "$repo")" || return 1
  case "$asset" in
    . | ..)
      echo "invalid release asset: ${asset}" >&2
      return 1
      ;;
    *$'\n'* | *$'\r'* | *..* | */*)
      echo "invalid release asset: ${asset}" >&2
      return 1
      ;;
  esac
  # shellcheck disable=SC2076
  if [[ ! "$asset" =~ $asset_re ]]; then
    echo "invalid release asset: ${asset}" >&2
    return 1
  fi
  printf 'https://github.com/%s/releases/download/%s/%s' "$repo" "$version" "$asset"
}

# action_repository_value REPO validates REPO instead of reading the environment.
action_repository_value() {
  local repo="$1"
  # shellcheck disable=SC2076
  local repo_re='^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$'
  case "$repo" in
    *$'\n'* | *$'\r'* | *..* | */*/*)
      echo "invalid action repository: ${repo}" >&2
      return 1
      ;;
  esac
  # shellcheck disable=SC2076
  if [[ ! "$repo" =~ $repo_re ]]; then
    echo "invalid action repository: ${repo}" >&2
    return 1
  fi
  printf '%s' "$repo"
}

is_sha256() {
  local sum="$1"
  # shellcheck disable=SC2076
  local hex_re='^[0-9a-f]{64}$'
  [ "${#sum}" -eq 64 ] || return 1
  # shellcheck disable=SC2076
  [[ "$sum" =~ $hex_re ]]
}

# checksum_for FILE ASSET prints the sha256 goreleaser recorded for ASSET.
checksum_for() {
  local file="$1"
  local asset="$2"
  local sum=""
  sum="$(tr -d '\r' < "$file" | awk -v asset="$asset" 'NF == 2 && $2 == asset && $1 ~ /^[0-9a-f]{64}$/ { print $1 }')"
  if ! is_sha256 "$sum"; then
    echo "checksums.txt does not contain exactly one sha256 for ${asset}" >&2
    return 1
  fi
  printf '%s' "$sum"
}

file_sha256() {
  local file="$1"
  local out=""
  if command -v openssl >/dev/null 2>&1; then
    out="$(openssl dgst -sha256 "$file")" || return 1
    printf '%s' "${out##* }"
    return 0
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    out="$(sha256sum "$file")" || return 1
    printf '%s' "${out%% *}"
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    out="$(shasum -a 256 "$file")" || return 1
    printf '%s' "${out%% *}"
    return 0
  fi
  echo "no sha256 tool found (tried openssl, sha256sum, shasum)" >&2
  return 1
}

verify_sha256() {
  local file="$1"
  local expected="$2"
  local actual=""
  actual="$(file_sha256 "$file")" || return 1
  if [ "$actual" != "$expected" ]; then
    echo "checksum mismatch for $(basename "$file")" >&2
    echo "expected ${expected}" >&2
    echo "actual   ${actual}" >&2
    return 1
  fi
}

# extract_binary ARCHIVE MEMBER DEST writes one archive member to DEST.
# The member is streamed to stdout, so archive paths cannot overwrite other files.
extract_binary() {
  local archive="$1"
  local binary_name="$2"
  local dest="$3"
  local tmp=""
  tmp="$(mktemp)" || return 1
  if [[ "$archive" == *.zip ]]; then
    if ! extract_zip "$archive" "$binary_name" "$tmp"; then
      rm -f "$tmp"
      return 1
    fi
  else
    if ! tar -xOzf "$archive" "$binary_name" >"$tmp"; then
      rm -f "$tmp"
      echo "failed to extract ${binary_name} from $(basename "$archive")" >&2
      return 1
    fi
  fi
  if [ ! -s "$tmp" ]; then
    rm -f "$tmp"
    echo "archive did not contain ${binary_name}" >&2
    return 1
  fi
  chmod 0755 "$tmp" || {
    rm -f "$tmp"
    return 1
  }
  mv "$tmp" "$dest"
}

extract_zip() {
  local archive="$1"
  local binary_name="$2"
  local dest="$3"
  if command -v unzip >/dev/null 2>&1; then
    if ! unzip -p "$archive" "$binary_name" >"$dest"; then
      echo "failed to extract ${binary_name} from $(basename "$archive")" >&2
      return 1
    fi
    return 0
  fi
  if ! tar -xOf "$archive" "$binary_name" >"$dest"; then
    echo "failed to extract ${binary_name} from $(basename "$archive")" >&2
    return 1
  fi
}

download() {
  local url="$1"
  local dest="$2"
  local tmp="${dest}.partial"
  rm -f "$tmp"
  if ! curl_fetch "$url" "$tmp"; then
    rm -f "$tmp"
    echo "download failed: ${url}" >&2
    return 1
  fi
  mv "$tmp" "$dest"
}

# curl_fetch URL DEST downloads URL to DEST.
# Public releases are fetched anonymously. The workflow token is a credential
# for the calling repository, so it is sent only when the anonymous request
# fails (rate limit, or a private release). --location is not --location-trusted,
# so curl will not forward Authorization to the asset host on a cross-host redirect.
curl_fetch() {
  local url="$1"
  local dest="$2"
  local curl_args=(
    --proto '=https'
    --tlsv1.2
    --fail
    --silent
    --show-error
    --location
    --connect-timeout 20
    --max-time 180
    --retry 3
    --retry-delay 2
    --output "$dest"
  )
  if curl "${curl_args[@]}" "$url"; then
    return 0
  fi
  if [ -z "${GITHUB_TOKEN:-}" ]; then
    return 1
  fi
  echo "anonymous download failed; retrying with credentials" >&2
  rm -f "$dest"
  curl "${curl_args[@]}" --header "Authorization: Bearer ${GITHUB_TOKEN}" "$url"
}

github_path_entry() {
  local dir="$1"
  local os="$2"
  if [ "$os" = "Windows" ] && command -v cygpath >/dev/null 2>&1; then
    cygpath -w "$dir"
    return 0
  fi
  printf '%s\n' "$dir"
}

append_github_path() {
  local dir="$1"
  local os="$2"
  local entry=""
  if [ -z "${GITHUB_PATH:-}" ]; then
    return 0
  fi
  entry="$(github_path_entry "$dir" "$os")" || return 1
  entry="$(trim "$entry")"
  case "$entry" in
    *$'\n'* | *$'\r'*)
      echo "install directory contains a newline" >&2
      return 1
      ;;
  esac
  printf '%s\n' "$entry" >>"$GITHUB_PATH"
}

write_outputs() {
  local version="$1"
  local bin="$2"
  if [ -z "${GITHUB_OUTPUT:-}" ]; then
    return 0
  fi
  case "${version}${bin}" in
    *$'\n'* | *$'\r'*)
      echo "refusing to write action output that contains a newline" >&2
      return 1
      ;;
  esac
  {
    printf 'version=%s\n' "$version"
    printf 'bin=%s\n' "$bin"
  } >>"$GITHUB_OUTPUT"
}

is_glob() {
  case "$1" in
    *'*'* | *'?'* | *'['*) return 0 ;;
  esac
  return 1
}

# list_json_files DIR prints *.json files under DIR, except node_modules.
list_json_files() {
  local dir="$1"
  find "$dir" \
    \( -type d -name node_modules -prune \) -o \
    \( -type f -name '*.json' -print \) |
    sort
}

expand_glob() {
  local pattern="$1"
  local nullglob=0
  local noglob=0
  local matches=()
  local match=""
  case "$pattern" in
    *[[:space:]]*)
      echo "globs cannot contain whitespace; pass a directory or a literal path: ${pattern}" >&2
      return 1
      ;;
  esac
  shopt -q nullglob && nullglob=1
  case $- in
    *f*) noglob=1 ;;
  esac
  shopt -s nullglob
  set +f
  # Intentional pathname expansion of a user glob. Whitespace is rejected above.
  # shellcheck disable=SC2206
  matches=( $pattern )
  if [ "$nullglob" -eq 0 ]; then
    shopt -u nullglob
  fi
  if [ "$noglob" -eq 1 ]; then
    set -f
  else
    set +f
  fi
  if [ "${#matches[@]}" -eq 0 ]; then
    echo "dashboard path not found: ${pattern}" >&2
    return 1
  fi
  for match in "${matches[@]}"; do
    if [ -d "$match" ]; then
      list_json_files "$match" || return 1
    elif [ -f "$match" ]; then
      printf '%s\n' "$match"
    else
      echo "dashboard path not found: ${match}" >&2
      return 1
    fi
  done
}

# collect_files TEXT prints the dashboard paths described by TEXT, one per line.
# TEXT is newline-separated files, directories, or shell globs.
collect_files() {
  local input="$1"
  local raw=""
  raw="$(collect_files_raw "$input")" || return 1
  if [ -z "$raw" ]; then
    return 0
  fi
  printf '%s\n' "$raw" | sort -u
}

collect_files_raw() {
  local input="$1"
  local line=""
  while IFS= read -r line || [ -n "$line" ]; do
    line="$(trim "$line")"
    [ -n "$line" ] || continue
    case "$line" in
      *'**'*)
        echo "recursive globs are not supported; pass a directory: ${line}" >&2
        return 1
        ;;
    esac
    if [ -d "$line" ]; then
      list_json_files "$line" || return 1
    elif [ -f "$line" ]; then
      printf '%s\n' "$line"
    elif is_glob "$line"; then
      expand_glob "$line" || return 1
    else
      echo "dashboard path not found: ${line}" >&2
      return 1
    fi
  done <<< "$input"
}

# build_lint_argv BIN ARGS FILE sets LINT_ARGV to the lint invocation.
# ARGS is split on whitespace. Globbing is disabled so "*" stays a flag.
build_lint_argv() {
  local bin="$1"
  local args="$2"
  local file="$3"
  local extra=()
  local noglob=0
  args="$(trim "$args")"
  LINT_ARGV=( "$bin" lint )
  if [ -n "$args" ]; then
    case $- in
      *f*) noglob=1 ;;
    esac
    set -f
    # shellcheck disable=SC2206
    extra=( $args )
    if [ "$noglob" -eq 0 ]; then
      set +f
    fi
    # extra is non-empty: args survived trim, so word-splitting yields a field.
    LINT_ARGV=( "$bin" lint "${extra[@]}" )
  fi
  LINT_ARGV+=( "$file" )
}
