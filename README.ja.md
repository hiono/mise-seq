# mise-seq

miseSEQは、miseを使用したツールインストールをGoライブラリ・CLIとして提供する。

---

## 機能

- マルチフォーマット対応: JSON、YAML、TOML、CUE
- 統一Loader API: フォーマット自動検出
- mise CLIラッパー: インストール、アップグレード、リスト、ステータス
- フック対応: インストール前後にスクリプト実行
- SHA256ステート管理: 未変更のフックをスキップ
- 順序付きインストール: tools_orderに従う
- デフォルトフック: 全ツールに適用
- mise設定: npm、experimental対応
- CLIサブコマンド: install、upgrade、list、status

---

## インストール

### バイナリ

Releasesからダウンロード

### ソース

```bash
go install github.com/mise-seq/config-loader@latest
```

### ライブラリ

```bash
go get github.com/mise-seq/config-loader
```

---

## クイックスタート

### CLI

```bash
# ツールをインストール
mise-seq -c tools.yaml

# ドライラン
mise-seq -c tools.yaml --dry-run

# ツールをアップグレード
mise-seq upgrade -c tools.yaml

# インストール済みツールを一覧表示
mise-seq list

# ステータスを表示
mise-seq status -c tools.yaml
```

### ライブラリ

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

	loader := config.NewLoader()
	cfg, err := loader.Parse("tools.yaml")
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	client := mise.NewClient()
	if err := client.InstallAllWithHooks(ctx, cfg); err != nil {
		log.Fatalf("インストールに失敗: %v", err)
	}
}
```

---

## 設定フォーマット

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

---

## 設定リファレンス

フィールドは特記がない限り省略可能。

### ツールフィールド

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| version | string | "latest" | ツールバージョン |
| exe | string | ツールキー | 実行ファイル名 |
| depends | array | [] | 依存関係 |
| preinstall | array | [] | インストール前フック |
| postinstall | array | [] | インストール後フック |

### 依存関係記法

```yaml
# 完全記法
tools:
  rust:
    version: 1.88
    depends:
      - gcc@latest
      - cargo@latest

# 省略記法
tools:
  rust:
    version: 1.88
    depends:
      - gcc
      - cargo
```

ポイント:
- @version省略時は@latest
- version省略時は"latest"
- exe省略時はツールキー名

---

## フック

### フックタイプ

- preinstall: インストール前に実行
- postinstall: インストール後に実行

### 実行タイミング

- install: 初回インストールのみ
- update: バージョン更新時のみ
- always: 常に実行

### フック例

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

### デフォルトフック

全ツールに適用:

```yaml
defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}"
  postinstall:
    - run: echo "Done installing {{.ToolName}}"
```

### ステート管理

SHA256マーカーでフックの変更を検出:

- 初実行: フックを実行、SHA256を保存
- 以降: SHA256を比較、変更なければスキップ
- --force-hooks: 強制実行
- --postinstall-on-update: 更新時にpostinstallを実行

---

## CLIリファレンス

### コマンド

| コマンド | 説明 |
|---------|------|
| install | 全ツールをインストール（デフォルト） |
| upgrade | インストール済みツールをアップグレード |
| list | インストール済みツールを一覧表示 |
| status | 設定ツールの状態を表示 |

### グローバルフラグ

| フラグ | 説明 |
|--------|------|
| -c <file> | 設定ファイル（デフォルト: tools.yaml） |
| --dry-run | ドライラン |
| --force-hooks | フックを強制実行 |
| --postinstall-on-update | 更新時にpostinstallを実行 |
| -v | 詳細出力 |
| --version | バージョンを表示 |
| --help | ヘルプを表示 |

### 環境変数

| 変数 | 説明 |
|------|------|
| DRY_RUN | ドライランを有効化 |
| DEBUG | デバッグ出力を有効化 |
| FORCE_HOOKS | フックを強制実行 |
| RUN_POSTINSTALL_ON_UPDATE | 更新時にpostinstallを実行 |
| STATE_DIR | カスタムステートディレクトリ |
| CUE_VERSION | ブートストラップ用CUEバージョン |
| MISE_SHIMS_DEFAULT | mise shimsパス |
| MISE_DATA_DIR | miseデータディレクトリ |

---

## APIリファレンス

### configパッケージ

```go
loader := config.NewLoader()
cfg, err := loader.Parse("tools.yaml")

tools := config.GetTools(cfg)
order := config.GetToolOrder(cfg)
hasDefaults := config.HasDefaults(cfg)
preinstall, postinstall := config.GetDefaultsHooks(cfg)
```

### miseパッケージ

```go
client := mise.NewClient()

result, err := client.InstallWithOutput(ctx, "jq@latest")
installed, err := client.IsInstalled(ctx, "jq")
installed, result, err := client.InstallIfNotInstalled(ctx, "jq@latest")
result, err := client.UpgradeWithOutput(ctx, "jq")
result, err := client.UpgradeIfInstalled(ctx, "jq")
tools, err := client.ListTools(ctx)
err := client.InstallAllWithHooks(ctx, cfg)
err := client.ApplySettings(ctx, cfg.Settings)

bootstrapper := mise.NewBootstrapper()
err := bootstrapper.EnsureMise(ctx)
err := bootstrapper.EnsureCue(ctx)
```

### hooksパッケージ

```go
runner := hooks.NewRunner(false)

result, err := runner.Run(ctx, "toolname", hooks.HookTypePreinstall, "echo hello")

results, err := runner.RunHooks(ctx, "toolname", hooks.HookTypePreinstall, []string{
    "echo first",
    "echo second",
})

results, err := runner.RunDefaultsHook(ctx, hooks.HookTypePreinstall, []string{
    "echo default hook",
})

runner := hooks.NewRunnerWithOptions(false, "/custom/state", true, false)
```

---

## 前提条件

- Go 1.21以上

### mise未安装時の動作

miseがシステムにインストールされていない場合：
- 自動インストールを実行（~/.local/bin/mise）
- curlまたはwgetが必要
- インストール後はPATHに追加して続行

---

## ライセンス

MIT
