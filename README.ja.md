# config-loader

CUE、JSON、YAML、TOML 設定ファイルを読み込み、[mise](https://github.com/jdx/mise) CLI をラップしてツール管理を行う Go ライブラリ。

---

## 機能

- **マルチフォーマット対応**: JSON、YAML、TOML、CUE
- **統合 Loader API**: フォーマットを自動検出、パース
- **mise CLI ラッパー**: Go でツールのインストール、アップグレード、一覧表示
- **フック対応**: インストール前後にフックを実行
- **順序付きインストール**: `tools_order` を尊重した順でインストール

---

## インストール

```bash
go get github.com/mise-seq/config-loader
```

---

## クイックスタート

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

    // 設定ファイルを読み込み
    loader := config.NewLoader()
    cfg, err := loader.Parse("tools.yaml")
    if err != nil {
        log.Fatalf("設定読み込み失敗: %v", err)
    }

    // フック付きでツールをインストール
    client := mise.NewClient()
    if err := client.InstallAllWithHooks(ctx, cfg); err != nil {
        log.Fatalf("インストール失敗: %v", err)
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
```

### JSON

```json
{
  "tools": {
    "jq": { "version": "latest" },
    "lazygit": { "version": "latest" }
  },
  "tools_order": ["jq", "lazygit"]
}
```

### TOML

```toml
[tools.jq]
version = "latest"

[tools.lazygit]
version = "latest"

tools_order = ["jq", "lazygit"]
```

---

## フック

以下のフック类型をサポート：

- `preinstall`: インストール前に実行
- `postinstall`: インストール後に実行

### フック例

```yaml
tools:
  lazygit:
    version: latest
    preinstall:
      - run: |
          echo "lazygit をインストール中..."
    postinstall:
      - run: |
          mkdir -p "$HOME/.config/lazygit"
```

---

## API リファレンス

### config パッケージ

```go
// 新規ローダーを作成
loader := config.NewLoader()

// 設定ファイルをパース（自動検出）
cfg, err := loader.Parse("tools.yaml")

// ツール一覧を取得
tools := config.GetTools(cfg)

// インストール順序を取得
order := config.GetToolOrder(cfg)
```

### mise パッケージ

```go
client := mise.NewClient()

// 単一ツールをインストール
result, err := client.InstallWithOutput(ctx, "jq@latest")

// ツールがインストール済みか確認
installed, err := client.IsInstalled(ctx, "jq")

// 未インストール時のみインストール
installed, result, err := client.InstallIfNotInstalled(ctx, "jq@latest")

// ツールをアップグレード
result, err := client.UpgradeWithOutput(ctx, "jq")

// インストール済みツールを一覧表示
tools, err := client.ListTools(ctx)

// フック付きでインストール（tools_order を尊重）
err := client.InstallAllWithHooks(ctx, cfg)
```

### hooks パッケージ

```go
runner := hooks.NewRunner(false)

// 单一のフックを実行
result, err := runner.Run(ctx, "echo hello")

// 複数のフックを実行
results, err := runner.RunHooks(ctx, []string{
    "echo first",
    "echo second",
})
```

---

## 前提条件

- Go 1.21+
- [mise](https://github.com/jdx/mise) CLI がシステムにインストール済み

---

## ライセンス

MIT
