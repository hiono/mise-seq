# mise-seq

[**English**](./README.md) | [**日本語**](./README.ja.md) | 中文

**mise-seq** 可以在已安装的 **`mise`** 基础上，为用户全局安装 **CLI 工具链**。

## 快速安装

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh
```

> 注意：此方法不验证 `install.sh` 本身。如需更高安全性，请参见
> 「推荐：验证后安装」。

## 推荐：验证后安装

安全起见，运行前先验证：

```sh
# 下载 SHA256SUMS 和 install.sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS \
  -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  -o install.sh

# 验证文件是否匹配校验和
sha256sum -c SHA256SUMS --ignore-missing

# 验证通过后运行安装
sh install.sh
```

## 什么是 mise-seq？

工具采用**顺序**安装方式，通过 `mise use -g TOOL@VERSION` 安装，并可执行以下生命周期钩子：

- `preinstall`（工具安装/更新前）
- `postinstall`（工具安装/更新后）

## 前提条件

- `mise` 已**系统级**安装（例如通过 `apt install mise`）。
- 每个用户在主目录下构建自己的环境。

## 为何创建此工具

`mise install` 可能并行安装多个工具。实际上，当某个工具依赖的后端或运行时尚未就绪时，并行安装可能会失败。

**mise-seq** 明确指定安装顺序（`tools_order`），通过 `mise use -g ...` 逐个安装工具。

## 配置（tools.yaml）

### 钩子名称

仅允许使用以下钩子名称：

- `preinstall`
- `postinstall`

其他钩子键（例如 `setup`）无效，验证时必须失败。

### 钩子项结构

每个钩子项的格式如下：

```yaml
- run: <字符串>                 # 必填
  when: [install|update|always] # 可选，必须为列表
  description: <字符串>           # 可选
```

注意事项：

- `when` 必须是**列表**。不支持字符串形式。
- 允许的值仅为：`install`、`update`、`always`。

### 脚本规则（重要）

`run` 的值以 **POSIX `sh`** 原样执行：

- 使用标准 shell 变量：`$HOME`、`${VAR}`
- **禁止**使用 mise 模板语法，如 `{{env.HOME}}`

`run` 的内容不在 CUE 验证范围内。

## 验证

执行任何安装前，运行时会执行以下操作：

1. `yq -e '.' tools.yaml`（YAML 解析完整性检查）
2. `cue vet -c=false schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'`（架构验证）

## 标记/状态目录

钩子通过标记追踪：

- 如果钩子失败，下次运行时会再次执行。
- 如果钩子定义发生变化，将自动重新执行。

默认标记目录：

- `${XDG_CACHE_HOME:-$HOME/.cache}/tools/state`

## CI

GitHub Actions 执行以下检查：

- ShellCheck（静态分析）
- shfmt（格式检查）

另有一个工作流（workflow_dispatch）可在独立的 mise 数据目录中，使用仓库提供的 `.tools/tools.yaml` 运行**预发布示例安装测试**。
