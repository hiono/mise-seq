#!/usr/bin/env bash
set -Eeuo pipefail
echo "MISE-SEQ SCRIPT STARTED" >&2
echo "DEBUG=$DEBUG" >&2
echo "Line count: $(wc -l < "$0")" >&2

TOOLS_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CFG="${TOOLS_DIR}/tools.yaml"
SCHEMA_CUE="${TOOLS_DIR}/schema/mise-seq.cue"

DRY_RUN="${DRY_RUN:-0}"
DEBUG="${DEBUG:-0}"
RUN_POSTINSTALL_ON_UPDATE="${RUN_POSTINSTALL_ON_UPDATE:-${RUN_SETUP_ON_UPDATE:-0}}"
FORCE_HOOKS="${FORCE_HOOKS:-${FORCE_SETUP:-0}}"

CUE_VERSION="${CUE_VERSION:-latest}"

STATE_DIR="${STATE_DIR:-${XDG_CACHE_HOME:-$HOME/.cache}/tools/state}"

log_info() { echo "[INFO] $*"; }
log_error() { echo "[ERROR] $*" >&2; }
log_debug() { [[ "$DEBUG" == "1" ]] && echo "[DEBUG] $*" || true; }

run() {
	if [[ "$DRY_RUN" == "1" ]]; then
		echo "[DRY_RUN] $*"
		return 0
	fi
	"$@"
}

die() {
	echo "ERROR: $*" >&2
	exit 1
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "Required command not found: $1"
}

require_cmd mise
log_debug "Checking mise command..."
mkdir -p "$STATE_DIR"

MISE_SHIMS_DEFAULT="${HOME}/.local/share/mise/shims"
MISE_DATA_DIR="${MISE_DATA_DIR:-$HOME/.local/share/mise}"
MISE_SHIMS_CUSTOM="${MISE_DATA_DIR}/shims"

for shims_dir in "$MISE_SHIMS_CUSTOM" "$MISE_SHIMS_DEFAULT"; do
	if [ -d "$shims_dir" ]; then
		export PATH="$shims_dir:$PATH"
	fi
done

MISE_BIN="${HOME}/.local/bin"
if [ -d "$MISE_BIN" ]; then
	export PATH="$MISE_BIN:$PATH"
fi

# Bootstrap validators using mise (aqua)
# Note: yq is no longer required for config parsing (using cue + jq instead)

