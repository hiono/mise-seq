# mise-seq

mise-seq は system-installed な `mise` の上で動作する順序付きインストーラである。

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
- 同名されているサンプル設定は参考用であり、必須ではない

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

例および実用的な使用例を以下に示す。

### バージョン確認

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

### 設定ディレクトリの初期化

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

### 設定ファイルの雛形生成（既存ファイルは上書きしない）

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
cue vet -c=false .tools/schema/mise-seq.cue tools.yaml -d '#MiseSeqConfig'
```

---

## サンプル設定

リポジトリには `.tools/tools.yaml` および `.tools/tools.toml` がサンプルとして含まれている。

- 必須ではない
- 特別な扱いは行われない
- フックの構造を示すためにコメントを含んでいる
- リリース前テストおよび参考用途である

実際に使用されるのは、常にユーザー自身の `tools.yaml` である。

---

## インストール（検証付き・任意）

セキュリティ強化のため、GitHub公式のSHA256 digestでリリースアセットを検証できる：

```sh
# リリースアセットのSHA256 digestを表示
gh release view v0.1.0 --repo=hiono/mise-seq --json=assets

# zipをダウンロードして検証（必要に応じて）
curl -fsSL https://github.com/hiono/mise-seq/releases/download/v0.1.0/mise-seq-release.zip -o mise-seq-release.zip
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
