#!/usr/bin/env sh
set -eu

REPO_RAW_BASE_DEFAULT="https://raw.githubusercontent.com/hiono/mise-seq/main"

TOOLS_DIR="${TOOLS_DIR:-$HOME/.tools}"
REPO_RAW_BASE="${REPO_RAW_BASE:-$REPO_RAW_BASE_DEFAULT}"

arg_tools="${1:-}"

is_url() {
  case "$1" in
    http://*|https://*) return 0 ;;
    *) return 1 ;;
  esac
}

download() {
  url="$1"
  out="$2"
  curl -fsSL "$url" > "$out"
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
  .tools/install.sh \
  .tools/tools.yaml \
  .tools/tools.sample.toml \
  .tools/schema/mise-seq.cue

do
  download "$REPO_RAW_BASE/$f" "$T/$(basename "$f")"
done

chmod +x "$T/install.sh"

# tools.yaml selection
if [ -n "$arg_tools" ]; then
  if is_url "$arg_tools"; then
    download "$arg_tools" "$T/tools.yaml"
  else
    [ -f "$arg_tools" ] || { echo "ERROR: tools.yaml not found: $arg_tools" >&2; exit 1; }
    cp "$arg_tools" "$T/tools.yaml"
  fi
fi

cp "$T/install.sh" "$TOOLS_DIR/install.sh"
cp "$T/tools.yaml" "$TOOLS_DIR/tools.yaml"
cp "$T/tools.sample.toml" "$TOOLS_DIR/tools.sample.toml" 2>/dev/null || true
cp "$T/mise-seq.cue" "$TOOLS_DIR/schema/mise-seq.cue"

echo "Installed mise-seq runtime to: $TOOLS_DIR"
exec "$TOOLS_DIR/install.sh"
