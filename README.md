# pz

Paizaスキルチェック用CLIツール

## 特徴

- Paizaのスキルチェックにおけるコーディングをローカルで行える
    - 好きなエディタを利用可
- 事前に作成したテンプレートを使用してコーディング
- テストパターンをチェック
- テストパターンの特定のパターンでデバッグ
- コーディング結果をクリップボードへコピー

できることと使い方の概要を以下の動画で紹介しています。
(クリックするとYouTubeページへジャンプします。)

[![概要動画](/img/explain.png)](http://www.youtube.com/watch?v=FTqw5-lfkNE "概要動画")

## 導入方法

### 動作環境の構築

- [Google Chrome](https://www.google.com/intl/ja_jp/chrome/)のインストール
    - Chrome以外のブラウザにはまだ対応していません
- [chromedriver](https://chromedriver.chromium.org/downloads)をダウンロードし、パスの通っているディレクトリに配置
    - 使用しているOS、Chromeのバージョンに合っているものを使うこと

### pzのインストール

以下のどちらかでインストール

#### バイナリファイルをダウンロード (現在準備中)

[Releases](https://github.com/reeve0930/pz/releases)にある最新版のバイナリファイルをダウンロードし、パスの通っているディレクトリに配置
    - 使用しているOSに合ったものを使うこと

#### `go get`でインストール

Goの環境構築ができている人は以下のコマンドでインストール

```zsh
go get github.com/reeve0930/pz
```

### 使い方

(現在作成中)