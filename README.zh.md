# mise-seq

mise-seq 是运行在 system-installed `mise` 之上的
顺序安装工具。

该工具使用用户定义的配置文件（`tools.yaml`），
以顺序方式逐个安装开发工具。

---

## 快速开始

### 使用本地 `tools.yaml`

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s ./tools.yaml
```

### 使用远程 `tools.yaml`

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s https://example.com/tools.yaml
```

- `tools.yaml` 由用户提供
- 仓库中包含的示例配置仅用于参考

---

## 背景

尽管 `mise` 支持并行安装多个工具，
但在存在隐式依赖或初始化顺序要求时，
并行安装可能导致失败。

mise-seq 通过以下方式避免这些问题：

- 按顺序安装工具
- 遵循显式顺序（`tools_order`）
- 在安装前后执行钩子

---

## 配置

### 基本工具定义

```yaml
tools:
  jq:
    version: latest
```

### 安装顺序（可选）

```yaml
tools_order:
  - jq
  - lazygit
```

---

## 钩子

支持以下钩子类型：

- `preinstall`
- `postinstall`

以下为示例及实际使用方法。

### 版本确认

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

### 初始化配置目录

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

### 生成配置模板（不覆盖已存在的文件）

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

### 执行规则

- 钩子使用 POSIX `sh` 执行
- 支持 `$HOME`、`${VAR}` 等标准环境变量
- 不支持 mise 模板语法（`{{env.HOME}}`）
- 钩子脚本内容不由 CUE 验证

---

## 验证

mise-seq 在执行前执行以下验证：

1. YAML 语法验证
2. 使用 CUE 进行模式验证

```sh
yq -e '.' tools.yaml
cue vet -c=false .tools/schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'
```

---

## 示例配置

仓库中包含 `.tools/tools.yaml` 和 `.tools/tools.toml` 作为示例配置。

- 非必需
- 无特殊行为
- 包含注释以展示钩子结构
- 仅用于参考和发布前验证

主要输入始终是用户提供的 `tools.yaml`。

---

## 安装（已验证，可选）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh -o install.sh

sha256sum -c SHA256SUMS --ignore-missing

sh install.sh ./tools.yaml
```

---

## 安装（简易）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s ./tools.yaml
```

此方式不会验证 `install.sh` 本身的完整性。

---

## 前置条件

- 系统中必须已安装 `mise`
- 工具按用户范围进行管理

---

## 许可证

MIT
