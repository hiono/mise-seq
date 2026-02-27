# mise-seq

Go 库和 CLI 工具，通过 [mise](https://github.com/jdx/mise) 安装工具，支持 preinstall/postinstall 钩子。

这是 [mise-seq.sh](https://github.com/mise-seq/mise-seq.sh) 的 Go 实现。

---

## 功能

- **多格式配置支持**: JSON、YAML、TOML、CUE
- **统一 Loader API**: 自动检测格式并解析
- **mise CLI 包装器**: 安装、升级、列出、状态检查
- **钩子支持**: 安装前后执行钩子
- **SHA256 状态管理**: 跳过未更改的钩子（支持强制选项）
- **顺序安装**: 尊重 `tools_order` 的安装顺序
- **Defaults**: 为所有工具应用默认钩子
- **Settings**: 应用 mise 设置 (npm, experimental)
- **CLI 子命令**: install、upgrade、list、status

---

## 安装

### 二进制

从 [Releases](https://github.com/mise-seq/config-loader/releases) 下载

### 源码

```bash
go install github.com/mise-seq/config-loader@latest
```

### 库

```bash
go get github.com/mise-seq/config-loader
```

---

## 快速开始

### CLI

```bash
# 从配置安装工具
mise-seq -c tools.yaml

# 试运行
mise-seq -c tools.yaml --dry-run

# 升级已安装的工具
mise-seq upgrade -c tools.yaml

# 列出已安装的工具
mise-seq list

# 显示状态
mise-seq status -c tools.yaml
```

### 库

```go
package main

import (
	"context"
	"log"

	"github.com/mise-seq/config-loader/config"
	"github.com/mise-seq/config-loader/mise"
)

func main() {
	ctx := context.Background()

	// 加载配置
	loader := config.NewLoader()
	cfg, err := loader.Parse("tools.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 带钩子安装工具
	client := mise.NewClient()
	if err := client.InstallAllWithHooks(ctx, cfg); err != nil {
		log.Fatalf("安装失败: %v", err)
	}
}
```

---

## 配置格式

### YAML

```yaml
tools:
  jq:
    version: latest
  lazygit:
    version: latest

tools_order:
  - jq
  - lazygit

defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}..."
  postinstall:
    - run: echo "Installed {{.ToolName}}"

settings:
  npm:
    package_manager: pnpm
  experimental: true
```

### JSON

```json
{
  "tools": {
    "jq": { "version": "latest" },
    "lazygit": { "version": "latest" }
  },
  "tools_order": ["jq", "lazygit"],
  "defaults": {
    "preinstall": [{ "run": "echo Installing..." }]
  }
}
```

### TOML

```toml
[tools.jq]
version = "latest"

[tools.lazygit]
version = "latest"

tools_order = ["jq", "lazygit"]

[defaults.preinstall]
run = "echo Installing..."

[settings.npm]
package_manager = "pnpm"
```

### CUE

```cue
MiseSeqConfig: {
    tools: {
        jq: version: "latest"
        lazygit: version: "latest"
    }
    tools_order: ["jq", "lazygit"]
    defaults: preinstall: [{run: "echo Installing..."}]
}
```

### 配置参考

所有字段都是**可选的**（未指定时使用默认值）。

#### 工具字段

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `version` | string | 否 | `"latest"` | 工具版本（如 `"1.22"`, `"latest"`, `""`） |
| `exe` | string | 否 | `<工具名>` | 可执行文件名（与工具键名不同时） |
| `depends` | array | 否 | `[]` | 依赖: `["tool@version"]` 或 `["tool"]`（= latest） |
| `preinstall` | array | 否 | `[]` | 安装前运行的钩子 |
| `postinstall` | array | 否 | `[]` | 安装后运行的钩子 |

#### 依赖语法

```yaml
# 完整语法
tools:
  rust:
    version: 1.88
    depends:
      - gcc@latest    # 显式指定 @version
      - cargo@latest  # 显式指定 @latest

# 简写（推荐）
tools:
  rust:
    version: 1.88
    depends:
      - gcc    # 等同于 gcc@latest
      - cargo  # 等同于 cargo@latest
```

**要点:**
- `@version` 可省略 → 默认为 `@latest`
- `version` 字段可省略 → 默认为 `"latest"`
- `exe` 字段可省略 → 默认为工具键名
- 空数组 `[]` 等同于省略字段

#### 最小配置（全部省略）

```yaml
# 全部默认: version=latest, exe=工具名, depends=[]
tools:
  jq:
  gcc:
  rust:
```

---

## 钩子

支持以下钩子类型：

- `preinstall`: 安装前运行
- `postinstall`: 安装后运行

### 钩子时机

钩子可配置为在特定事件时运行：

- `install`: 仅首次安装时
- `update`: 仅版本升级时
- `always`: 始终运行

### 钩子示例

```yaml
tools:
  lazygit:
    version: latest
    preinstall:
      - run: |
          echo "Installing lazygit..."
        when: ["install"]
    postinstall:
      - run: |
          mkdir -p "$HOME/.config/lazygit"
        when: ["always"]
```

### Defaults

为所有工具应用钩子：

```yaml
defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}"
  postinstall:
    - run: echo "Done installing {{.ToolName}}"
```

### 状态管理

钩子使用 SHA256 标记跳过未更改的钩子：

- 首次运行：执行钩子，保存 SHA256
- 后续运行：比较 SHA256，一致则跳过
- `--force-hooks`: 强制执行即使未更改
- `--postinstall-on-update`: 版本变更时运行 postinstall

---

## CLI 参考

### 命令

| 命令     | 描述                   |
|---------|----------------------|
| `install` | 安装所有工具（默认）   |
| `upgrade` | 升级已安装的工具       |
| `list`    | 列出已安装的工具       |
| `status`  | 显示配置工具的状态     |

### 全局标志

| 标志                      | 描述                     |
|--------------------------|------------------------|
| `-c <file>`              | 配置文件（默认: tools.yaml） |
| `--dry-run`              | 试运行模式               |
| `--force-hooks`          | 强制执行钩子             |
| `--postinstall-on-update`| 更新时运行 postinstall   |
| `-v`                     | 详细输出                 |
| `--version`              | 显示版本                 |
| `--help`                 | 显示帮助                 |

### 环境变量

| 变量                       | 描述                  |
|---------------------------|---------------------|
| `DRY_RUN`                 | 启用试运行            |
| `DEBUG`                  | 启用调试输出          |
| `FORCE_HOOKS`           | 强制执行钩子          |
| `RUN_POSTINSTALL_ON_UPDATE`| 更新时运行 postinstall |
| `STATE_DIR`               | 自定义状态目录        |
| `CUE_VERSION`           | CUE 版本             |
| `MISE_SHIMS_DEFAULT`    | Mise shims 路径      |
| `MISE_DATA_DIR`          | Mise 数据目录        |

---

## API 参考

### config 包

```go
// 创建新加载器
loader := config.NewLoader()

// 解析配置文件（自动检测）
cfg, err := loader.Parse("tools.yaml")

// 获取工具列表
tools := config.GetTools(cfg)

// 获取安装顺序
order := config.GetToolOrder(cfg)

// 检查是否有默认值
hasDefaults := config.HasDefaults(cfg)

// 获取默认钩子
preinstall, postinstall := config.GetDefaultsHooks(cfg)
```

### mise 包

```go
client := mise.NewClient()

// 安装单个工具
result, err := client.InstallWithOutput(ctx, "jq@latest")

// 检查工具是否已安装
installed, err := client.IsInstalled(ctx, "jq")

// 未安装时安装
installed, result, err := client.InstallIfNotInstalled(ctx, "jq@latest")

// 升级工具
result, err := client.UpgradeWithOutput(ctx, "jq")

// 已安装则升级
result, err := client.UpgradeIfInstalled(ctx, "jq")

// 列出已安装的工具
tools, err := client.ListTools(ctx)

// 带钩子安装（尊重 tools_order）
err := client.InstallAllWithHooks(ctx, cfg)

// 应用设置
err := client.ApplySettings(ctx, cfg.Settings)

// Bootstrap: 确保 mise/cue 可用
bootstrapper := mise.NewBootstrapper()
err := bootstrapper.EnsureMise(ctx)
err := bootstrapper.EnsureCue(ctx)
```

### hooks 包

```go
runner := hooks.NewRunner(false)

// 带状态管理运行钩子
result, err := runner.Run(ctx, "toolname", hooks.HookTypePreinstall, "echo hello")

// 运行多个钩子
results, err := runner.RunHooks(ctx, "toolname", hooks.HookTypePreinstall, []string{
    "echo first",
    "echo second",
})

// 运行默认钩子
results, err := runner.RunDefaultsHook(ctx, hooks.HookTypePreinstall, []string{
    "echo default hook",
})

// 带自定义选项创建运行器
runner := hooks.NewRunnerWithOptions(false, "/custom/state", true, false)
```

---

## 前置条件

- Go 1.21+

### 自动安装行为

如果系统上未安装 mise：
- 自动安装 mise 到 ~/.local/bin/mise
- 需要 curl 或 wget
- 安装后添加到 PATH

---

## 许可证

MIT
