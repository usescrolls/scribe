#!/usr/bin/env bash
set -euo pipefail

# Publish a final GitHub release from binaries already uploaded to the CDN.
# Usage:
#   ./scripts/release-github-bridge.sh v1.2.3
#   ./scripts/release-github-bridge.sh v1.2.3 owner/repo
#
# Required tools: git, gh, curl
# Required auth: `gh auth login`
# Optional env vars:
#   PUBLIC_DOWNLOAD_BASE / SCRIBE_DOWNLOAD_BASE
#   GITHUB_REPOSITORY
#   GITHUB_TARGET
#   GITHUB_RELEASE_TITLE
#   GITHUB_RELEASE_NOTES
#   GITHUB_RELEASE_NOTES_FILE
#   LINUX_ASSET / MACOS_ASSET / WINDOWS_ASSET
#
# If present, a repo-root .env file is loaded automatically.

DEFAULT_DOWNLOAD_BASE="https://cdn.usescrolls.com/scribe"
DEFAULT_LINUX_ASSET="scribe-linux-amd64"
DEFAULT_MACOS_ASSET="scribe-darwin-arm64"
DEFAULT_WINDOWS_ASSET="scribe-windows-amd64.exe"
DEFAULT_RELEASE_NOTES=$'Migration release.\n\nFuture updates are served from GitLab/CDN.'
TEMP_DIR=""

log() {
    printf '%s\n' "$*"
}

fail() {
    printf 'Error: %s\n' "$*" >&2
    exit 1
}

cleanup_temp_dir() {
    if [[ -n "${TEMP_DIR:-}" ]]; then
        rm -rf -- "${TEMP_DIR}"
    fi
}

usage() {
    cat <<'EOF'
Publish a GitHub release from binaries that already exist on the CDN.

Usage:
  ./scripts/release-github-bridge.sh <tag> [owner/repo]

Examples:
  ./scripts/release-github-bridge.sh v1.2.3
  ./scripts/release-github-bridge.sh 1.2.3 usescrolls/scribe

Environment overrides:
  PUBLIC_DOWNLOAD_BASE / SCRIBE_DOWNLOAD_BASE
  GITHUB_REPOSITORY
  GITHUB_TARGET
  GITHUB_RELEASE_TITLE
  GITHUB_RELEASE_NOTES
  GITHUB_RELEASE_NOTES_FILE
  LINUX_ASSET / MACOS_ASSET / WINDOWS_ASSET
EOF
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

normalize_tag() {
    local raw_tag="$1"

    [[ -n "${raw_tag}" ]] || fail "tag is required"
    case "${raw_tag}" in
        v*) printf '%s\n' "${raw_tag}" ;;
        *) printf 'v%s\n' "${raw_tag}" ;;
    esac
}

derive_github_repository() {
    local remote_url=""

    if [[ -n "${GITHUB_REPOSITORY:-}" ]]; then
        printf '%s\n' "${GITHUB_REPOSITORY}"
        return 0
    fi

    remote_url="$(git remote get-url origin 2>/dev/null || true)"
    [[ -n "${remote_url}" ]] || fail "unable to derive GitHub repository; pass owner/repo or set GITHUB_REPOSITORY"

    case "${remote_url}" in
        git@github.com:*.git)
            printf '%s\n' "${remote_url#git@github.com:}" | sed 's/\.git$//'
            ;;
        git@github.com:*)
            printf '%s\n' "${remote_url#git@github.com:}"
            ;;
        https://github.com/*.git)
            printf '%s\n' "${remote_url#https://github.com/}" | sed 's/\.git$//'
            ;;
        https://github.com/*)
            printf '%s\n' "${remote_url#https://github.com/}"
            ;;
        *)
            fail "origin is not a GitHub remote; pass owner/repo or set GITHUB_REPOSITORY"
            ;;
    esac
}

derive_target_commit() {
    local tag="$1"
    local target=""

    if [[ -n "${GITHUB_TARGET:-}" ]]; then
        printf '%s\n' "${GITHUB_TARGET}"
        return 0
    fi

    target="$(git rev-parse --verify "${tag}^{commit}" 2>/dev/null || true)"
    [[ -n "${target}" ]] || fail "tag ${tag} was not found locally; fetch tags or set GITHUB_TARGET"

    printf '%s\n' "${target}"
}

