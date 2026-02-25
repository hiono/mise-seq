# mise-seq

[**English**](./README.md) | 日本語 | [**中文**](./README.zh.md)

`mise-seq` は、システムにインストールされた `mise` を使って、ユーザーごとにCLIツールチェーンを導入します。

## クイックインストール

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh
```

> 注意: この方法は `install.sh` 自体の検証はしません。セキュリティしたい場合は
> 「推奨: 検証済みインストール」を見てね。

## 推奨: 検証済みインストール

安全にしたい場合は、動かす前にチェックを入れる：

```sh
# SHA256SUMS と install.sh を落とす
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS \
  -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  -o install.sh

# チェックサムが合致するか確認
sha256sum -c SHA256SUMS --ignore-missing

# チェック通ったら動かす
sh install.sh
```

## 続き

ツールは **順番に** インストールします。`mise use -g TOOL@VERSION` で入れながら、以下のフックを実行できます：

- `preinstall`：ツール導入・更新の前に走る
- `postinstall`：ツール導入・更新のあとに走る

## 必要なもの

- `mise` がシステムに入った状態（例：`apt install mise` で入れる）
- 各ユーザーはホームディレクトリ下で自分用の環境を作る

## なぜこれを作ったか

`mise install` は複数ツールを並行で入れようとします。で、あるツールが依存するバックエンドやランタイムがまだ準備できてないと、並行入れは失敗します。

`mise-seq` はインストール順を明確に指定できます（`tools_order` を使う）。ツールを1つずつ `mise use -g ...` で入れていくだけです。

## 設定（tools.yaml）

### フック名

使えるフックはこの2つだけです：

- `preinstall`
- `postinstall`

それ以外（`setup` とか）は無効。検証で弾きます。

### フックのかたち

```yaml
- run: <文字列>                 # 必須
  when: [install|update|always] # 省略可、リストで書く
  description: <文字列>           # 省略可
```

注意：

- `when` は **リスト形式** で書いてください。文字列だとだめ
- 使える値は `install`、`update`、`always` の3つだけ

### スクリプトの書き方（重要）

`run` の値は **そのまま POSIX `sh`** で動きます：

- 普通のシェル変数を使ってください：`$HOME`、`${VAR}`
- `{{env.HOME}}` みたいな mise テンプレート構文は **使えません**

`run` の内容は CUE 検証の対象外です。

## 検証

インストールを動かす前に、以下のチェックが入ります：

1. `yq -e '.' tools.yaml`（YAML がパースできるか）
2. `cue vet -c=false schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'`（スキーマチェック）

## マーカーと状態ディレクトリ

フックはマーカーで管理します：

- フックが失敗したら、次回また動く
- フックの定義が変わったら、自動的に再実行

デフォルトのマーカーディレクトリは `${XDG_CACHE_HOME:-$HOME/.cache}/tools/state` です。

## CI

GitHub Actions で以下をチェックします：

- ShellCheck（静的解析）
- shfmt（フォーマット）

もう1つ、workflow_dispatch があって、リポジトリに入ってる
`.tools/tools.yaml` を使って、分離された mise データディレクトリで
**プレリリースサンプルインストールテスト** を動かせます。
