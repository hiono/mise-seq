# mise-seq

[mise](https://github.com/jdx/mise) を使ってツールをインストールする Go ライブラリ・CLI ツール。preinstall/postinstall フック対応。

[ mise-seq.sh](https://github.com/mise-seq/mise-seq.sh) の Go 実装。

---

## 機能

- **マルチフォーマット対応**: JSON、YAML、TOML、CUE
- **統合 Loader API**: フォーマットを自動検出、パース
- **mise CLI ラッパー**: Go でツールのインストール、アップグレード、一覧表示、状態確認
- **フック対応**: インストール前後にフックを実行
- **SHA256 ステート管理**: 未変更のフックをスキップ（フォースオプション対応）
- **順序付きインストール**: `tools_order` を尊重した順でインストール
- **Defaults**: 全ツールにデフォルトフックを適用
- **Settings**: mise 設定 적용 (npm, experimental)
- **CLI サブコマンド**: install、upgrade、list、status

---

## インストール

### バイナリ

[Releases](https://github.com/mise-seq/config-loader/releases) からダウンロード

### ソースから

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

# 状态を表示
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

### 設定リファレンス

すべてのフィールドは**省略可能**です（未指定はデフォルト値）。

#### ツールフィールド

| フィールド | 型 | 必須 | デフォルト | 説明 |
|-----------|------|------|-----------|------|
| `version` | string | ✗ | `"latest"` | ツールバージョン（例: `"1.22"`, `"latest"`, `""`） |
| `exe` | string | ✗ | `<ツール名>` | 実行ファイル名（ツールキーと異なる場合） |
| `depends` | array | ✗ | `[]` | 依存関係: `["tool@version"]` または `["tool"]`（= latest） |
| `preinstall` | array | ✗ | `[]` | インストール前に実行するフック |
| `postinstall` | array | ✗ | `[]` | インストール後に実行するフック |

#### 依存関係の記法

```yaml
# 完全記法
tools:
  rust:
    version: 1.88
    depends:
      - gcc@latest    # 明示的な @version
      - cargo@latest  # 明示的な @latest

# 省略記法（推奨）
tools:
  rust:
    version: 1.88
    depends:
      - gcc    # gcc@latest と同じ
      - cargo  # cargo@latest と同じ
```

**ポイント:**
- `@version` は省略可能 → 省略すると `@latest`
- `version` フィールドは省略可能 → 省略すると `"latest"`
- `exe` フィールドは省略可能 → 省略するとツールキー名
- 空配列 `[]` はフィールド省略と同じ

#### 最小設定（すべて省略）

```yaml
# すべてデフォルト: version=latest, exe=ツール名, depends=[]
tools:
  jq:
  gcc:
  rust:
```

---

## フック

以下のフック类型をサポート：

- `preinstall`: インストール前に実行
- `postinstall`: インストール後に実行

### フックタイミング

フックは特定のイベントで実行するように設定可能：

- `install`: 初回のみ
- `update`: バージョンが変わった時のみ
- `always`: 常に実行

### フック例

```yaml
tools:
  lazygit:
    version: latest
    preinstall:
      - run: |
          echo "lazygit をインストール中..."
        when: ["install"]
    postinstall:
      - run: |
          mkdir -p "$HOME/.config/lazygit"
        when: ["always"]
```

### Defaults

全ツールにフックを適用：

```yaml
defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}"
  postinstall:
    - run: echo "Done installing {{.ToolName}}"
```

### ステート管理

フックは SHA256 ハッシュを使って未変更のフックをスキップ：

- 初回実行: フックを実行し、SHA256 を保存
- 2回目以降: SHA256 を比較、一致ならスキップ
- `--force-hooks`: 未変更でも強制実行
- `--postinstall-on-update`: バージョン変更時に postinstall を実行

---

## CLI リファレンス

### コマンド

| コマンド   | 説明                        |
|-----------|---------------------------|
| `install` | 全ツールをインストール（デフォルト） |
| `upgrade` | インストール済みツールをアップグレード |
| `list`    | インストール済みツールを一覧表示     |
| `status`  | 設定ツールの状態を表示              |

### グローバルフラグ

| フラグ                       | 説明                        |
|-----------------------------|---------------------------|
| `-c <file>`                 | 設定ファイル（デフォルト: tools.yaml） |
| `--dry-run`                 | ドライランモード             |
| `--force-hooks`             | フックを強制実行             |
| `--postinstall-on-update`   | 更新時に postinstall を実行  |
| `-v`                        | 詳細出力                    |
| `--version`                 | バージョン表示              |
| `--help`                    | ヘルプ表示                  |

### 環境変数

| 変数                         | 説明                    |
|-----------------------------|------------------------|
| `DRY_RUN`                   | ドライランを有効にする     |
| `DEBUG`                     | デバッグ出力を有効にする |
| `FORCE_HOOKS`              | フックを強制実行         |
| `RUN_POSTINSTALL_ON_UPDATE` | 更新時に postinstall を実行 |
| `STATE_DIR`                 | カスタムステートディレクトリ |
| `CUE_VERSION`               | CUE バージョン           |
| `MISE_SHIMS_DEFAULT`        | mise shims パス         |
| `MISE_DATA_DIR`             | mise データディレクトリ  |

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

// defaults があるか確認
hasDefaults := config.HasDefaults(cfg)

// デフォルトフックを取得
preinstall, postinstall := config.GetDefaultsHooks(cfg)
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

// インストール済みならアップグレード
result, err := client.UpgradeIfInstalled(ctx, "jq")

// インストール済みツールを一覧表示
tools, err := client.ListTools(ctx)

// フック付きでインストール（tools_order を尊重）
err := client.InstallAllWithHooks(ctx, cfg)

// 設定を適用
err := client.ApplySettings(ctx, cfg.Settings)

// Bootstrap: mise/cue の存在確認
bootstrapper := mise.NewBootstrapper()
err := bootstrapper.EnsureMise(ctx)
err := bootstrapper.EnsureCue(ctx)
```

### hooks パッケージ

```go
runner := hooks.NewRunner(false)

// ステート管理付きでフックを実行
result, err := runner.Run(ctx, "toolname", hooks.HookTypePreinstall, "echo hello")

// 複数のフックを実行
results, err := runner.RunHooks(ctx, "toolname", hooks.HookTypePreinstall, []string{
    "echo first",
    "echo second",
})

// デフォルトフックを実行
results, err := runner.RunDefaultsHook(ctx, hooks.HookTypePreinstall, []string{
    "echo default hook",
})

// カスタムオプションでランナー作成
runner := hooks.NewRunnerWithOptions(false, "/custom/state", true, false)
```

---

## 前提条件

- Go 1.21+
- [mise](https://github.com/jdx/mise) CLI がシステムにインストール済み

---

## ライセンス

MIT
