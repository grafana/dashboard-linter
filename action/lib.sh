#!/usr/bin/env bash
# Shared helpers for the composite action. Asset names match .goreleaser.yaml:
#   {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
# with tar.gz everywhere except windows, which is a zip. .Version has no "v".

# nocasematch would let a URL or tag slip past the literal case patterns below.
shopt -u nocasematch

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

# is_commit_sha REF is true for a full Git commit object name.
# GitHub action pins use lowercase hex.
is_commit_sha() {
  local ref="$1"
  # shellcheck disable=SC2076
  local sha_re='^[0-9a-f]{40}$'
  [ "${#ref}" -eq 40 ] || return 1
  # shellcheck disable=SC2076
  [[ "$ref" =~ $sha_re ]]
}

# list_has_line NEEDLE LIST is true when NEEDLE is one entire line of LIST.
# The comparison is literal, so a tag name cannot act as a glob.
list_has_line() {
  local needle="$1"
  local list="$2"
  local item=""
  [ -n "$needle" ] || return 1
  while IFS= read -r item || [ -n "$item" ]; do
    if [ "$item" = "$needle" ]; then
      return 0
    fi
  done <<< "$list"
  return 1
}

# parse_ls_remote_tag LINE sets TAG_OBJECT, TAG_NAME, and TAG_PEELED.
# TAG_PEELED is 1 for the peeled commit line of an annotated tag
# ("refs/tags/<name>^{}"). Lines that are not tag refs are rejected.
parse_ls_remote_tag() {
  local line="$1"
  local ref=""
  local name=""
  TAG_OBJECT=""
  TAG_NAME=""
  TAG_PEELED=0
  case "$line" in
    *$'\t'*) ;;
    *) return 1 ;;
  esac
  TAG_OBJECT="${line%%$'\t'*}"
  ref="${line#*$'\t'}"
  case "$ref" in
    *$'\t'* | *[[:space:]]*) return 1 ;;
  esac
  case "$ref" in
    refs/tags/*) ;;
    *) return 1 ;;
  esac
  is_commit_sha "$TAG_OBJECT" || return 1
  name="${ref#refs/tags/}"
  case "$name" in
    *'^{}')
      TAG_PEELED=1
      name="${name%'^{}'}"
      ;;
  esac
  case "$name" in
    "" | */* | *'^{}'* | *$'\n'* | *$'\r'* | *[[:space:]]*) return 1 ;;
  esac
  TAG_NAME="$name"
}

# version_from_tag_lines SHA LINES prints the one release tag whose commit is SHA.
# LINES is `git ls-remote --tags` output. Annotated tags match the peeled
# commit, not the tag object. More than one release tag fails: a second tag
# created after the pin must not change which release is downloaded.
version_from_tag_lines() {
  local sha="$1"
  local lines="$2"
  local line=""
  local annotated=""
  local matches=""
  local match=""
  local display=""
  local match_count=0

  # A caller may have turned this back on after the file was sourced.
  shopt -u nocasematch

  if ! is_commit_sha "$sha"; then
    echo "invalid commit sha" >&2
    return 1
  fi
  # A hostile tag advertisement must not pin the runner in an O(n^2) parse.
  if [ "${#lines}" -gt 262144 ]; then
    echo "tag list is too large to map ${sha} to a release; set version" >&2
    return 1
  fi

  while IFS= read -r line || [ -n "$line" ]; do
    line="$(trim "$line")"
    [ -n "$line" ] || continue
    parse_ls_remote_tag "$line" || continue
    if [ "$TAG_PEELED" -ne 1 ]; then
      continue
    fi
    if ! list_has_line "$TAG_NAME" "$annotated"; then
      annotated="${annotated}${TAG_NAME}"$'\n'
    fi
    if [ "$TAG_OBJECT" = "$sha" ] && version_is_valid "$TAG_NAME"; then
      if ! list_has_line "$TAG_NAME" "$matches"; then
        matches="${matches}${TAG_NAME}"$'\n'
      fi
    fi
  done <<< "$lines"

  while IFS= read -r line || [ -n "$line" ]; do
    line="$(trim "$line")"
    [ -n "$line" ] || continue
    parse_ls_remote_tag "$line" || continue
    if [ "$TAG_PEELED" -eq 1 ]; then
      continue
    fi
    # This line is the tag object of an annotated tag, not the commit.
    if list_has_line "$TAG_NAME" "$annotated"; then
      continue
    fi
    if [ "$TAG_OBJECT" = "$sha" ] && version_is_valid "$TAG_NAME"; then
      if ! list_has_line "$TAG_NAME" "$matches"; then
        matches="${matches}${TAG_NAME}"$'\n'
      fi
    fi
  done <<< "$lines"

  while IFS= read -r match || [ -n "$match" ]; do
    [ -n "$match" ] || continue
    match_count=$((match_count + 1))
    if [ -z "$display" ]; then
      display="$match"
    else
      display="${display}, ${match}"
    fi
  done <<< "$matches"

  if [ "$match_count" -eq 1 ]; then
    printf '%s' "$display"
    return 0
  fi
  if [ "$match_count" -eq 0 ]; then
    echo "no release tag points at ${sha}; set version" >&2
    return 1
  fi
  echo "commit ${sha} matches more than one release tag (${display}); set version" >&2
  return 1
}

