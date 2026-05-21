#!/usr/bin/env bash
set -euo pipefail

# Build and publish the macOS arm64 release asset from a tagged macOS checkout.
# Required env vars: R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_ENDPOINT.
# Optional env vars: PUBLIC_DOWNLOAD_BASE, CDN_BUCKET, CDN_PREFIX, GITLAB_PROJECT_URL,
# MACOS_MIN_VERSION.
# If present, a repo-root .env file is loaded automatically.

DEFAULT_DOWNLOAD_BASE="https://cdn.usescrolls.com/scribe"
DEFAULT_CDN_BUCKET="agenthub-plugins"
DEFAULT_CDN_PREFIX="scribe"
DEFAULT_MACOS_MIN_VERSION="11.0"
MACOS_CLI_ASSET="scribe-cli-darwin-arm64"
MACOS_DESKTOP_ASSET="scribe-desktop-darwin-arm64"
LEGACY_MACOS_ASSET="scribe-darwin-arm64"
WINDOWS_ASSET="scribe-windows-amd64.exe"
LINUX_CLI_ASSET="scribe-cli-linux-amd64"
LINUX_DESKTOP_ASSET="scribe-desktop-linux-amd64"
LINUX_ASSET="scribe-linux-amd64"

log() {
    printf '%s\n' "$*"
}

fail() {
    printf 'Error: %s\n' "$*" >&2
    exit 1
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

repo_root() {
    git rev-parse --show-toplevel 2>/dev/null || fail "this script must run inside the git repository"
}

load_dotenv() {
    local env_file="$1"

    [[ -f "${env_file}" ]] || return 0

    log "Loading environment from ${env_file}"
    set -a
    # shellcheck disable=SC1090
    . "${env_file}"
    set +a
}

ensure_clean_worktree() {
    if [[ -n "$(git status --porcelain)" ]]; then
        fail "working tree is dirty; commit or stash changes before releasing"
    fi
}

detect_tag() {
    local tags=""
    local count=""

    tags="$(git tag --points-at HEAD --list 'v*' | sed '/^$/d')"
    count="$(printf '%s\n' "${tags}" | sed '/^$/d' | wc -l | tr -d ' ')"

    if [[ "${count}" -ne 1 ]]; then
        fail "expected exactly one v* tag on HEAD, found ${count}"
    fi

    printf '%s\n' "${tags}"
}

derive_project_url() {
    local remote_url=""
    local candidate=""

    if [[ -n "${GITLAB_PROJECT_URL:-}" ]]; then
        printf '%s\n' "${GITLAB_PROJECT_URL}"
        return 0
    fi
    if [[ -n "${CI_PROJECT_URL:-}" ]]; then
        printf '%s\n' "${CI_PROJECT_URL}"
        return 0
    fi

    for candidate in gitlab origin; do
        if remote_url="$(git remote get-url "${candidate}" 2>/dev/null)"; then
            case "${remote_url}" in
                git@gitlab.com:*.git)
                    printf 'https://gitlab.com/%s\n' "${remote_url#git@gitlab.com:}" | sed 's/\.git$//'
                    return 0
                    ;;
                https://gitlab.com/*.git)
                    printf '%s\n' "${remote_url%.git}"
                    return 0
                    ;;
                https://gitlab.com/*)
                    printf '%s\n' "${remote_url}"
                    return 0
                    ;;
            esac
        fi
    done

    fail "unable to derive GitLab project URL; set GITLAB_PROJECT_URL"
}

build_release_asset_url() {
    local download_base="$1"
    local tag="$2"
    local asset_name="$3"

    printf '%s/releases/%s/%s\n' "${download_base%/}" "${tag}" "${asset_name}"
}

build_release_asset_key() {
    local prefix="$1"
    local tag="$2"
    local asset_name="$3"

    storage_key "${prefix}" "releases/${tag}/${asset_name}"
}

wait_for_ci_assets() {
    local download_base="$1"
    local tag="$2"
    local asset_url=""

    for asset_url in \
        "$(build_release_asset_url "${download_base}" "${tag}" "${LINUX_CLI_ASSET}")" \
        "$(build_release_asset_url "${download_base}" "${tag}" "${LINUX_DESKTOP_ASSET}")" \
        "$(build_release_asset_url "${download_base}" "${tag}" "${LINUX_ASSET}")" \
        "$(build_release_asset_url "${download_base}" "${tag}" "${WINDOWS_ASSET}")"
    do
        wait_for_public_asset "${asset_url}" "required CI asset is not available yet"
    done
}

