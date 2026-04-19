#!/usr/bin/env bash
set -euo pipefail

# Build and publish the macOS arm64 release asset from a tagged macOS checkout.
# Required env vars: R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_ENDPOINT.
# Optional env vars: PUBLIC_DOWNLOAD_BASE, CDN_BUCKET, GITLAB_PROJECT_URL.
# If present, a repo-root .env file is loaded automatically.

DEFAULT_DOWNLOAD_BASE="https://cdn.usescrolls.com"
DEFAULT_CDN_BUCKET="usescrolls-cdn"
ASSET_NAME="scribe-darwin-arm64"
WINDOWS_ASSET="scribe-windows-amd64.exe"
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
    mapfile -t tags < <(git tag --points-at HEAD --list 'v*')
    if [[ "${#tags[@]}" -ne 1 ]]; then
        fail "expected exactly one v* tag on HEAD, found ${#tags[@]}"
    fi
    printf '%s\n' "${tags[0]}"
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

wait_for_ci_assets() {
    local download_base="$1"
    local asset_url=""

    for asset_url in \
        "${download_base}/${LINUX_ASSET}" \
        "${download_base}/${WINDOWS_ASSET}"
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

main() {
    local root_dir=""
    local tag=""
    local version=""
    local project_url=""
    local public_download_base="${PUBLIC_DOWNLOAD_BASE:-${DEFAULT_DOWNLOAD_BASE}}"
    local cdn_bucket="${CDN_BUCKET:-${DEFAULT_CDN_BUCKET}}"
    local release_dir="build/release"
    local asset_path="${release_dir}/${ASSET_NAME}"
    local macos_version=""
    local latest_file=""
    local rclone_config=""

    [[ "$(uname -s)" == "Darwin" ]] || fail "this script must run on macOS"

    require_command git
    require_command go
    require_command corepack
    require_command make
    require_command rclone
    require_command curl
    require_command sw_vers

    root_dir="$(repo_root)"
    cd "${root_dir}"
    load_dotenv "${root_dir}/.env"

    ensure_clean_worktree
    tag="$(detect_tag)"
    version="${tag#v}"
    project_url="$(derive_project_url)"
    macos_version="$(sw_vers -productVersion)"

    public_download_base="${PUBLIC_DOWNLOAD_BASE:-${DEFAULT_DOWNLOAD_BASE}}"
    cdn_bucket="${CDN_BUCKET:-${DEFAULT_CDN_BUCKET}}"

    log "Releasing ${tag} from $(git rev-parse --short HEAD)"

    corepack enable
    require_command pnpm
    pnpm --dir frontend install --frozen-lockfile --prefer-offline
    make build-frontend

    mkdir -p "${release_dir}"
    rm -f "${asset_path}"

    CGO_ENABLED=1 \
    GOOS=darwin \
    GOARCH=arm64 \
    MACOSX_DEPLOYMENT_TARGET="${macos_version}" \
    CGO_CFLAGS="-mmacosx-version-min=${macos_version}" \
    CGO_LDFLAGS="-mmacosx-version-min=${macos_version}" \
    go build \
        -ldflags="-s -w -X gitlab.com/usescrolls/scribe/internal.Version=${version} -X gitlab.com/usescrolls/scribe/internal.PublicDownloadBase=${public_download_base}" \
        -o "${asset_path}" \
        .

    wait_for_ci_assets "${public_download_base}"

    latest_file="$(mktemp)"
    rclone_config="$(mktemp)"
    trap 'rm -f "${latest_file}" "${rclone_config}"' EXIT

    configure_rclone "${rclone_config}"
    export RCLONE_CONFIG="${rclone_config}"

    rclone copyto "${asset_path}" "r2:${cdn_bucket}/${ASSET_NAME}"
    wait_for_public_asset "${public_download_base}/${ASSET_NAME}" "uploaded macOS asset is not available yet"

    cat >"${latest_file}" <<EOF
{
  "tag_name": "${tag}",
  "html_url": "${project_url}/-/releases/${tag}",
  "published_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "assets": [
    {"name": "${LINUX_ASSET}", "browser_download_url": "${public_download_base}/${LINUX_ASSET}"},
    {"name": "${ASSET_NAME}", "browser_download_url": "${public_download_base}/${ASSET_NAME}"},
    {"name": "${WINDOWS_ASSET}", "browser_download_url": "${public_download_base}/${WINDOWS_ASSET}"}
  ]
}
EOF

    rclone copyto "${latest_file}" "r2:${cdn_bucket}/releases/latest"

    log "Uploaded ${ASSET_NAME} and refreshed releases/latest for ${tag}"
}

main "$@"