# git_executable prints an absolute path to git.
# In Actions, a git binary under the workspace or temp dir is refused:
# an earlier step can prepend PATH and plant a fake git that forges tags.
git_executable() {
  local bin=""
  bin="$(command -v git 2>/dev/null)" || {
    echo "git is required to map a pinned action sha to a release tag; set version instead" >&2
    return 1
  }
  case "$bin" in
    /*) ;;
    *)
      echo "refusing to run git from a relative path" >&2
      return 1
      ;;
  esac
  case "$bin" in
    *$'\n'* | *$'\r'*)
      echo "refusing to run git from a path that contains a newline" >&2
      return 1
      ;;
  esac
  if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
    case "$bin" in
      "${GITHUB_WORKSPACE:-/nonexistent-workspace}"/* | \
        "${RUNNER_TEMP:-/nonexistent-runner-temp}"/* | \
        "${TMPDIR:-/tmp}"/*)
        echo "refusing to run git from the workspace or a temp directory" >&2
        return 1
        ;;
    esac
  fi
  printf '%s' "$bin"
}

# git_ls_remote_tags URL [auth] prints `git ls-remote --tags` for URL.
# The command runs outside any repository, with TLS verification forced and
# only the https protocol allowed, so a checkout cannot redirect it with
# url.insteadOf, a custom CA, or a credential helper. auth sends GITHUB_TOKEN
# through askpass. The token is not placed on the git command line.
# Anonymous mode removes it from the environment first.
git_ls_remote_tags() {
  local url="$1"
  local mode="${2:-anon}"
  local work=""
  local askpass=""
  local token_file=""
  local path_after_host=""
  local git_bin=""
  local old_umask=""
  local status=0
  local errexit=0

  shopt -u nocasematch
  case "$url" in
    https://github.com/*/*.git) ;;
    *)
      echo "refusing tag lookup for unexpected URL" >&2
      return 1
      ;;
  esac
  path_after_host="${url#https://github.com/}"
  case "$path_after_host" in
    */*/* | *[!A-Za-z0-9_./-]*)
      echo "refusing tag lookup for unexpected URL" >&2
      return 1
      ;;
  esac
  git_bin="$(git_executable)" || return 1

  if [ "$mode" = "auth" ]; then
    # GitHub tokens are letters, digits, and underscore. Anything else is
    # refused so the askpass exchange cannot be split or expanded.
    case "${GITHUB_TOKEN:-}" in
      "")
        echo "could not list release tags: no token for authenticated retry" >&2
        return 1
        ;;
      *[!A-Za-z0-9_]*)
        echo "refusing to use a token that contains unexpected characters" >&2
        return 1
        ;;
    esac
    if [ "${#GITHUB_TOKEN}" -gt 512 ]; then
      echo "refusing to use a token that is too long" >&2
      return 1
    fi
  fi

  old_umask="$(umask)"
  umask 077
  work="$(mktemp -d)" || {
    umask "$old_umask"
    return 1
  }
  chmod 700 "$work" || {
    rm -rf "$work"
    umask "$old_umask"
    return 1
  }
  if [ "$mode" = "auth" ]; then
    # The token is a file, not an environment variable, so a traced askpass
    # shell cannot print it. The file is mode 600 inside a 700 directory.
    token_file="${work}/token"
    printf '%s\n' "$GITHUB_TOKEN" >"$token_file" || {
      rm -rf "$work"
      umask "$old_umask"
      return 1
    }
    chmod 600 "$token_file" || {
      rm -rf "$work"
      umask "$old_umask"
      return 1
    }
    askpass="${work}/askpass"
    cat >"$askpass" <<'EOF'
#!/bin/sh
case "$1" in
  *Username*) printf '%s\n' "x-access-token" ;;
  *) cat -- "$TOKEN_PATH" ;;
esac
EOF
    chmod 700 "$askpass" || {
      rm -rf "$work"
      umask "$old_umask"
      return 1
    }
  fi
  umask "$old_umask"

  case $- in
    *e*) errexit=1 ;;
  esac
  set +e
  (
    # SHELLOPTS is readonly. set +x covers this shell; the askpass script
    # reads the token from a file so a traced child does not expand it.
    set +x
    set +v
    unset CDPATH BASH_ENV ENV \
      LD_PRELOAD DYLD_INSERT_LIBRARIES DYLD_LIBRARY_PATH \
      GIT_DIR GIT_WORK_TREE GIT_EXEC_PATH GIT_TEMPLATE_DIR GIT_NAMESPACE \
      GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT \
      GIT_SSL_NO_VERIFY GIT_SSL_CAINFO GIT_SSL_CAPATH GIT_SSL_VERSION \
      GIT_CURL_VERBOSE GIT_TRACE GIT_TRACE_PACKET GIT_TRACE_PERFORMANCE \
      GIT_TRACE_SETUP GIT_TRACE_CURL GIT_TRANSFER_TRACE \
      GIT_SSH GIT_SSH_COMMAND GIT_SSH_VARIANT GIT_PROXY_COMMAND \
      SSL_CERT_FILE SSL_CERT_DIR CURL_CA_BUNDLE REQUESTS_CA_BUNDLE
    cd -P "$work" >/dev/null || exit 1
    export HOME="$work"
    export XDG_CONFIG_HOME="$work"
    export GIT_TERMINAL_PROMPT=0
    export GIT_CONFIG_NOSYSTEM=1
    export GIT_CONFIG_GLOBAL=/dev/null
    export GIT_CONFIG_SYSTEM=/dev/null
    export GIT_CEILING_DIRECTORIES="$work"
    export GIT_ALLOW_PROTOCOL=https
    unset GITHUB_TOKEN GH_TOKEN GITHUB_ENTERPRISE_TOKEN GH_ENTERPRISE_TOKEN
    if [ "$mode" != "auth" ]; then
      export GIT_ASKPASS=true
      unset SSH_ASKPASS SSH_ASKPASS_REQUIRE TOKEN_PATH
    else
      export TOKEN_PATH="$token_file"
      export GIT_ASKPASS="$askpass"
      export SSH_ASKPASS="$askpass"
      export SSH_ASKPASS_REQUIRE=force
    fi
    "$git_bin" \
      -c credential.helper= \
      -c http.sslVerify=true \
      -c http.sslVersion=tlsv1.2 \
      -c http.followRedirects=false \
      -c http.connectTimeout=20 \
      -c http.lowSpeedLimit=1000 \
      -c http.lowSpeedTime=30 \
      -c protocol.file.allow=never \
      ls-remote --tags -- "$url"
  )
  status=$?
  if [ "$errexit" -eq 1 ]; then
    set -e
  fi
  rm -rf "$work"
  return "$status"
}

# fetch_remote_tags REPO prints tags for owner/repo.
# Public repositories are listed anonymously. GITHUB_TOKEN is sent only when
# that fails, matching the release download.
fetch_remote_tags() {
  local repo="$1"
  local url=""
  local tags=""
  repo="$(action_repository_value "$repo")" || return 1
  url="https://github.com/${repo}.git"
  if tags="$(git_ls_remote_tags "$url")"; then
    printf '%s\n' "$tags"
    return 0
  fi
  if [ -z "${GITHUB_TOKEN:-}" ]; then
    echo "could not list release tags for ${repo}; set version to skip the lookup" >&2
    return 1
  fi
  echo "anonymous tag lookup failed; retrying with credentials" >&2
  if ! tags="$(git_ls_remote_tags "$url" auth)"; then
    echo "could not list release tags for ${repo}; set version to skip the lookup" >&2
    return 1
  fi
  printf '%s\n' "$tags"
}

# version_for_pinned_sha SHA prints the release tag that points at SHA.
version_for_pinned_sha() {
  local sha="$1"
  local repo=""
  local tags=""
  local version=""
  if ! is_commit_sha "$sha"; then
    echo "invalid commit sha" >&2
    return 1
  fi
  repo="$(action_repository)" || return 1
  tags="$(fetch_remote_tags "$repo")" || return 1
  version="$(version_from_tag_lines "$sha" "$tags")" || return 1
  echo "resolved release ${version} from action ref ${sha}" >&2
  printf '%s' "$version"
}

# action_ref prints the ref this action was invoked with.
# Precedence:
#   1. ACTION_REF, set by action.yml from github.action_ref. The GITHUB_ACTION_REF
#      name cannot be used for that copy: composite run steps clear it after the
#      step env is applied (https://github.com/actions/runner/issues/2525).
#   2. The ref in GITHUB_ACTION_PATH (.../_actions/owner/name/<ref>). That path
#      is set for composite steps even when the GitHub-prefixed variables are empty.
#   3. GITHUB_ACTION_REF, for a non-composite caller that still has it set.
action_ref() {
  local ref=""
  ref="$(trim "${ACTION_REF:-}")"
  if [ -n "$ref" ]; then
    printf '%s' "$ref"
    return 0
  fi
  if [ -n "${GITHUB_ACTION_PATH:-}" ]; then
    ref="$(ref_from_action_path "$GITHUB_ACTION_PATH" || true)"
  fi
  if [ -n "$ref" ]; then
    printf '%s' "$ref"
    return 0
  fi
  ref="$(trim "${GITHUB_ACTION_REF:-}")"
  printf '%s' "$ref"
}

# resolve_version INPUT REF prints the release tag to install.
# An explicit input wins. Otherwise REF may be a release tag
# (v0.4.0 when the action is pinned to that tag) or the commit a single
# release tag points at. install.sh passes action_ref.
resolve_version() {
  local input="$1"
  local ref="$2"
  local version=""
  input="$(trim "$input")"
  ref="$(trim "$ref")"
  if [ -n "$input" ]; then
    version="$input"
  else
    version="${ref#refs/tags/}"
    version="$(trim "$version")"
    if ! version_is_valid "$version" && is_commit_sha "$version"; then
      version="$(version_for_pinned_sha "$version")" || return 1
    fi
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

# action_path_fields PATH prints owner, name, and ref on separate lines.
# PATH is GITHUB_ACTION_PATH, for example
# /home/runner/work/_actions/grafana/dashboard-linter/<ref>. The ref is
# everything after owner/name, so a branch ref may contain slashes.
# Windows separators are accepted.
action_path_fields() {
  local path="$1"
  local rest=""
  local owner=""
  local name=""
  local ref=""
  path="${path//\\//}"
  path="${path%/}"
  rest="${path#*_actions/}"
  if [ "$rest" = "$path" ] || [ -z "$rest" ]; then
    return 1
  fi
  owner="${rest%%/*}"
  rest="${rest#*/}"
  if [ -z "$owner" ] || [ "$rest" = "$owner" ]; then
    return 1
  fi
  name="${rest%%/*}"
  ref="${rest#*/}"
  if [ -z "$name" ] || [ -z "$ref" ] || [ "$ref" = "$rest" ]; then
    return 1
  fi
  case "$owner" in
    *$'\n'* | *$'\r'*) return 1 ;;
  esac
  case "$name" in
    *$'\n'* | *$'\r'*) return 1 ;;
  esac
  case "$ref" in
    *$'\n'* | *$'\r'*) return 1 ;;
  esac
  printf '%s\n%s\n%s\n' "$owner" "$name" "$ref"
}

