*This project has been created as part of the 42 curriculum by amakino, takawaka.*

# The Answer Protocol (TAP)

> TODO: 正式な提出用READMEはTASKS.md記載の必須フォーマット（1行目のイタリック文言、Description/Instructions/Resources、Architecture/Protocol Implementation/Combat System/Quest System/World Design/Server Logging/Group Contributions/Building and Running/Testingの各セクション、英語必須）に沿って後で整備すること。以下は開発中のGo学習メモ。

## RFC

`RFC`という文書形式自体の説明

### 何の略か、何のための文書か

**RFC = Request for Comments**（意見募集）。1969年、初期のインターネット(ARPANET)の技術者たちが「こういう仕様どう？意見ちょうだい」という形でメモを回覧しはじめたのが起源で、そのまま名前が定着した。現在は**IETF**(Internet Engineering Task Force)という団体が管理していて、インターネットで使われるプロトコル(TCP/IP, HTTP, SMTP, DNSなど)のほぼ全てが、最終的にRFCという形式の文書で正式に定義されている。

つまりRFCは「みんなが同じ実装をできるように、通信の約束事を文章で厳密に定めたもの」。ブラウザを作る会社とサーバーを作る会社が違っても、同じRFCに従っていれば通信できる、というのがポイント（今回の課題で「他チームのサーバーと繋がらないといけない」のと本質的に同じ理由）。

### 基本的な構成要素

実際のRFCにだいたい共通して出てくる要素：

- **番号**: 発行された順に連番が振られる。一度発行されたRFCは内容を変更できず、改訂する場合は新しい番号のRFCが発行されて古い方を「obsolete(廃止)」にする
- **ステータス分類**: `Standards Track`(正式な標準)、`Informational`(参考情報)、`Experimental`(実験的)、`Best Current Practice`(推奨運用)など。今回の課題のRFCは`Category: Experimental`を名乗っている(実際は42が課題用に作った架空の文書だが、体裁を real RFC に寄せている)
- **キーワード規定(RFC 2119)**: `MUST`/`MUST NOT`/`SHOULD`/`SHOULD NOT`/`MAY`などの単語を、文書内で「絶対守るべき」「推奨」「任意」という強制力の強さを表す専門用語として定義したルール。RFCを読むときはこの単語の重みを意識する必要がある(`protocol-rfc.html`もこの規約に従うと明記している)
- **ABNF(RFC 5234)**: "Augmented Backus-Naur Form"の略で、コマンドの文法を数式のように厳密に書くための記法。例えば`command-line = command-name [SP arguments] LF`のような書き方がこれ。曖昧な自然言語だけでなく、機械的にも解釈できる文法定義を添えるのがRFCの伝統
- **セクション構成の型**: Introduction → 用語定義 → 本体仕様 → エラー処理 → セキュリティ考慮事項 → 参考文献 → 著者情報、という並びはほぼ全RFCで共通のテンプレート

### 実在するRFCの例

`protocol-rfc.html`の11章(References)に挙がっている本物のRFC：

- **RFC 2119**: 前述の MUST/SHOULD/MAY のキーワード定義そのもの
- **RFC 5234**: ABNF記法自体の仕様
- **RFC 793**: TCP(今回サーバーが使う通信の土台そのもの)
- **RFC 3629**: UTF-8エンコーディングの仕様

つまり今回配布された`protocol-rfc.html`は、これら本物のRFCのルール(MUST/SHOULDの使い方、ABNFの書き方)を借りて、42が課題用に独自プロトコルを定義した**非公式の模倣RFC**、という位置づけ。中身の位置づけとしては「本物ではないが、本物と同じ厳密さで読むことが要求されている文書」。

## Goメモ

Goの基本文法・標準ライブラリの整理

### `package main` の意味

Goのソースファイルは先頭で必ず所属パッケージを宣言する。これは同じフォルダ内の全`.go`ファイルに共通する所属先。

- `package main` は特別な宣言で、「実行可能なプログラムである」という宣言
- ビルドツール(`go build`)は「`package main`かつ`func main()`がある」フォルダを実行可能ファイルとしてコンパイルする
- 逆に `package mylib` のような名前なら、他のプログラムから`import`して使うライブラリ扱いになり、単体では実行できない

### `log` パッケージ

ログ出力専用の標準ライブラリ。`fmt.Println`等と違い、自動でタイムスタンプが付き、致命的エラー用の便利関数もある。

```go
log.Println("hello")           // 2024/01/01 12:00:00 hello  ← 自動で日時が先頭に付く
log.Printf("count=%d", 5)      // フォーマット付き出力
log.Fatalf("boom: %v", err)    // 出力した直後にプログラムを終了(os.Exit(1))させる
```

出力先はデフォルトで標準エラー出力(stderr)。

### `net` パッケージ

ネットワーク通信（TCP/UDP/IPソケット）全般を扱う標準ライブラリ。C言語の`socket()/bind()/listen()/accept()/connect()`相当の機能がまとまっている。

```go
net.Listen("tcp", ":4242")   // TCPソケットを作ってポート4242で待ち受け開始 → net.Listenerを返す
ln.Accept()                   // 誰かが接続してくるまで待って、繋がったら net.Conn を返す
conn.RemoteAddr()             // 接続してきた相手のIPアドレス:ポートを取得
conn.Close()                  // 接続を閉じる
```

### Goのエラー処理パターン（例外なし、戻り値ベース）

Goには例外(try/catch)が無い。関数は「結果」と「エラー」の2つを同時に返すのが基本パターンで、呼び出し側は毎回チェックする。

```go
ln, err := net.Listen("tcp", ":4242")
if err != nil {
    log.Fatalf("listen failed: %v", err)
}
```

`:=` は「宣言＋代入」を同時に行う演算子（型は右辺から自動推論される）。

### `net.Listener` はインターフェース

具体的な構造体ではなく、以下のメソッドを持つ型なら何でも当てはまる「インターフェース」型。

```go
type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() Addr
}
```

`net.Listen("tcp", ...)`が返す実体は`*net.TCPListener`だが、これが上記3メソッドを実装しているため`net.Listener`型の変数に代入できる。Unixドメインソケット用の`*net.UnixListener`やTLSを被せた`*tls.Listener`など実装が複数あっても、呼び出し側はインターフェースだけ見て統一的に扱える（構造的部分型／ダックタイピング）。

### `error` もインターフェース

```go
type error interface {
	Error() string
}
```

`Error() string`を持つ型なら何でも`error`として扱える。`nil`は「エラーが起きなかった」を意味する。`%v`で表示すると内部で`err.Error()`が呼ばれて文字列化される。

`net.Listen`の失敗時に実際に返るのは`*net.OpError`という具体型で、操作名・アドレス・OSレベルのエラー(`syscall.Errno`)まで構造化されているが、普段は`%v`で丸ごと文字列化するか、特定のエラー種別判定が必要な時だけ`errors.Is()`等を使う。


