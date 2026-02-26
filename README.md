# mise-seq

mise-seq is a sequential installer that operates on top of a system-installed
`mise`.

It provides a mechanism to install developer tools one by one using a user-defined
configuration file (`tools.yaml`).

---

## Quick start

### Using a local `tools.yaml`

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s ./tools.yaml
```

### Using a remote `tools.yaml`

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s https://example.com/tools.yaml
```

- `tools.yaml` is provided by the user
- The bundled sample configuration is provided for reference only

---

## Background

While `mise` supports parallel installation of multiple tools, implicit ordering
constraints or runtime dependencies may cause failures.

mise-seq avoids these issues by adopting the following approach:

- Install tools sequentially
- Respect an explicit order (`tools_order`)
- Execute hooks before and after installation

---

## Configuration

### Basic tool definition

```yaml
tools:
  jq:
    version: latest
```

### Installation order (optional)

```yaml
tools_order:
  - jq
  - lazygit
```

---

## Hooks

The following hook types are supported:

- `preinstall`
- `postinstall`

An example is shown below.

```yaml
tools:
  jq:
    version: latest
    postinstall:
      - when: [install, update]
        run: |
          set -eu
          jq --version
```

### Practical postinstall examples

These examples are safe to use as reference for your own configuration.

**1. Version verification**

```yaml
tools:
  jq:
    version: latest
    postinstall:
      - when: [install, update]
        run: |
          set -eu
          jq --version
```

**2. Initialize config directory**

```yaml
tools:
  lazygit:
    version: latest
    postinstall:
      - when: [install]
        run: |
          set -eu
          mkdir -p "$HOME/.config/lazygit"
```

**3. Generate config template (if not exists)**

```yaml
tools:
  rg:
    version: latest
    postinstall:
      - when: [install]
        run: |
          set -eu
          cfg="$HOME/.config/ripgrep/config"
          if [ ! -f "$cfg" ]; then
            mkdir -p "$(dirname "$cfg")"
            cat >"$cfg" <<'EOF'
--smart-case
--hidden
EOF
          fi
```

### Execution rules

- Hooks are executed using POSIX `sh`
- Standard environment variables such as `$HOME` and `${VAR}` are available
- mise template syntax (`{{env.HOME}}`) is not supported
- Hook script contents are not validated by CUE

---

## Validation

Before execution, mise-seq performs the following validations:

1. YAML syntax validation
2. Schema validation using CUE

```sh
yq -e '.' tools.yaml
cue vet -c=false .tools/schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'
```

---

## Sample configuration

The repository includes `.tools/tools.yaml` and `.tools/tools.toml` as sample configuration.

- It is not required
- It has no special behavior
- It contains comments to show hook structure
- It is used for reference and pre-release validation

The primary input is always the user-provided `tools.yaml`.

---

## Installation (verified, optional)

For additional security, you can verify the release asset using GitHub's built-in SHA256 digest:

```sh
# View SHA256 digest for release assets
gh release view v0.1.0 --repo=hiono/mise-seq --json=assets

# Download and verify zip (if needed)
curl -fsSL https://github.com/hiono/mise-seq/releases/download/v0.1.0/mise-seq-release.zip -o mise-seq-release.zip
```

---

## Installation (convenience)

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s ./tools.yaml
```

In this mode, the integrity of `install.sh` itself is not verified.

---

## Prerequisites

- `mise` must be installed system-wide
- Tools are managed on a per-user basis

---

## License

MIT
