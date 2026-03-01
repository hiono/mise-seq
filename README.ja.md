# mise-seq

[English](./README.md) | [日本語](./README.ja.md) | [中文](./README.zh.md)

---

miseを使用したツールインストールをGoライブラリ・CLIとして提供する。

---

## 機能

- マルチフォーマット対応: JSON、YAML、TOML、CUE
- 統一Loader API: フォーマット自動検出
- mise CLIラッパー: インストール、アップグレード、リスト、ステータス
- プラグイン対応: インストール前にmiseプラグインを自動追加
- パッケージ形式: runtime:tool (npm:difit, go:github.com/...)
- 順序付きリスト: リスト順=インストール順
- ツール無効化: disabled: trueでインストールをスキップ
- フック対応: インストール前後にスクリプト実行
- SHA256ステート管理: 未変更のフックをスキップ
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
  - name: jq
    version: latest
  - name: lazygit
    version: latest
  - name: glab
    version: latest
    plugin: glab
  - name: difit
    package: npm:difit
  - name: gemini
    package: npm:@google/gemini-cli
    disabled: true

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
  "tools": [
    { "name": "jq", "version": "latest" },
    { "name": "lazygit", "version": "latest" }
  ],
  "defaults": {
    "preinstall": [{ "run": "echo Installing..." }]
  }
}
```

### TOML

```toml
[[tools]]
name = "jq"
version = "latest"

[[tools]]
name = "lazygit"
version = "latest"

[defaults.preinstall]
run = "echo Installing..."

[settings.npm]
package_manager = "pnpm"
```

### CUE

```cue
MiseSeqConfig: {
    tools: [
        {name: "jq", version: "latest"},
        {name: "lazygit", version: "latest"}
    ]
    defaults: preinstall: [{run: "echo Installing..."}]
}
```

---

## 設定リファレンス

フィールドは省略可能。

### ツールフィールド

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| name | string | 必須 | ツール名 |
| version | string | "latest" | ツールバージョン |
| package | string | - | runtime:tool形式 (例: npm:difit) |
| plugin | string | - | インストール前に追加するmiseプラグイン |
| exe | string | name | 実行ファイル名 |
| disabled | bool | false | インストールをスキップ |
| depends | array | [] | 依存関係（ツール名のみ） |
| preinstall | array | [] | インストール前フック |
| postinstall | array | [] | インストール後フック |

### 依存関係記法

```yaml
tools:
  - name: rust
    version: 1.88
    depends:
      - gcc
      - cargo
```

ポイント:
- version、省略時は"latest"
- exe、省略時はツール名

---

## フック

### フックタイプ

preinstall: インストール前
postinstall: インストール後

### 実行タイミング

install: 初回のみ
update: バージョン更新時のみ
always: 常に実行

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

```yaml
defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}"
  postinstall:
    - run: echo "Done installing {{.ToolName}}"
```

### ステート管理

SHA256マーカーで変更を検出:
- 初実行: フックを実行、SHA256を保存
- 以降: SHA256を比較、未変更ならスキップ
- --force-hooks: 強制実行
- --postinstall-on-update: 更新時にpostinstallを実行

---

## CLIリファレンス

### コマンド

| コマンド | 説明 |
|---------|------|
| install | 全ツールをインストール |
| upgrade | ツールをアップグレード |
| list | ツールを一覧表示 |
| status | 設定ツールの状態を表示 |

### グローバルフラグ

| フラグ | 説明 |
|--------|------|
| -c file | 設定ファイル |
| --dry-run | ドライラン |
| --force-hooks | フックを強制実行 |
| --postinstall-on-update | 更新時にpostinstallを実行 |
| -v | 詳細出力 |
| --version | バージョン |
| --help | ヘルプ |

### 環境変数

| 変数 | 説明 |
|------|------|
| DRY_RUN | ドライラン |
| DEBUG | デバッグ出力 |
| FORCE_HOOKS | フック強制実行 |
| RUN_POSTINSTALL_ON_UPDATE | 更新時postinstall |
| STATE_DIR | ステートディレクトリ |
| CUE_VERSION | CUEバージョン |
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

// 単一ツールをインストール
result, err := client.InstallWithOutput(ctx, "jq@latest")

// ツールがインストール済みか確認
installed, err := client.IsInstalled(ctx, "jq")

// ツールがmiseで管理されているか確認
managed, err := client.IsManagedByMise(ctx, "jq")

// 未インストールの場合のみインストール
installed, result, err := client.InstallIfNotInstalled(ctx, "jq@latest")

// グローバルデフォルトを設定（mise use -g相当）
err := client.SetGlobal(ctx, "jq@latest")

// ツールをアップグレード
result, err := client.UpgradeWithOutput(ctx, "jq")

// インストール済みならアップグレード
result, err := client.UpgradeIfInstalled(ctx, "jq")

// インストール済みツールを一覧表示
tools, err := client.ListTools(ctx)

// フック付きでインストール（tools_orderに従う、自动的にinstall/upgradeを判定）
err := client.InstallAllWithHooks(ctx, cfg, runPostinstallOnUpdate)

// InstallAllWithHooks: runPostinstallOnUpdate=trueでupgrade時にpostinstallを実行

// 設定を適用
err := client.ApplySettings(ctx, cfg.Settings)

// ブートストラップ: mise/cueの確保
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

Go 1.21以上

### mise自動インストール

mise未安装の場合:
- ~/.local/bin/mise にダウンロード
- PATHに追加
- 以降コマンドでmise可以利用

インストール後:

```bash
exec $SHELL
```

確認: `mise --version`

---

## ライセンス

MIT
