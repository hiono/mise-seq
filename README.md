# mise-seq

**mise-seq** is a **sequential installer** built on top of a
system-installed `mise`.

It exists for one reason:

> **Use `curl | sh` to install tools sequentially with your own
> `tools.yaml`.**

---

## Quick Install (use your own `tools.yaml`)

### Use a local `tools.yaml`

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s ./tools.yaml
```

### Use a remote `tools.yaml`

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s https://example.com/tools.yaml
```

- `tools.yaml` is **your own configuration**
- mise-seq is **NOT limited to the bundled sample**
- The bundled sample is **for reference and testing only**

---

## Why mise-seq?

`mise install` may install multiple tools in parallel.
In practice, parallel installs can fail when tools depend on each other.

**mise-seq** solves this by:

- Installing tools **one by one**
- Respecting an explicit order (`tools_order`)
- Running lifecycle hooks (`preinstall` / `postinstall`)

---

## Configuration (`tools.yaml`)

### Tool definition

```yaml
tools:
  jq:
    version: latest
```

### Explicit install order (optional)

```yaml
tools_order:
  - jq
  - lazygit
```

---

## Hooks

Only these hooks are allowed:

- `preinstall`
- `postinstall`

Example:

```yaml
preinstall:
  - run: echo "before install"
    when: [install, update]
```

### Important rules

- `run` is executed **as-is with POSIX `sh`**
- Use standard shell variables (`$HOME`, `${VAR}`)
- **Do NOT** use mise template syntax (`{{env.HOME}}`)
- Hook scripts are **not validated** by CUE

---

## Validation

Before installing anything, mise-seq runs:

1. YAML parse check

   ```sh
   yq -e '.' tools.yaml
   ```

2. Schema validation

   ```sh
   cue vet -c=false schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'
   ```

---

## Sample `tools.yaml`

The repository includes `.tools/tools.yaml` as a **sample**.

- It is **not required**
- It is **not special**
- It exists only as a reference and for release testing

Your own `tools.yaml` is always supported and is the primary use case.

---

## Installation (verified, optional)

If you want reproducibility and verification:

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS \
  -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  -o install.sh

sha256sum -c SHA256SUMS --ignore-missing

sh install.sh ./tools.yaml
```

---

## Installation (convenience)

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s ./tools.yaml
```

⚠️ This does not verify `install.sh` itself.
Use the verified method above if security or reproducibility matters.

---

## Prerequisites

- `mise` must already be installed system-wide
- Each user manages tools in their own home directory

---

## License

MIT
