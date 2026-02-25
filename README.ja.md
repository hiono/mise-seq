# mise-seq

[**English**](./README.md) | 日本語 | [**中文**](./README.zh.md)

**mise-seq** は、system-installed な `mise` の上で動作する
**順序付きインストーラ**です。

目的はひとつだけです。

> **curl | sh で、自分の tools.yaml をそのまま使うこと**

---

## クイックインストール（自分の tools.yaml を使う）

### ローカルの tools.yaml を使う

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s ./tools.yaml
```

### URL 上の tools.yaml を使う

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s https://example.com/tools.yaml
```

- `tools.yaml` は **あなた自身の設定ファイル**です
- mise-seq は **サンプル専用ではありません**
- サンプルは **参考・テスト用**です

---

## なぜ mise-seq か？

`mise install` は複数ツールを並列にインストールします。

実際には：

- 依存関係が間に合わない
- バックエンドが未初期化

といった問題が起きがちです。

**mise-seq** は：

- ツールを **1つずつ**
- 明示した順序で
- フック付きで

確実にインストールします。

---

## 設定（tools.yaml）

### ツール定義

```yaml
tools:
  jq:
    version: latest
```

### インストール順（任意）

```yaml
tools_order:
  - jq
  - lazygit
```

---

## フック

使用できるフックは **2つだけ**です。

- `preinstall`
- `postinstall`

```yaml
preinstall:
  - run: echo "install前"
    when: [install, update]
```

### 重要なルール

- `run` は **そのまま sh で実行**
- `$HOME`, `${VAR}` を使用
- `{{env.HOME}}` は使わない
- スクリプト内容は **CUEでは検証しません**

---

## 検証

mise-seq は実行前に必ず：

1. YAML構文チェック
2. CUEスキーマ検証

を行います。

---

## サンプル tools.yaml について

`.tools/tools.yaml` は **参考用**です。

- 必須ではありません
- 特別扱いされません
- リリース前テスト用です

**主役は常にあなたの tools.yaml です。**

---

## インストール（検証付き・任意）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/SHA256SUMS \
  -o SHA256SUMS
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  -o install.sh

sha256sum -c SHA256SUMS --ignore-missing

sh install.sh ./tools.yaml
```

---

## インストール（簡易）

```sh
curl -fsSL https://raw.githubusercontent.com/hiono/mise-seq/v0.1.0/install.sh \
  | sh -s ./tools.yaml
```

⚠️ この方法では `install.sh` 自体は検証されません。

---

## 前提条件

- `mise` が事前にインストールされていること
- ユーザーごとのホームディレクトリで管理されます

---

## ライセンス

MIT