build_release_asset_url() {
    local download_base="$1"
    local tag="$2"
    local asset_name="$3"

    printf '%s/releases/%s/%s\n' "${download_base%/}" "${tag}" "${asset_name}"
}

download_asset() {
    local asset_url="$1"
    local destination="$2"

    log "Downloading ${asset_url}"
    curl -L --silent --show-error --fail "${asset_url}" -o "${destination}"
}

build_notes_file() {
    local notes_file="$1"
    local public_download_base="$2"

    if [[ -n "${GITHUB_RELEASE_NOTES_FILE:-}" ]]; then
        [[ -f "${GITHUB_RELEASE_NOTES_FILE}" ]] || fail "notes file does not exist: ${GITHUB_RELEASE_NOTES_FILE}"
        cp "${GITHUB_RELEASE_NOTES_FILE}" "${notes_file}"
        return 0
    fi

    cat >"${notes_file}" <<EOF
${GITHUB_RELEASE_NOTES:-${DEFAULT_RELEASE_NOTES}}

Release assets now live at:
${public_download_base%/}
EOF
}

main() {
    local root_dir=""
    local tag=""
    local repo=""
    local target_commit=""
    local public_download_base=""
    local linux_asset=""
    local macos_asset=""
    local windows_asset=""
    local title=""
    local notes_file=""
    local assets=()
    local asset_name=""

    if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
        usage
        return 0
    fi

    [[ $# -ge 1 && $# -le 2 ]] || {
        usage >&2
        exit 1
    }

    require_command git
    require_command gh
    require_command curl
    require_command mktemp

    root_dir="$(repo_root)"
    cd "${root_dir}"
    load_dotenv "${root_dir}/.env"

    gh auth status >/dev/null 2>&1 || fail "GitHub CLI is not authenticated; run 'gh auth login'"

    tag="$(normalize_tag "$1")"
    repo="${2:-$(derive_github_repository)}"
    public_download_base="${PUBLIC_DOWNLOAD_BASE:-${SCRIBE_DOWNLOAD_BASE:-${DEFAULT_DOWNLOAD_BASE}}}"
    linux_asset="${LINUX_ASSET:-${DEFAULT_LINUX_ASSET}}"
    macos_asset="${MACOS_ASSET:-${DEFAULT_MACOS_ASSET}}"
    windows_asset="${WINDOWS_ASSET:-${DEFAULT_WINDOWS_ASSET}}"
    title="${GITHUB_RELEASE_TITLE:-${tag}}"

    TEMP_DIR="$(mktemp -d)"
    trap cleanup_temp_dir EXIT

    notes_file="${TEMP_DIR}/release-notes.md"
    build_notes_file "${notes_file}" "${public_download_base}"

    for asset_name in "${linux_asset}" "${macos_asset}" "${windows_asset}"; do
        download_asset \
            "$(build_release_asset_url "${public_download_base}" "${tag}" "${asset_name}")" \
            "${TEMP_DIR}/${asset_name}"
        assets+=("${TEMP_DIR}/${asset_name}")
    done

    if gh release view "${tag}" --repo "${repo}" >/dev/null 2>&1; then
        log "Updating GitHub release ${repo}@${tag}"
        gh release upload "${tag}" "${assets[@]}" --repo "${repo}" --clobber
        gh release edit "${tag}" \
            --repo "${repo}" \
            --title "${title}" \
            --notes-file "${notes_file}" \
            --draft=false \
            --prerelease=false \
            --latest
    else
        target_commit="$(derive_target_commit "${tag}")"
        log "Creating GitHub release ${repo}@${tag}"
        gh release create "${tag}" "${assets[@]}" \
            --repo "${repo}" \
            --target "${target_commit}" \
            --title "${title}" \
            --notes-file "${notes_file}" \
            --latest
    fi

    log "GitHub release ${repo}@${tag} is ready"
}

main "$@"
