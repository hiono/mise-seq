# mise-seq

[**English**](./README.md) | [**日本語**](./README.ja.md) | 中文

**mise-seq** 是构建在 system-installed `mise` 之上的
**顺序安装工具**。

它的核心目标只有一个：

> **通过 curl | sh，直接使用你自己的 tools.yaml**

---

## 快速安装（使用你自己的 tools.yaml）

### 使用本地 tools.yaml

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s ./tools.yaml
```

### 使用远程 tools.yaml

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s https://example.com/tools.yaml
```

- `tools.yaml` 是你**自己的配置**
- mise-seq **不限于示例文件**
- 示例仅用于参考和测试

---

## 为什么使用 mise-seq？

`mise install` 可能并行安装多个工具，
在实践中容易失败。

mise-seq 会：

- 按顺序逐个安装
- 遵循 `tools_order`
- 支持生命周期钩子

---

## 配置（tools.yaml）

### 工具定义

```yaml
tools:
  jq:
    version: latest
```

### 明确安装顺序（可选）

```yaml
tools_order:
  - jq
  - lazygit
```

---

## 钩子

仅允许以下钩子：

- `preinstall`
- `postinstall`

示例：

```yaml
preinstall:
  - run: echo "before install"
    when: [install, update]
```

### 重要规则

- `run` 以 **POSIX `sh`** 原样执行
- 使用标准 shell 变量：`$HOME`、`${VAR}`
- **禁止**使用 mise 模板语法（`{{env.HOME}}`）
- 钩子脚本 **不在 CUE 验证范围内**

---

## 验证

执行安装前，mise-seq 会运行：

1. YAML 解析检查
2. 架构验证

---

## 示例 tools.yaml

仓库中的 `.tools/tools.yaml` 仅作为 **示例**。

- **非必需**
- **非特殊**
- 仅用于参考和发布测试

你的 `tools.yaml` 始终被支持，是主要用例。

---

## 安装（已验证，可选）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS \
  -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  -o install.sh

sha256sum -c SHA256SUMS --ignore-missing

sh install.sh ./tools.yaml
```

---

## 安装（简便）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s ./tools.yaml
```

⚠️ 此方式不会验证 `install.sh` 本身。

---

## 前提条件

- `mise` 必须已系统级安装
- 每个用户在主目录下管理工具

---

## 许可证

MIT