wait_for_public_asset() {
    local asset_url="$1"
    local failure_message="$2"
    local attempt=0

    for attempt in {1..12}; do
        if curl -fsSI "${asset_url}" >/dev/null; then
            return 0
        fi
        sleep 5
    done

    fail "${failure_message}: ${asset_url}"
}

ensure_public_download_base_matches_prefix() {
    local public_download_base="$1"
    local cdn_prefix="$2"

    [[ -z "${cdn_prefix}" ]] && return 0

    case "${public_download_base%/}" in
        */"${cdn_prefix}") ;;
        *)
            fail "PUBLIC_DOWNLOAD_BASE must end with /${cdn_prefix}. Current value: ${public_download_base}"
            ;;
    esac
}

configure_rclone() {
    : "${R2_ACCESS_KEY_ID:?R2_ACCESS_KEY_ID is required}"
    : "${R2_SECRET_ACCESS_KEY:?R2_SECRET_ACCESS_KEY is required}"
    : "${R2_ENDPOINT:?R2_ENDPOINT is required}"

    local config_file="$1"
    cat >"${config_file}" <<EOF
[r2]
type = s3
provider = Cloudflare
access_key_id = ${R2_ACCESS_KEY_ID}
secret_access_key = ${R2_SECRET_ACCESS_KEY}
endpoint = ${R2_ENDPOINT}
no_check_bucket = true
EOF
}

storage_key() {
    local prefix="$1"
    local path="$2"

    prefix="${prefix%/}"
    if [[ -z "${prefix}" ]]; then
        printf '%s\n' "${path}"
        return 0
    fi

    printf '%s/%s\n' "${prefix}" "${path}"
}

