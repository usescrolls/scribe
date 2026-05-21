#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d)"
HOME_DIR="$TMP_DIR/home"
MOCK_BIN="$TMP_DIR/bin"

cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

mkdir -p "$HOME_DIR" "$MOCK_BIN"

cat > "$MOCK_BIN/uname" <<'EOF'
#!/usr/bin/env bash
case "${1:-}" in
    -s) echo Linux ;;
    -m) echo x86_64 ;;
    *)
        if [ -x /usr/bin/uname ]; then
            /usr/bin/uname "$@"
        else
            /bin/uname "$@"
        fi
        ;;
esac
EOF
chmod +x "$MOCK_BIN/uname"

cat > "$MOCK_BIN/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

output=""
url=""
while [ "$#" -gt 0 ]; do
    case "$1" in
        -o)
            shift
            output="$1"
            ;;
        -*)
            ;;
        *)
            url="$1"
            ;;
    esac
    shift
done

write_payload() {
    if [ -n "$output" ]; then
        printf '%s' "$1" > "$output"
    else
        printf '%s' "$1"
    fi
}

case "$url" in
    */releases/latest)
        write_payload '{"tag_name":"v9.9.9","html_url":"https://example.invalid/scribe/releases/v9.9.9","published_at":"2026-01-01T00:00:00Z","assets":[{"name":"scribe-cli-linux-amd64","browser_download_url":"https://example.invalid/scribe/releases/v9.9.9/scribe-cli-linux-amd64"},{"name":"scribe-desktop-linux-amd64","browser_download_url":"https://example.invalid/scribe/releases/v9.9.9/scribe-desktop-linux-amd64"}]}'
        ;;
    *scribe-cli-linux-amd64)
        write_payload '#!/usr/bin/env sh
if [ "${1:-}" = "version" ]; then
    echo "scribe version 9.9.9"
else
    echo "scribe cli"
fi
'
        ;;
    *scribe-desktop-linux-amd64)
        write_payload '#!/usr/bin/env sh
if [ "${1:-}" = "version" ]; then
    echo "scribe version 9.9.9"
else
    echo "scribe desktop"
fi
'
        ;;
    */icons/icon.png)
        write_payload 'mock icon'
        ;;
    *)
        echo "unexpected curl URL: $url" >&2
        exit 22
        ;;
esac
EOF
chmod +x "$MOCK_BIN/curl"

cat > "$MOCK_BIN/xdg-mime" <<'EOF'
#!/usr/bin/env sh
exit 0
EOF
chmod +x "$MOCK_BIN/xdg-mime"

cat > "$MOCK_BIN/update-desktop-database" <<'EOF'
#!/usr/bin/env sh
exit 0
EOF
chmod +x "$MOCK_BIN/update-desktop-database"

export HOME="$HOME_DIR"
export PATH="$MOCK_BIN:$PATH"
export SHELL=/bin/bash
export PUBLIC_DOWNLOAD_BASE=https://example.invalid/scribe
unset DISPLAY
unset WAYLAND_DISPLAY

bash "$REPO_ROOT/scripts/install.sh" > "$TMP_DIR/install.out"

assert_executable() {
    if [ ! -x "$1" ]; then
        echo "expected executable: $1" >&2
        exit 1
    fi
}

assert_file_contains() {
    local path="$1"
    local expected="$2"

    if ! grep -Fq "$expected" "$path"; then
        echo "expected $path to contain: $expected" >&2
        exit 1
    fi
}

assert_executable "$HOME_DIR/.local/bin/scribe"
assert_executable "$HOME_DIR/.local/lib/scribe/scribe-desktop"

assert_file_contains "$HOME_DIR/.local/share/applications/scribe.desktop" "Exec=$HOME_DIR/.local/lib/scribe/scribe-desktop %u"
assert_file_contains "$HOME_DIR/.config/autostart/scribe.desktop" "Exec=$HOME_DIR/.local/lib/scribe/scribe-desktop"

INSTALL_MANIFEST="$HOME_DIR/.scribe/install.json"
assert_file_contains "$INSTALL_MANIFEST" "\"version\": \"9.9.9\""
assert_file_contains "$INSTALL_MANIFEST" "\"cliPath\": \"$HOME_DIR/.local/bin/scribe\""
assert_file_contains "$INSTALL_MANIFEST" "\"desktopPath\": \"$HOME_DIR/.local/lib/scribe/scribe-desktop\""
assert_file_contains "$INSTALL_MANIFEST" "\"installedComponents\": [\"cli\", \"desktop\"]"
assert_file_contains "$INSTALL_MANIFEST" "\"publicDownloadBase\": \"https://example.invalid/scribe\""

assert_file_contains "$REPO_ROOT/scripts/install.sh" '<string>scribe-desktop</string>'
assert_file_contains "$REPO_ROOT/scripts/install.sh" 'APP_BINARY="$APP_BUNDLE/Contents/MacOS/$DESKTOP_BINARY_NAME"'
assert_file_contains "$REPO_ROOT/scripts/install.sh" 'pkill -x "$BINARY_NAME"'
assert_file_contains "$REPO_ROOT/scripts/install.sh" 'pkill -x "$DESKTOP_BINARY_NAME"'

echo "installer contract passed"
