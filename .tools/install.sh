#!/usr/bin/env bash
set -Eeuo pipefail

TOOLS_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CFG="${TOOLS_DIR}/tools.yaml"
SCHEMA_CUE="${TOOLS_DIR}/schema/mise-seq.cue"

DRY_RUN="${DRY_RUN:-0}"
RUN_POSTINSTALL_ON_UPDATE="${RUN_POSTINSTALL_ON_UPDATE:-${RUN_SETUP_ON_UPDATE:-0}}"
FORCE_HOOKS="${FORCE_HOOKS:-${FORCE_SETUP:-0}}"

YQ_VERSION="${YQ_VERSION:-4.44.3}"
CUE_VERSION="${CUE_VERSION:-latest}"

STATE_DIR="${STATE_DIR:-${XDG_CACHE_HOME:-$HOME/.cache}/tools/state}"

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
mkdir -p "$STATE_DIR"

# Ensure mise shims are in PATH (check both default and custom MISE_DATA_DIR locations)
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

# Bootstrap validators (user-scoped under mise)
run mise use -g "yq@${YQ_VERSION}" >/dev/null
run mise use -g "cue@${CUE_VERSION}" >/dev/null

# Re-export PATH after mise use to pick up new shims
for shims_dir in "$MISE_SHIMS_CUSTOM" "$MISE_SHIMS_DEFAULT"; do
	if [ -d "$shims_dir" ]; then
		export PATH="$shims_dir:$PATH"
	fi
done

require_cmd yq
require_cmd cue
YQ="$(command -v yq)"

[[ -f "$CFG" ]] || die "Config not found: $CFG"

# 1) YAML parse sanity
$YQ -e '.' "$CFG" >/dev/null

# 2) CUE schema validation
if [[ -f "$SCHEMA_CUE" ]]; then
	cue vet -c=false "$SCHEMA_CUE" "$CFG" -d '#MiseSeqConfig'
fi

sanitize_id() {
	echo "$1" | tr '/:' '__'
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

hook_hash() {
	local expr="$1"
	local json
	json="$($YQ -o=json -I=0 "$expr" "$CFG")"
	printf '%s' "$json" | sha256sum | awk '{print $1}'
}

# when must be a list in the schema. If omitted -> always.
select_scripts() {
	local expr="$1" phase="$2"
	$YQ -r --arg phase "$phase" '
    (('"$expr"') // [])[]
    | (.when // ["always"]) as $when
    | select(($when | index("always")) != null or ($when | index($phase)) != null)
    | (.run // "")
    | select(gsub("\\s+"; "") | length > 0)
    | . + "\u0000"
  ' "$CFG" || true
}

run_script_block() {
	local label="$1" script="$2"
	echo ">> ${label}"
	if [[ "$DRY_RUN" == "1" ]]; then
		echo "[DRY_RUN] sh -c <script>"
		echo "----------"
		echo "$script"
		echo "----------"
		return 0
	fi
	# Run as POSIX sh (as-is). No mise templating.
	sh -c "$script"
}

run_hook_if_needed() {
	local key="$1" expr="$2" phase="$3" label="$4"

	local len
	len="$($YQ -r "$expr | length // 0" "$CFG")"
	[[ "$len" == "0" ]] && return 0

	local mpath cur_hash old_hash
	mpath="$(marker_path "$key")"
	cur_hash="$(hook_hash "$expr")"
	old_hash="$(read_marker "$mpath")"

	local need=0
	if [[ "$FORCE_HOOKS" == "1" ]]; then
		need=1
	elif [[ -z "$old_hash" ]]; then
		need=1
	elif [[ "$old_hash" != "$cur_hash" ]]; then
		need=1
	fi

	[[ "$need" == "0" ]] && return 0

	local scripts
	scripts="$(select_scripts "$expr" "$phase")"
	if [[ -n "$scripts" ]]; then
		while IFS= read -r -d '' s; do
			run_script_block "$label" "$s"
		done < <(printf '%s' "$scripts")
	fi

	write_marker "$mpath" "$cur_hash"
}

run_defaults() {
	local hook="$1"
	local key="defaults.${hook}"
	local expr=".defaults.${hook} // []"
	local label="defaults ${hook}"
	run_hook_if_needed "$key" "$expr" "always" "$label"
}

run_tool_hook() {
	local tool="$1" hook="$2" phase="$3"
	local safe
	safe="$(sanitize_id "$tool")"
	local key="tool.${safe}.${hook}"
	local expr=".tools[\"${tool}\"].${hook} // []"
	local label="${tool} ${hook} (${phase})"
	run_hook_if_needed "$key" "$expr" "$phase" "$label"
}

read_tool_order() {
	local has
	has="$($YQ -r 'has("tools_order")' "$CFG" 2>/dev/null || echo false)"
	if [[ "$has" == "true" ]]; then
		mapfile -t ORDER < <($YQ -r '.tools_order[]' "$CFG")
		printf '%s\n' "${ORDER[@]}"
	else
		mapfile -t KEYS < <($YQ -r '.tools | keys | .[]' "$CFG")
		printf '%s\n' "${KEYS[@]}"
	fi
}

run_defaults preinstall

mapfile -t TOOL_NAMES < <(read_tool_order)

for tool in "${TOOL_NAMES[@]}"; do
	[[ -z "${tool//[[:space:]]/}" ]] && continue

	exists="$($YQ -r --arg t "$tool" '.tools | has($t)' "$CFG" 2>/dev/null || echo false)"
	[[ "$exists" != "true" ]] && continue

	ver="$($YQ -r --arg t "$tool" '.tools[$t].version' "$CFG")"
	spec="${tool}@${ver}"

	if ! is_managed_by_mise "$tool"; then
		run_tool_hook "$tool" preinstall install

		echo "NOT managed by mise: ${tool} -> mise use -g ${spec}"
		run mise use -g "$spec"

		run_tool_hook "$tool" postinstall install
	else
		run_tool_hook "$tool" preinstall update

		echo "Managed by mise: ${tool} -> mise update ${tool}"
		run mise update "$tool"

		if [[ "$RUN_POSTINSTALL_ON_UPDATE" == "1" ]]; then
			run_tool_hook "$tool" postinstall update
		else
			run_tool_hook "$tool" postinstall install
		fi
	fi
done

run_defaults postinstall
