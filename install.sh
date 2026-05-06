#!/usr/bin/env sh
set -eu

REPO_OWNER="${YTT_REPO_OWNER:-IFAKA}"
REPO_NAME="${YTT_REPO_NAME:-yt-transcription}"
REF="${YTT_REF:-main}"
PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
BIN_PATH="$BIN_DIR/ytt"
TMP_DIR="${TMPDIR:-/tmp}/ytt-install-$$"

cleanup() {
	rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "ytt install: missing required command: $1" >&2
		exit 1
	fi
}

run_privileged() {
	if "$@" 2>/dev/null; then
		return 0
	fi

	if command -v sudo >/dev/null 2>&1; then
		sudo "$@"
		return 0
	fi

	echo "ytt install: failed to run: $*" >&2
	echo "Try setting BIN_DIR to a writable directory, for example:" >&2
	echo "  BIN_DIR=\"\$HOME/.local/bin\" sh install.sh" >&2
	exit 1
}

need_cmd go

mkdir -p "$TMP_DIR"

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" 2>/dev/null && pwd || pwd)"
if [ -f "$SCRIPT_DIR/ytt.go" ] && [ -f "$SCRIPT_DIR/go.mod" ]; then
	SRC_DIR="$SCRIPT_DIR"
else
	need_cmd curl
	SRC_DIR="$TMP_DIR/src"
	mkdir -p "$SRC_DIR"
	RAW_URL="https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/$REF"
	curl -fsSL "$RAW_URL/ytt.go" -o "$SRC_DIR/ytt.go"
	curl -fsSL "$RAW_URL/go.mod" -o "$SRC_DIR/go.mod"
	curl -fsSL "$RAW_URL/go.sum" -o "$SRC_DIR/go.sum"
fi

(cd "$SRC_DIR" && go build -o "$TMP_DIR/ytt" ytt.go)

run_privileged install -d "$BIN_DIR"
run_privileged install -m 0755 "$TMP_DIR/ytt" "$BIN_PATH"

echo "ytt installed to $BIN_PATH"
echo "Run: ytt <youtube_url_or_id>"