# repository_from_action_path PATH prints owner/repo from an action directory.
repository_from_action_path() {
  local parts=""
  local owner=""
  local rest=""
  local name=""
  parts="$(action_path_fields "$1")" || return 1
  owner="${parts%%$'\n'*}"
  rest="${parts#*$'\n'}"
  name="${rest%%$'\n'*}"
  if [ -z "$owner" ] || [ -z "$name" ]; then
    return 1
  fi
  printf '%s/%s' "$owner" "$name"
}

# ref_from_action_path PATH prints the action ref from an action directory.
ref_from_action_path() {
  local parts=""
  local rest=""
  local ref=""
  parts="$(action_path_fields "$1")" || return 1
  rest="${parts#*$'\n'}"
  ref="${rest#*$'\n'}"
  if [ -z "$ref" ]; then
    return 1
  fi
  printf '%s' "$ref"
}

action_repository() {
  local repo=""
  # ACTION_REPOSITORY is the copy action.yml makes. GITHUB_ACTION_REPOSITORY is
  # read last: composite run steps clear it, and a previous action can leave a
  # stale value behind (https://github.com/actions/runner/issues/803).
  repo="$(trim "${ACTION_REPOSITORY:-}")"
  if [ -z "$repo" ] && [ -n "${GITHUB_ACTION_PATH:-}" ]; then
    repo="$(repository_from_action_path "$GITHUB_ACTION_PATH" || true)"
  fi
  if [ -z "$repo" ]; then
    repo="$(trim "${GITHUB_ACTION_REPOSITORY:-}")"
  fi
  if [ -z "$repo" ]; then
    if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
      echo "could not determine the action repository from ACTION_REPOSITORY, GITHUB_ACTION_PATH, or GITHUB_ACTION_REPOSITORY" >&2
      return 1
    fi
    repo="grafana/dashboard-linter"
  fi
  action_repository_value "$repo"
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