apply_mise_settings() {
	log_info "Applying mise settings from config..."
	
	# Use cue to extract settings and apply them directly
	local npm_pkg
	npm_pkg="$($CUE export "$CFG" --out json 2>/dev/null | grep -o '"npm":{[^}]*}' | grep -o '"package_manager":"[^"]*"' | cut -d'"' -f4 || echo "")"
	
	if [ -n "$npm_pkg" ]; then
		log_info "Setting npm.package_manager = $npm_pkg"
		mise settings set "npm.package_manager" "$npm_pkg" 2>/dev/null || true
	fi

	local experimental
	experimental="$($CUE export "$CFG" --out json 2>/dev/null | grep -o '"experimental":[^,}]*' | cut -d':' -f2 | tr -d ' ' || echo "")"
	
	if [ -n "$experimental" ] && [ "$experimental" != "null" ]; then
		log_info "Setting experimental = $experimental"
		mise settings set "experimental" "$experimental" 2>/dev/null || true
	fi
}

import_tools_to_mise() {
	log_info "Importing tools from config to mise..."
	
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	
	if [ -z "$cfg" ]; then
		log_debug "No config exported"
		return
	fi

	log_info "Tools found, adding to mise..."
	
	# Extract tools using grep - find all tool entries with version
	echo "$cfg" | grep -oE '"[a-zA-Z0-9:_/-]+":\s*\{[^}]*"version":\s*"[^"]*"' | while read -r line; do
		tool="$(echo "$line" | sed 's/:.*//' | tr -d '"')"
		version="$(echo "$line" | grep -o '"version":\s*"[^"]*"' | cut -d'"' -f4)"
		if [ -n "$tool" ] && [ -n "$version" ]; then
			log_info "Adding tool: $tool@$version"
			mise add "$tool@$version" 2>/dev/null || true
		fi
	done
	
	# Also try running mise install for any local mise configs
	log_info "Running mise install..."
	mise install 2>/dev/null || true
}

echo "=== Starting mise-seq ===" >&2
echo "=== Checking for cue ===" >&2

if ! command -v cue >/dev/null 2>&1; then
	echo "=== Running mise install ===" >&2
	log_info "Running mise install to install tools from config..."
	run mise install >/dev/null 2>&1 || true

	echo "=== Installing cue ===" >&2
	log_info "Installing bootstrap: cue@${CUE_VERSION}"
	run mise use -g "cue@${CUE_VERSION}" >/dev/null
fi

echo "=== After cue check ===" >&2
require_cmd cue
log_debug "CUE command found: $(command -v cue)"
CUE="$(command -v cue)"

log_debug "Using config: $CFG"
log_debug "Using schema: $SCHEMA_CUE"
log_debug "State directory: $STATE_DIR"
log_debug "DEBUG variable: $DEBUG"

log_debug "Checking config file..."
[[ -f "$CFG" ]] || die "Config not found: $CFG"

log_debug "Config file exists at: $CFG"
log_debug "Calling cfg_json..."
cue_output="$($CUE export "$CFG" --out json 2>&1)" || true
log_debug "cue output: ${cue_output:0:100}..."
test_json="$(echo "$cue_output" 2>/dev/null || echo '{}')"
log_debug "cfg_json returned: ${test_json:0:100}..."

# cfg_json function - returns cached JSON
cfg_json() {
	echo "$test_json"
}

# Now apply settings and import tools AFTER config is loaded
echo "=== Applying mise settings ===" >&2
apply_mise_settings

echo "=== Importing tools to mise ===" >&2
import_tools_to_mise

# Get tools_order array
get_tools_order() {
	log_debug "get_tools_order called"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	# Extract tools_order from JSON using grep/sed
	order="$(echo "$cfg" | grep -oE '"tools_order":\s*\[[^\]]*\]' | sed 's/.*\[//' | tr -d '[]"\n' | tr ',' '\n' | grep -v '^$')"
	log_debug "tools_order: $order"
	echo "$order"
}

# Check if tool exists
tool_exists() {
	local tool="$1"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	echo "$cfg" | grep -q "\"$tool\":" && echo "true" || echo "false"
}

# Get tool version
get_tool_version() {
	local tool="$1"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	version="$(echo "$cfg" | grep -oE "\"$tool\":\s*\{[^}]*\"version\":\s*\"[^\"]*\"" | grep -oE '"version":\s*"[^"]*"' | cut -d'"' -f4 | head -1)"
	echo "${version:-latest}"
}

# Get hook list for a tool (preinstall/postinstall)
get_hooks() {
	local tool="$1" phase="$2"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	# Extract hooks using grep - find tool section, then extract phase array
	echo "$cfg" | grep -oE "\"$tool\":\s*\{[^}]*\"$phase\":\s*\[[^\]]*\]" | sed "s/.*\"$phase\":\s*\[//" | tr -d '[]"' | tr ',' '\n'
}

# Get defaults hooks
get_default_hooks() {
	local phase="$1"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	echo "$cfg" | grep -oE "\"defaults\":\s*\{[^}]*\"$phase\":\s*\[[^\]]*\]" | sed "s/.*\"$phase\":\s*\[//" | tr -d '[]"' | tr ',' '\n'
}

# Get hook list for a tool (preinstall/postinstall)
get_hooks() {
	local tool="$1" phase="$2"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	echo "$cfg" | grep -oE "\"$tool\":\s*{[^}]*\"$phase\":\s*\[[^\]]*\]" | sed "s/.*\"$phase\":\s*\[//" | tr -d '[]"' | tr ',' '\n'
}

# Get defaults hooks
get_default_hooks() {
	local phase="$1"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	echo "$cfg" | grep -oE "\"defaults\":\s*{[^}]*\"$phase\":\s*\[[^\]]*\]" | sed "s/.*\"$phase\":\s*\[//" | tr -d '[]"' | tr ',' '\n'
}

is_managed_by_mise() {
	local tool="$1"
	mise which "$tool" >/dev/null 2>&1
}

marker_path() {
	local key="$1"
	printf '%s/%s.sha256' "$STATE_DIR" "$key"
}

read_marker() {
	local p="$1"
	[[ -f "$p" ]] && cat "$p" || true
}

write_marker() {
	local p="$1" hash="$2"
	if [[ "$DRY_RUN" == "1" ]]; then
		echo "[DRY_RUN] write marker $p ($hash)"
		return 0
	fi
	printf '%s\n' "$hash" >"$p"
}

sanitize_id() {
	local s="$1"
	echo "$s" | tr -cd '[:alnum:]_-' | cut -c1-64
}

# Get hook list for a tool (preinstall/postinstall)
get_tool_hooks() {
	local tool="$1" hook="$2"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	# Extract hooks from tool section using grep
	echo "$cfg" | grep -oE "\"$tool\":\s*\{[^}]*\"$hook\":\s*\[[^\]]*\]" | sed "s/.*\"$hook\":\s*\[//" | tr -d '[]"' | tr ',' '\n'
}

# Get defaults hooks
get_defaults_hooks() {
	local hook="$1"
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	# Extract hooks from defaults section using grep
	echo "$cfg" | grep -oE "\"defaults\":\s*\{[^}]*\"$hook\":\s*\[[^\]]*\]" | sed "s/.*\"$hook\":\s*\[//" | tr -d '[]"' | tr ',' '\n'
}

hook_hash_from_json() {
	local json="$1"
	# Simple hash without jq
	echo "$json" | tr -d '\n' | sha256sum | cut -d' ' -f1
}

run_defaults() {
	local hook="$1"
	local key="defaults.${hook}"
	local label="defaults ${hook}"

	# Get defaults hooks using JSON
	local hooks_json
	hooks_json="$(get_defaults_hooks "$hook")"

	local len
	len="$(echo "$hooks_json" | grep -c 'run')" || len=0
	[[ "$len" == "0" ]] && {
		log_debug "Hook ${label}: no scripts defined (skipped)"
		return 0
	}

	local mpath cur_hash old_hash
	mpath="$(marker_path "$key")"
	cur_hash="$(hook_hash_from_json "$hooks_json")"
	old_hash="$(read_marker "$mpath")"

	local need=0
	if [[ "$FORCE_HOOKS" == "1" ]]; then
		need=1
		log_debug "Hook ${label}: FORCE_HOOKS=1 (running)"
	elif [[ -z "$old_hash" ]]; then
		need=1
		log_debug "Hook ${label}: first run (running)"
	elif [[ "$old_hash" != "$cur_hash" ]]; then
		need=1
		log_debug "Hook ${label}: hash changed (running)"
	else
		log_debug "Hook ${label}: unchanged (skipped)"
	fi

	[[ "$need" == "0" ]] && return 0

	log_info "Running hook: ${label}"
	local script
	# Extract run scripts from hooks_json - each block is separated by "run":"
	while echo "$hooks_json" | grep -q '"run"'; do
		script="$(echo "$hooks_json" | grep -oE '"run":\s*"[^"]*"' | head -1 | sed 's/.*"run":\s*"//;s/"$//' | sed 's/\\n/\n/g')"
		[[ -z "$script" ]] && break
		run_script_block "$label" "$script"
		# Remove processed script
		hooks_json="$(echo "$hooks_json" | sed 's/.*"run":\s*"[^"]*"//')"
	done

	write_marker "$mpath" "$cur_hash"
}

run_tool_hook() {
	local tool="$1" hook="$2" phase="$3"
	local safe
	safe="$(sanitize_id "$tool")"
	local key="tool.${safe}.${hook}"
	local label="${tool} ${hook} (${phase})"

	# Get hooks using grep on JSON from cue export
	local hooks_json
	hooks_json="$(get_hooks "$tool" "$hook")"

	local len
	len="$(echo "$hooks_json" | grep -c 'run')" || len=0
	[[ "$len" == "0" ]] && {
		log_debug "Hook ${label}: no scripts defined (skipped)"
		return 0
	}

	local mpath cur_hash old_hash
	mpath="$(marker_path "$key")"
	cur_hash="$(hook_hash_from_json "$hooks_json")"
	old_hash="$(read_marker "$mpath")"

	local need=0
	if [[ "$FORCE_HOOKS" == "1" ]]; then
		need=1
		log_debug "Hook ${label}: FORCE_HOOKS=1 (running)"
	elif [[ -z "$old_hash" ]]; then
		need=1
		log_debug "Hook ${label}: first run (running)"
	elif [[ "$old_hash" != "$cur_hash" ]]; then
		need=1
		log_debug "Hook ${label}: hash changed (running)"
	else
		log_debug "Hook ${label}: unchanged (skipped)"
	fi

	[[ "$need" == "0" ]] && return 0

	log_info "Running hook: ${label}"
	local script
	# Extract run scripts from hooks_json - each block is separated by "run":"
	while echo "$hooks_json" | grep -q '"run"'; do
		script="$(echo "$hooks_json" | grep -oE '"run":\s*"[^"]*"' | head -1 | sed 's/.*"run":\s*"//;s/"$//' | sed 's/\\n/\n/g')"
		[[ -z "$script" ]] && break
		run_script_block "$label" "$script"
		# Remove processed script
		hooks_json="$(echo "$hooks_json" | sed 's/.*"run":\s*"[^"]*"//')"
	done

	write_marker "$mpath" "$cur_hash"
}

read_tool_order() {
	# Check if tools_order exists, otherwise get all tool keys
	local cfg
	cfg="$($CUE export "$CFG" --out json 2>/dev/null)"
	
	# Try tools_order first
	local order
	order="$(echo "$cfg" | grep -oE '"tools_order":\s*\[[^\]]*\]' | sed 's/.*\[//' | tr -d '[]"\n' | tr ',' '\n')"
	
	if [[ -n "$order" ]]; then
		echo "$order"
	else
		# Get all tool keys from tools section
		echo "$cfg" | grep -oE '"[a-zA-Z0-9:_/-]+":\s*\{' | sed 's/.*"\([^"]*\)".*/\1/'
	fi
}

run_defaults preinstall
log_debug "After run_defaults preinstall"

log_debug "About to call read_tool_order..."
mapfile -t TOOL_NAMES < <(read_tool_order)
log_debug "After read_tool_order: ${#TOOL_NAMES[@]} tools"
log_debug "Tools to process: ${TOOL_NAMES[*]}"

echo "DEBUG: TOOL_NAMES count = ${#TOOL_NAMES[@]}"
echo "DEBUG: TOOL_NAMES = ${TOOL_NAMES[*]}"

if [ ${#TOOL_NAMES[@]} -eq 0 ]; then
    log_error "No tools to process!"
    exit 1
fi

for tool in "${TOOL_NAMES[@]}"; do
	[[ -z "${tool//[[:space:]]/}" ]] && continue

	log_debug "Checking tool: $tool"
	exists="$(tool_exists "$tool")"
	log_debug "tool_exists('$tool') = $exists"
	[[ "$exists" != "true" ]] && { log_debug "Skipping $tool (already installed)"; continue; }

	ver="$(get_tool_version "$tool")"
	spec="${tool}@${ver}"

	log_debug "Processing tool: $spec"

	if ! is_managed_by_mise "$tool"; then
		run_tool_hook "$tool" preinstall install

		log_info "Installing: ${spec}"
		log_debug "Running: mise use -g $spec"
		if ! run mise use -g "$spec" 2>&1 | tee >(cat >&2); then
			log_error "Failed to install: ${spec} (postinstall skipped)"
			continue
		fi

		run_tool_hook "$tool" postinstall install
	else
		log_info "Already managed by mise: ${tool} (skipping)"
	fi
done

run_defaults postinstall

log_info "=== mise-seq installation complete ==="
log_debug "Done!"
