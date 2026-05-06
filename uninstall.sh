#!/usr/bin/env sh
set -eu

PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
BIN_PATH="$BIN_DIR/ytt"

run_privileged() {
	if "$@" 2>/dev/null; then
		return 0
	fi

	if command -v sudo >/dev/null 2>&1; then
		sudo "$@"
		return 0
	fi

	echo "ytt uninstall: failed to run: $*" >&2
	echo "Try setting BIN_DIR to the directory where ytt was installed." >&2
	exit 1
}

if [ ! -e "$BIN_PATH" ]; then
	echo "ytt is not installed at $BIN_PATH"
	exit 0
fi

run_privileged rm -f "$BIN_PATH"
echo "ytt removed from $BIN_PATH"
