#!/usr/bin/env sh
set -eu

arg_tools="${1:-}"
arg_ref="${2:-}"

get_latest_tag() {
    latest=$(curl -fsSL --max-time 10 -H "User-Agent: mise-seq-installer" \
        "https://api.github.com/repos/hiono/mise-seq/releases/latest" 2>/dev/null | \
        grep '"tag_name"' | sed 's/.*"\([^"]*\)".*/\1/')
    if [ -z "$latest" ]; then
        echo "WARNING: Could not fetch latest tag, using v0.1.0" >&2
        echo "v0.1.0"
    else
        echo "$latest"
    fi
}

get_commit_sha() {
    local ref="$1"
    curl -fsSL --max-time 10 -H "User-Agent: mise-seq-installer" \
        "https://api.github.com/repos/hiono/mise-seq/contents/install.sh?ref=$ref" 2>/dev/null | \
        grep -o '"sha":"[^"]*"' | sed 's/.*"\([^"]*\)".*/\1/' | cut -c1-7
}

detect_ref_from_url() {
    local url="$1"
    echo "$url" | sed -E 's|https://raw.githubusercontent.com/hiono/mise-seq/([^/]+)/.*|\1|'
}

REPO_RAW_BASE_DEFAULT="https://raw.githubusercontent.com/hiono/mise-seq/$(get_latest_tag)"

if [ -z "$arg_ref" ] && [ -n "${REPO_RAW_BASE:-}" ]; then
    :
elif [ -z "$arg_ref" ]; then
    arg_ref="main"
fi

case "$arg_ref" in
    main) REPO_RAW_BASE="https://raw.githubusercontent.com/hiono/mise-seq/main" ;;
    v[0-9]*) REPO_RAW_BASE="https://raw.githubusercontent.com/hiono/mise-seq/$arg_ref" ;;
    https://raw.githubusercontent.com/*)
        ref="$(detect_ref_from_url "$arg_ref")"
        sha="$(get_commit_sha "$ref")"
        if [ -n "$sha" ]; then
            REPO_RAW_BASE="https://raw.githubusercontent.com/hiono/mise-seq/$sha"
        else
            REPO_RAW_BASE="$(echo "$arg_ref" | sed -E 's|/blob/[^/]+/|/|; s|/tree/[^/]+/|/|; s|/[^/]+$||')"
        fi
        ;;
    *)    REPO_RAW_BASE="${REPO_RAW_BASE:-$REPO_RAW_BASE_DEFAULT}" ;;
esac

case "$REPO_RAW_BASE" in
    https://raw.githubusercontent.com/hiono/mise-seq/*) ;;
    *) echo "ERROR: Invalid repository URL: $REPO_RAW_BASE" >&2; exit 1 ;;
esac

TOOLS_DIR="${TOOLS_DIR:-$HOME/.tools}"

is_url() {
	case "$1" in
	http://* | https://*) return 0 ;;
	*) return 1 ;;
	esac
}

download() {
	url="$1"
	out="$2"
	mkdir -p "$(dirname "$out")"
	echo "Downloading: $url -> $out" >&2
	curl -fsSL --fail "$url" -o "$out" || {
		echo "ERROR: failed to download: $url" >&2
		exit 1
	}
}

if ! command -v mise >/dev/null 2>&1; then
	echo "ERROR: mise is required and must be installed system-wide (e.g. via apt)." >&2
	exit 1
fi

mkdir -p "$TOOLS_DIR" "$TOOLS_DIR/schema"

TMPDIR="${TMPDIR:-/tmp}"
T="$TMPDIR/mise-seq.$$"
mkdir -p "$T"

# Runtime files (from repo)
for f in \
	.tools/mise-seq.sh \
	.tools/tools.yaml \
	.tools/tools.toml \
	.tools/schema/mise-seq.cue; do
	download "$REPO_RAW_BASE/$f" "$T/$(basename "$f")"
done

chmod +x "$T/mise-seq.sh"

# tools.yaml selection
# Usage: curl .../install.sh [url|absolute-path]
# If omitted, uses bundled tools.yaml from repo
if [ -n "$arg_tools" ]; then
	if is_url "$arg_tools"; then
		download "$arg_tools" "$T/tools.yaml"
	else
		# Require absolute path for local files
		case "$arg_tools" in
		/*) [ -f "$arg_tools" ] && cp "$arg_tools" "$T/tools.yaml" ;;
		*)
			echo "ERROR: Local file requires absolute path: $arg_tools" >&2
			echo "Usage: curl .../install.sh https://example.com/tools.yaml" >&2
			echo "   or: curl .../install.sh /absolute/path/to/tools.yaml" >&2
			exit 1
			;;
		esac
	fi
fi

cp "$T/mise-seq.sh" "$TOOLS_DIR/mise-seq.sh"
cp "$T/tools.yaml" "$TOOLS_DIR/tools.yaml"
cp "$T/tools.toml" "$TOOLS_DIR/tools.toml" 2>/dev/null || true
cp "$T/mise-seq.cue" "$TOOLS_DIR/schema/mise-seq.cue"

echo "Installed mise-seq runtime to: $TOOLS_DIR"
export DEBUG="${DEBUG:-0}"
exec "$TOOLS_DIR/mise-seq.sh"
