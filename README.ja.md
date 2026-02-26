# mise-seq

mise-seq は system-installed な `mise` の上で動作する
順序付きインストーラである。

ユーザーが定義した設定ファイル（`tools.yaml`）を用いて、
開発ツールを 1 つずつ確実にインストールするための仕組みを提供する。

---

## クイックスタート

### ローカルの `tools.yaml` を使用する場合

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s ./tools.yaml
```

### URL 上の `tools.yaml` を使用する場合

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s https://example.com/tools.yaml
```

- `tools.yaml` はユーザーが用意する
- 同梱されているサンプル設定は参考用であり、必須ではない

---

## 背景

`mise` は複数のツールを並列にインストール可能であるが、
暗黙的な依存関係や初期化順序の問題により失敗する場合がある。

mise-seq は以下の方針により、この問題を回避する。

- ツールを順番にインストールする
- 明示的な順序（`tools_order`）を尊重する
- インストール前後にフックを実行する

---

## 設定

### 基本的なツール定義

```yaml
tools:
  jq:
    version: latest
```

### インストール順序（任意）

```yaml
tools_order:
  - jq
  - lazygit
```

---

## フック

使用可能なフックは以下の 2 種類である。

- `preinstall`
- `postinstall`

例を以下に示す。

```yaml
preinstall:
  - run: echo "install前"
    when: [install, update]
```

### 実行ルール

- フックは POSIX `sh` として実行される
- `$HOME` や `${VAR}` などの環境変数を使用可能である
- mise テンプレート構文（`{{env.HOME}}`）は使用不可である
- フックスクリプトの内容は CUE による検証対象外である

---

## 検証

mise-seq は実行前に以下の検証を行う。

1. YAML の構文チェック
2. CUE によるスキーマ検証

```sh
yq -e '.' tools.yaml
cue vet -c=false schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'
```

---

## サンプル設定について

リポジトリには `.tools/tools.yaml` がサンプルとして含まれている。

- 必須ではない
- 特別な扱いは行われない
- リリース前テストおよび参考用途である

実際に使用されるのは、常にユーザー自身の `tools.yaml` である。

---

## インストール（検証付き・任意）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh -o install.sh

sha256sum -c SHA256SUMS --ignore-missing

sh install.sh ./tools.yaml
```

---

## インストール（簡易）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh | sh -s ./tools.yaml
```

本方式では `install.sh` 自体の検証は行われない。

---

## 前提条件

- `mise` が system-wide にインストールされていること
- ツールはユーザー単位で管理される

---

## ライセンス

MIT