main() {
    local root_dir=""
    local tag=""
    local version=""
    local project_url=""
    local public_download_base="${PUBLIC_DOWNLOAD_BASE:-${DEFAULT_DOWNLOAD_BASE}}"
    local cdn_bucket="${CDN_BUCKET:-${DEFAULT_CDN_BUCKET}}"
    local cdn_prefix="${CDN_PREFIX:-${DEFAULT_CDN_PREFIX}}"
    local release_dir="build/release"
    local cli_asset_path="${release_dir}/${MACOS_CLI_ASSET}"
    local desktop_asset_path="${release_dir}/${MACOS_DESKTOP_ASSET}"
    local legacy_asset_path="${release_dir}/${LEGACY_MACOS_ASSET}"
    local macos_min_version="${MACOS_MIN_VERSION:-${DEFAULT_MACOS_MIN_VERSION}}"
    local latest_file=""
    local legacy_latest_file=""
    local rclone_config=""

    [[ "$(uname -s)" == "Darwin" ]] || fail "this script must run on macOS"

    require_command git
    require_command go
    require_command corepack
    require_command make
    require_command rclone
    require_command curl

    root_dir="$(repo_root)"
    cd "${root_dir}"
    load_dotenv "${root_dir}/.env"

    ensure_clean_worktree
    tag="$(detect_tag)"
    version="${tag#v}"
    project_url="$(derive_project_url)"

    public_download_base="${PUBLIC_DOWNLOAD_BASE:-${DEFAULT_DOWNLOAD_BASE}}"
    cdn_bucket="${CDN_BUCKET:-${DEFAULT_CDN_BUCKET}}"
    cdn_prefix="${CDN_PREFIX:-${DEFAULT_CDN_PREFIX}}"
    ensure_public_download_base_matches_prefix "${public_download_base}" "${cdn_prefix}"

    log "Releasing ${tag} from $(git rev-parse --short HEAD)"
    log "Building macOS CLI and desktop binaries with minimum deployment target ${macos_min_version}"

    corepack enable
    require_command pnpm
    pnpm --dir frontend install --frozen-lockfile --prefer-offline
    make build-frontend

    mkdir -p "${release_dir}"
    rm -f "${cli_asset_path}" "${desktop_asset_path}" "${legacy_asset_path}"

    CGO_ENABLED=0 \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
        -ldflags="-s -w -X gitlab.com/usescrolls/scribe/internal.Version=${version} -X gitlab.com/usescrolls/scribe/internal.PublicDownloadBase=${public_download_base}" \
        -o "${cli_asset_path}" \
        ./cmd/scribe

    CGO_ENABLED=1 \
    GOOS=darwin \
    GOARCH=arm64 \
    MACOSX_DEPLOYMENT_TARGET="${macos_min_version}" \
    CGO_CFLAGS="-mmacosx-version-min=${macos_min_version}" \
    CGO_LDFLAGS="-mmacosx-version-min=${macos_min_version}" \
    go build \
        -ldflags="-s -w -X gitlab.com/usescrolls/scribe/internal.Version=${version} -X gitlab.com/usescrolls/scribe/internal.PublicDownloadBase=${public_download_base}" \
        -o "${desktop_asset_path}" \
        .

    cp "${desktop_asset_path}" "${legacy_asset_path}"

    wait_for_ci_assets "${public_download_base}" "${tag}"

    latest_file="$(mktemp)"
    legacy_latest_file="$(mktemp)"
    rclone_config="$(mktemp)"
    trap "rm -f -- \"${latest_file}\" \"${legacy_latest_file}\" \"${rclone_config}\"" EXIT

    configure_rclone "${rclone_config}"
    export RCLONE_CONFIG="${rclone_config}"

    rclone copyto "${cli_asset_path}" "r2:${cdn_bucket}/$(build_release_asset_key "${cdn_prefix}" "${tag}" "${MACOS_CLI_ASSET}")"
    rclone copyto "${cli_asset_path}" "r2:${cdn_bucket}/$(storage_key "${cdn_prefix}" "${MACOS_CLI_ASSET}")"
    rclone copyto "${desktop_asset_path}" "r2:${cdn_bucket}/$(build_release_asset_key "${cdn_prefix}" "${tag}" "${MACOS_DESKTOP_ASSET}")"
    rclone copyto "${desktop_asset_path}" "r2:${cdn_bucket}/$(storage_key "${cdn_prefix}" "${MACOS_DESKTOP_ASSET}")"
    rclone copyto "${legacy_asset_path}" "r2:${cdn_bucket}/$(build_release_asset_key "${cdn_prefix}" "${tag}" "${LEGACY_MACOS_ASSET}")"
    rclone copyto "${legacy_asset_path}" "r2:${cdn_bucket}/$(storage_key "${cdn_prefix}" "${LEGACY_MACOS_ASSET}")"
    wait_for_public_asset "$(build_release_asset_url "${public_download_base}" "${tag}" "${MACOS_CLI_ASSET}")" "uploaded macOS CLI asset is not available yet"
    wait_for_public_asset "$(build_release_asset_url "${public_download_base}" "${tag}" "${MACOS_DESKTOP_ASSET}")" "uploaded macOS desktop asset is not available yet"

    cat >"${latest_file}" <<EOF
{
  "tag_name": "${tag}",
  "html_url": "${project_url}/-/releases/${tag}",
  "published_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "assets": [
    {"name": "${LINUX_CLI_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${LINUX_CLI_ASSET}")"},
    {"name": "${LINUX_DESKTOP_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${LINUX_DESKTOP_ASSET}")"},
    {"name": "${LINUX_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${LINUX_ASSET}")"},
    {"name": "${MACOS_CLI_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${MACOS_CLI_ASSET}")"},
    {"name": "${MACOS_DESKTOP_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${MACOS_DESKTOP_ASSET}")"},
    {"name": "${LEGACY_MACOS_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${LEGACY_MACOS_ASSET}")"},
    {"name": "${WINDOWS_ASSET}", "browser_download_url": "$(build_release_asset_url "${public_download_base}" "${tag}" "${WINDOWS_ASSET}")"}
  ]
}
EOF

    cp "${latest_file}" "${legacy_latest_file}"

    rclone copyto "${latest_file}" "r2:${cdn_bucket}/$(storage_key "${cdn_prefix}" "releases/latest")"
    # Keep the legacy manifest path alive so existing installs that still check
    # /releases/latest can discover the namespaced asset URLs.
    rclone copyto "${legacy_latest_file}" "r2:${cdn_bucket}/releases/latest"

    log "Uploaded macOS split assets to ${cdn_prefix}/releases/${tag}/ and refreshed ${cdn_prefix}/releases/latest for ${tag}"
}

main "$@"
