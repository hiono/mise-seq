# mise-seq

**English** | [日本語](./README.ja.md) | [中文](./README.zh.md)

**mise-seq** bootstraps a **global, per-user CLI toolchain** on top of a
**system-installed `mise`**.

## Quick Install

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh
```

> Note: This method does not verify `install.sh` itself. For enhanced security,
> see "Recommended: Verified Install" below.

## Recommended: Verified Install

For security-conscious users, verify before running:

```sh
# Download SHA256SUMS and install.sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS \
  -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  -o install.sh

# Verify files match checksums
sha256sum -c SHA256SUMS --ignore-missing

# Run verified install
sh install.sh
```

## What is mise-seq?

It installs tools **sequentially** using `mise use -g TOOL@VERSION` and can
run lifecycle hooks:

- `preinstall`  (before installing/updating a tool)
- `postinstall` (after installing/updating a tool)

## Prerequisites

- `mise` is already installed **system-wide** (e.g. via `apt install mise`).
- Each user builds their own environment under their home directory.

## Why this exists

`mise install` may install multiple tools in parallel. In practice, parallel
installs can fail when a tool depends on another backend/runtime that is not
fully ready yet.

**mise-seq** makes the install order explicit (`tools_order`) and installs
tools one-by-one via `mise use -g ...`.

## Configuration (tools.yaml)

### Hook names

Only these hook names are allowed:

- `preinstall`
- `postinstall`

Any other hook key (e.g. `setup`) is invalid and must fail validation.

### Hook item shape

Each hook item looks like:

```yaml
- run: <string>                 # required
  when: [install|update|always] # optional, must be a LIST
  description: <string>         # optional
```

Notes:

- `when` must be a **list**. String form is not allowed.
- Allowed values are exactly: `install`, `update`, `always`.

### Shell script rule (IMPORTANT)

The `run` value is executed **as-is with POSIX `sh`**:

- Use standard shell variables: `$HOME`, `${VAR}`
- Do **not** use mise template syntax like `{{env.HOME}}`

The contents of `run` are **out of scope** for CUE validation.

## Validation

Before executing any installs, the runtime performs:

1) `yq -e '.' tools.yaml` (YAML parse sanity)
2) `cue vet -c=false schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'`
   (schema validation)

## Marker / state directory

Hooks are tracked with markers so:

- if a hook fails, it runs again next time
- if hook definitions change, they re-run automatically

Default marker directory:

- `${XDG_CACHE_HOME:-$HOME/.cache}/tools/state`

## CI

GitHub Actions runs:

- ShellCheck (static analysis)
- shfmt (format check)

A separate workflow (workflow_dispatch) can run a **pre-release sample
install test** using the repo-provided `.tools/tools.yaml` in an isolated
mise data directory.
