# The Answer Protocol コード1行ずつ解説(Go初心者向け)

ソースを**上から順に、1文ずつ**読んでいくための解説です。

## このドキュメントの読み方

- 各章は、ひとつのファイル(または関連するファイルのまとまり)です。まずそのファイルの役割を1〜2行で書き、そのあとコードを**数行ずつのかたまり**に切って、各行を説明します。
- コードの下の `-` の箇条書きが、上のコードの**1行ずつの解説**です。
- 初めて出てくる Go の文法は、そこで説明します(後ろの章では説明を省略します)。
- 読む順番は「分からないものが少ない順」に並べています。前の章で説明済みのものは、後ろの章で使えるようになっています。
- 対象は、`cmd/server/`・`cmd/cli/`・`cmd/gui/` の実装コード(第1〜15章)と、ゲームのデータ `data/world.json`(第16章)、ビルドと設定(第17章)、テスト(第18章)です。**このリポジトリの全ファイルを説明しています**(冒頭の「全ファイル索引」を参照)。
- コード中の単独の `...` は省略の印です。完全な関数は対応するソースファイルを参照します。7-10の `handleXxx` は共通の流れを説明するための例です。

## 読む順番(全体の予定)

| 章 | ファイル | 内容 |
|---|---|---|
| 1 | `cmd/server/main.go` | サーバーの起動 |
| 2 | `cmd/server/room.go` | 部屋の型・隠し出口 |
| 3 | `cmd/server/player.go` | プレイヤーの型・HP回復 |
| 4 | `cmd/server/locale.go` | 日英切替 |
| 5 | `cmd/server/world.go` | 世界データの型・読み込み・検査 |
| 6 | `cmd/server/client_conn.go` | 接続ごとの送信キュー |
| 7 | `cmd/server/server.go` | サーバー本体・各コマンド |
| 8 | `hazard.go` `hardcore.go` `combat.go` `defeat.go` | 戦闘・危険・死亡・倒した状態 |
| 9 | `quest.go` `endings.go` `item_effects.go` `odyssey.go` | クエスト・エンディングと祝福・アイテム効果・最大HP |
| 10 | `chat.go` `group.go` | チャット・グループ |
| 11 | `notify.go` `flavor.go` | 通知文・実況文・ヒント |
| 12 | `player_store.go` `item_store.go` | セーブ |
| 13 | `logging.go` | ログ・不正検知 |
| 14 | `cmd/cli/main.go` | CLIクライアント |
| 15 | `cmd/gui/*.go`、`cmd/server/gui_state.go` | GUIクライアント |
| 16 | `data/world.json` | ゲームの内容(データ) |
| 17 | `Makefile` `go.mod` `.gitignore` `saves/` | ビルドと設定 |
| 18 | `*_test.go` | テスト全体 |

**読み方のおすすめ**:1章から順に読むのが安全です。時間が無ければ、まず **16章(データ)→ 7章(コマンドの骨格)→ 8章(戦闘)→ 9章(クエスト)** を読むと、ゲームの中心が分かります。

## 全ファイル索引

どのファイルが、どの章で説明されているかの一覧です(ファイル名から探すときに使ってください)。

| ファイル | 説明している場所 |
|---|---|
| `cmd/server/main.go` | 第1章 |
| `cmd/server/room.go` | 第2章 |
| `cmd/server/player.go` | 第3章 |
| `cmd/server/locale.go` | 第4章 |
| `cmd/server/world.go` | 第5章 |
| `cmd/server/client_conn.go` | 第6章 |
| `cmd/server/server.go` | 第7章 |
| `cmd/server/hazard.go` / `hardcore.go` | 8-1 / 8-2 |
| `cmd/server/combat.go` | 8-3〜8-6 |
| `cmd/server/defeat.go` | 8-7 |
| `cmd/server/odyssey.go` / `quest.go` / `endings.go` / `item_effects.go` | 9-1 / 9-2 / 9-3 / 9-4 |
| `cmd/server/chat.go` / `group.go` | 10-1 / 10-2 |
| `cmd/server/notify.go` / `flavor.go` | 11-1 / 11-2 |
| `cmd/server/player_store.go` / `item_store.go` | 12-1 / 12-2 |
| `cmd/server/logging.go` | 第13章 |
| `cmd/server/gui_state.go` | 15-18 |
| `cmd/cli/main.go` | 第14章 |
| `cmd/gui/protocol.go` / `model.go` / `retro.go` / `ui_locale.go` | 15-1 / 15-2 / 15-3 / 15-4 |
| `cmd/gui/art_assets.go` / `item_photos.go` | 15-5 / 15-6 |
| `cmd/gui/main.go` / `ui_layout.go` / `ui_actions.go` / `ui_journal.go` | 15-7 / 15-8 / 15-9 / 15-10 |
| `cmd/gui/ui_bars.go` / `ui_story.go` / `ui_effects.go` | 15-11 / 15-12 / 15-13 |
| `cmd/gui/ui_combat.go` / `ui_map.go` / `ui_endings.go` / `ui_item_effects.go` | 15-14 / 15-15 / 15-16 / 15-17 |
| `cmd/gui/assets/`(フォント・部屋・NPC・倒れた敵・アイテムの画像) | 15-3(フォント)、15-5(部屋・NPC・倒れた敵)、15-6(アイテム) |
| `data/world.json` | 第16章 |
| `Makefile` / `go.mod` / `go.sum` / `.gitignore` / `saves/` | 第17章 |
| `cmd/*/*_test.go`(42ファイル) | 第18章 |

※ `memo/ART_PROMPTS.md` と `README.md` はコードではないので、この解説の対象外です。
---

# 第1章 `cmd/server/main.go`(32行)

**役割**:サーバープログラムの入口。ポート4242で接続を待ち、人が来るたびに処理を別の流れ(goroutine)に任せます。

## 1-1 パッケージ宣言とimport

```go
package main

import (
	"net"
	"os"
)
```

- `package main`
  - このファイルが「`main` パッケージ」に属するという宣言です。Goのファイルは必ず先頭にこれを書きます。
  - `main` という名前だけは特別で、**実行できるプログラム**になります。他の名前(`package foo`)は部品(ライブラリ)になります。
  - 同じフォルダ(`cmd/server/`)の他の `.go` ファイルも全部 `package main` です。**同じパッケージのファイルは1つのプログラムとして合体する**ので、`main.go` から `server.go` の関数を `import` なしで呼べます。
- `import ( ... )`
  - 他のパッケージを使う宣言です。丸括弧の中に1行ずつ書きます。
- `"net"`
  - ネットワーク通信の標準ライブラリです。TCPで待ち受けるために使います。
- `"os"`
  - OSの機能を使う標準ライブラリです。ここでは「プロセス番号(pid)」を調べるために使います。
- 使わないimportがあると**コンパイルエラー**になります(Goの決まり)。逆に言うと、ここに書いてあるものは必ずどこかで使われています。

## 1-2 `main` 関数の始まり

```go
func main() {
	closeLog := setupLogging()
	defer closeLog()

	server := NewServer()
```

- `func main() {`
  - `func` は関数を定義する印です。`main` という名前の関数が、**実行プログラムの入口**になります。パッケージの変数や`init`関数の初期化が済んだあとに呼ばれます。
  - `()` は引数なし。`{` から対応する `}` までが関数の中身です。
- `closeLog := setupLogging()`
  - `setupLogging()` は `logging.go` にある関数で、ログの出力先を準備します。
  - `:=` は「**変数を新しく作って、同時に値を入れる**」書き方です。型は右側から自動で決まります(型推論)。関数の中でしか使えません。
  - この関数は「ログファイルを閉じる関数」を返します。それを `closeLog` という変数に入れています。**関数も値として変数に入れられる**のがGoの特徴です。
- `defer closeLog()`
  - `defer` は「**この関数(`main`)が終わるときに実行する**」という予約です。
  - `main`が通常の`return`で終わるときにログファイルを閉じます。`os.Exit`での終了では`defer`は実行されないので、下のエラー処理では`closeLog()`を先に呼びます。
  - 今はまだ実行されません。`main` の最後で実行されます。
- 空行は読みやすさのためだけです。
- `server := NewServer()`
  - `NewServer()` は `server.go` にある関数で、`*Server`(サーバー全体の状態)を作って返します。
  - 中で世界データ(`world.json`)を読み込んだり、前回のセーブを復元したりします。失敗したらそこでプログラムが終了します。
  - 戻り値を `server` という変数に入れます。

## 1-3 待ち受けの開始

```go
	listener, err := net.Listen("tcp", ":4242")
	if err != nil {
		closeLog()
		fatal("listen_failed", err)
	}
	defer listener.Close()
```

- `listener, err := net.Listen("tcp", ":4242")`
  - `net.Listen` は「TCPで待ち受けを始める」関数です。引数は(プロトコル, アドレス)です。
  - `":4242"` は「このコンピュータの**どのネットワークの口でも**、4242番ポートで待つ」という意味です。ホスト名を省略して、コロンとポート番号だけ書いています。
  - **戻り値が2つ**あります。Goの関数は複数の値を返せます。ここでは `listener`(待ち受けの道具)と `err`(エラー)です。
  - **Goにはtry/catchがありません**。失敗しうる関数は、最後の戻り値で `error` を返すのが決まりです。成功したときは `err` が `nil`(何も無い)になります。
- `if err != nil {`
  - 「エラーがあるなら」という意味です。`!=` は「等しくない」、`nil` は「何も無い」です。
  - **Goのコードで一番よく出てくる形**です。失敗しうる呼び出しの直後に必ずこれを書きます。
- `closeLog()`
  - この後 `fatal` でプログラムが強制終了します。強制終了すると `defer` が**実行されない**ので、ここで手動でログを閉じています。
- `fatal("listen_failed", err)`
  - `logging.go` の関数です。エラーをログに書いて `os.Exit(1)`(異常終了)します。
  - 4242番がすでに使われているときなどに、ここに来ます。
- `}`
  - `if` の中身の終わりです。
- `defer listener.Close()`
  - `main` が終わるときに、待ち受けを閉じる予約です。
  - 実際には下の `for` が無限ループなので、通常はここまで来ません。それでも「開いたものには `defer ... Close()` を付ける」のがGoの作法なので書いてあります。

## 1-4 起動ログ

```go
	logger.Info("server_started", "port", 4242, "pid", os.Getpid())
```

- `logger`
  - `logging.go` で定義されている、**パッケージ全体で共有するログ出力の道具**(グローバル変数)です。
- `.Info(...)`
  - 「情報レベル」のログを1件出します。
- 引数は `"server_started"` が**メッセージ**、残りは **キーと値のペア**(`"port"`→`4242`、`"pid"`→プロセス番号)です。
  - `log/slog` という標準ライブラリの書き方で、出力はJSONになります。例:`{"time":"...","level":"INFO","msg":"server_started","port":4242,"pid":12345}`
- `os.Getpid()`
  - 自分のプロセス番号を返します。ここで `os` パッケージが使われています。

## 1-5 接続を受け付けるループ

```go
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("accept_failed", "error", err.Error())
			continue
		}

		go server.handleClient(conn)
	}
}
```

- `for {`
  - 条件なしの `for` は**無限ループ**です。Goには `while` が無く、繰り返しは全部 `for` で書きます。
  - サーバーは終了するまで接続を待ち続けるので、無限ループになります。
- `conn, err := listener.Accept()`
  - `Accept()` は「**誰かが接続してくるまでここで待つ**」関数です(待っている間、この行から先に進みません)。
  - 接続が来ると、その人との通信用の `conn`(`net.Conn`型)が返ります。`conn` に書くと相手に届き、`conn` から読むと相手が送った内容が届きます。
  - ここの `:=` は、ループの中で毎回新しい `conn` と `err` を作っています(外側の `err` とは別物です)。
- `if err != nil {`
  - 接続の受け付けに失敗したとき(ネットワークの一時的な不具合など)です。
- `logger.Warn(...)`
  - 警告レベルのログです。`err.Error()` はエラーを**文字列**にします。
- `continue`
  - 「ループの残りを飛ばして、次の繰り返しへ」という意味です。ここでは、1回失敗してもサーバーを止めず、また `Accept()` に戻ります。
- `go server.handleClient(conn)`
  - **この1行がサーバーの要**です。
  - `go` を前に付けて関数を呼ぶと、その関数が**別の流れ(goroutine)で並行に実行**されます。呼んだ側(ここの `for`)は処理の終わりを待たず、すぐ次の行に進みます。
  - つまり「この人の対応は `handleClient` に任せて、自分はすぐ次の接続を待つ」という意味です。人が10人いれば、`handleClient` が10個、同時に動きます。
  - goroutine はOSのスレッドよりずっと軽いので、気軽に大量に作れます。
  - `server.handleClient` は `server.go` にあるメソッドです。`server` が**レシーバ**になって、関数の中で `s` として使われます。
- 最後の `}` が `for` の終わり、その次の `}` が `main` の終わりです。

### この章のまとめ

```
main()
 ├─ ログの準備
 ├─ NewServer()  … 世界データとセーブを読み込む
 ├─ net.Listen   … 4242番ポートで待ち受け開始
 └─ 無限ループ
     └─ Accept() で接続を待つ → 来たら go handleClient(conn) で別goroutineに任せる
```

次章以降は、この `handleClient` の中で使われる部品(部屋・プレイヤー・世界データ…)を順に見ていきます。

---

# 第2章 `cmd/server/room.go`(44行)

**役割**:「部屋」を表すデータの型と、「このプレイヤーに見える出口」を計算する関数を定義するファイルです。

## 2-1 全体

```go
package main

type RoomHazard struct {
	Type           string `json:"type"`
	RequiredItemID string `json:"required_item_id,omitempty"`
	CrewLoss       int    `json:"crew_loss,omitempty"`
	MinPartyTotal  int    `json:"min_party_total,omitempty"`
}

type Room struct {
	ID          string            `json:"id"`
	Name        LocalizedText     `json:"name"`
	Description LocalizedText     `json:"description"`
	Exits       map[string]string `json:"exits"`
	// SecretExits are exits that only exist for players who have reached all the listed endings.
	SecretExits map[string]SecretExit `json:"secret_exits,omitempty"`
	Hazard      *RoomHazard           `json:"hazard,omitempty"`
}

// SecretExit is an exit that opens once the player has reached every ending in RequiresEndings.
type SecretExit struct {
	Room            string   `json:"room"`
	RequiresEndings []string `json:"requires_endings"`
}

// exitsFor returns the exits this player can see and use: the normal ones plus any secret exit they have unlocked.
func (r *Room) exitsFor(player *Player) map[string]string {
	...
}
```

## 2-2 `struct`(構造体)とは

- `type RoomHazard struct { ... }`
  - `type 名前 struct { フィールド... }` で、**複数のデータをひとまとめにした新しい型**を作ります。他の言語の「クラスのデータ部分」にあたります。
  - 中に書く `Type string` は「`Type` という名前の、文字列型のフィールド」という意味です(名前が先、型が後)。
  - 名前が**大文字で始まる**フィールドは、パッケージの外から見えます。また、後で出てくる **JSONの変換は、大文字で始まるフィールドしか対象にしません**。ここのフィールドが全部大文字始まりなのは、この理由が大きいです。

## 2-3 `RoomHazard`(部屋の危険)の各行

```go
	Type           string `json:"type"`
```

- `Type string`
  - 危険の種類を表す文字列です。値は `world.json` の中で `"lethal"`(入ると死ぬ)、`"item_gate"`(アイテムが無いと死ぬ)、`"crew_gate"`、`"crew_cost"` のどれかです。
- `` `json:"type"` ``
  - バッククォート(`` ` ``)で囲んだ部分は**構造体タグ**です。コードの動作自体は変えませんが、`encoding/json` ライブラリが読み取ります。
  - 意味は「JSONでは、このフィールドのキー名は `type`(小文字)」です。Goのフィールド名は `Type`(大文字)でも、JSONの `"type": ...` と結び付けられます。

```go
	RequiredItemID string `json:"required_item_id,omitempty"`
	CrewLoss       int    `json:"crew_loss,omitempty"`
	MinPartyTotal  int    `json:"min_party_total,omitempty"`
```

- `RequiredItemID string`
  - `item_gate` の部屋で、**通過に必要なアイテムのID**です。
- `CrewLoss int`
  - `crew_gate` / `crew_cost` で、**減る仲間の人数**です。
- `MinPartyTotal int`
  - `crew_gate` で、**通過に必要な最低人数**(自分を含む)です。
- `,omitempty`
  - 「値が**空(ゼロ値)なら、JSONに書き出すとき省略する**」という指定です。読み込みでは影響しません。
  - 危険の種類によって使うフィールドが違うので、使わないものは省略されます。

## 2-4 `Room`(部屋)の各行

```go
type Room struct {
	ID          string            `json:"id"`
	Name        LocalizedText     `json:"name"`
	Description LocalizedText     `json:"description"`
	Exits       map[string]string `json:"exits"`
	SecretExits map[string]SecretExit `json:"secret_exits,omitempty"`
	Hazard      *RoomHazard           `json:"hazard,omitempty"`
}
```

- `ID string`
  - 部屋のID。例:`"loc.hall_of_fates"`。
- `Name LocalizedText`
  - 部屋の名前。`LocalizedText` は第4章で説明します。`{"en": "Hall of the Fates", "ja": "運命の間"}` のように**言語ごとの文字列**を持つ型です。
- `Description LocalizedText`
  - 部屋の説明文。同じく言語別です。
- `Exits map[string]string`
  - 出口の一覧です。`map[string]string` は「**文字列をキーにして文字列を引く辞書**」です。
  - 例:`{"north": "loc.argo_iolcus", "east": "loc.ody_troy_shore"}`。つまり「north に進むと `loc.argo_iolcus` の部屋に行く」です。
  - `MOVE east` と打ったとき、サーバーは `"east"` をキーにして行き先を引きます(正確には、2-6 の `exitsFor` を通した出口の辞書から引きます)。
- `SecretExits map[string]SecretExit`
  - **条件を満たした人にだけ現れる出口**です。キー(方角)は `Exits` と同じ形ですが、値は行き先の部屋に加えて「必要なエンディング」を持つ `SecretExit` です。
  - 運命の間の南の出口(最終エンディングの後に開く「織られざる機」への道)がこれです。
  - `omitempty` なので、隠し出口の無い部屋ではJSONに書き出しません。
- `Hazard *RoomHazard`
  - 部屋の危険です。型の前の `*` は**ポインタ**(「`RoomHazard` そのものではなく、その置き場所を指す矢印」)です。
  - ポインタには「**何も指していない状態 = `nil`**」があります。そのため「**危険が無い部屋は `nil`**」と表せます。
  - ポインタにせず `RoomHazard` をそのまま持つと、危険の無い部屋も「種類が空の危険」を持つことになり、「あるか無いか」を区別しにくくなります。
  - `omitempty` は、`nil` のときJSONに書き出さない、という意味にもなります。
  - 使う側は `if room.Hazard == nil { 危険なし }` と書きます(`hazard.go` で出てきます)。

## 2-5 `SecretExit`(隠し出口)

```go
type SecretExit struct {
	Room            string   `json:"room"`
	RequiresEndings []string `json:"requires_endings"`
}
```

- `Room string`
  - 行き先の部屋のID です。
- `RequiresEndings []string`
  - この出口が開くために**達成していなければならないエンディングのID**の一覧です。`[]string` は「文字列のスライス(長さが変えられる配列)」です。
  - 運命の間の設定では `["ending.final"]` の1つだけです。複数書けば、全部を達成した人にだけ開きます(AND条件)。

`world.json` での書き方はこうです。

```json
"loc.hall_of_fates": {
  "exits": {"west": "loc.argo_iolcus", "north": "loc.troy_ida", "east": "loc.ody_troy_shore"},
  "secret_exits": {
    "south": {"room": "loc.unwoven_loom", "requires_endings": ["ending.final"]}
  }
}
```

## 2-6 `exitsFor`— このプレイヤーに見える出口

```go
func (r *Room) exitsFor(player *Player) map[string]string {
	exits := make(map[string]string, len(r.Exits)+len(r.SecretExits))
	for dir, dest := range r.Exits {
		exits[dir] = dest
	}
	for dir, secret := range r.SecretExits {
		unlocked := true
		for _, endingID := range secret.RequiresEndings {
			if !player.Endings[endingID] {
				unlocked = false
			}
		}
		if unlocked {
			exits[dir] = secret.Room
		}
	}
	return exits
}
```

- `func (r *Room) exitsFor(player *Player) map[string]string {`
  - `func (r *Room) 名前(...)` は**メソッド**です。「`Room` に付いた関数」で、`room.exitsFor(player)` のように呼びます。`r` は呼び出された `Room` 自身(他の言語の `this`/`self`)です。`*Room` とポインタにしているのは、部屋のデータを丸ごとコピーせずに済むからです。
  - 引数は「見る人」である `player` で、戻り値は「方角 → 行き先の部屋ID」の辞書です。
  - メソッド名が小文字で始まる(`exitsFor`)のは、パッケージの外に公開しない、という意味です。
- `exits := make(map[string]string, len(r.Exits)+len(r.SecretExits))`
  - `make` で空の辞書を作ります。2つ目の引数は「だいたいこのくらいの数を入れる」という**大きさの目安**で、速度のための指定です(無くても動きます)。
  - **新しい辞書を作る**のがポイントです。`r.Exits` を直接書き換えると、1人が隠し出口を開けただけで**全員の部屋データが変わってしまう**からです。
- `for dir, dest := range r.Exits { exits[dir] = dest }`
  - 普通の出口を、全部そのまま新しい辞書に写します。`range` は辞書を1件ずつ取り出すループで、`dir` が方角、`dest` が行き先です。
- `for dir, secret := range r.SecretExits {`
  - 次に、隠し出口を1件ずつ調べます。
- `unlocked := true` と内側の `for`
  - 「開いている」と仮定して、必要なエンディングを1つずつ確認します。`player.Endings` は「達成したエンディングID → `true`」の辞書(第3章)で、**書かれていないキーを引くと `false`** になる(ゼロ値)ので、未達成のIDは自然に `false` になります。
  - 1つでも未達成なら `unlocked = false` にします。
- `if unlocked { exits[dir] = secret.Room }`
  - 全部達成していたら、その方角に行き先を足します。
- `return exits`
  - 組み立てた辞書を返します。

この関数は `LOOK`(見える出口を返す)と `MOVE`(進めるか判定する)の**両方が同じ答えを使う**ために、1か所にまとめてあります。片方だけ直し忘れて「見えるのに進めない」「見えないのに進める」にならないようにするためです(7-12、7-13)。

### この章のまとめ
`world.json` の1つの部屋が、この `Room` 1つに読み込まれます。

```json
"loc.argo_crete": {
  "id": "loc.argo_crete",
  "name": {"en": "...", "ja": "..."},
  "exits": {"north": "loc.argo_iolcus"},
  "hazard": {"type": "crew_cost", "crew_loss": 2}
}
```

---

# 第3章 `cmd/server/player.go`(95行)

**役割**:プレイヤー1人分のデータ(`Player`)と、それに対する小さな操作を定義します。`Player` は**そのままセーブファイルに書き出される**型です。

## 3-1 パッケージとimport

```go
package main

import "time"
```

- `import "time"`
  - importが1つだけのときは、丸括弧なしで1行に書けます。
  - `time` は時刻・時間の標準ライブラリです。HP自動回復の時刻計算に使います。

## 3-2 クエスト進行の型

```go
type PlayerQuest struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}
```

- 1つのクエストについて、プレイヤーが持つ状態です。
- `Status`:`"active"`(進行中)または `"completed"`(達成済み)。
- `Progress`:進捗の数。たとえば「3個集める」クエストで今2個なら `2`。

## 3-3 `Player` 構造体

```go
type Player struct {
	Name             string                  `json:"name"`
	HP               int                     `json:"hp"`
	MaxHPBonus       int                     `json:"max_hp_bonus,omitempty"` // earned from quests; kept after death
	RoomID           string                  `json:"room_id"`
	Inventory        []string                `json:"inventory"`
```

- `Name`:プレイヤー名。
- `HP`:今の体力。
- `MaxHPBonus int`
  - **クエストで増えた最大HPの合計**です。最大HPの基本は100で、クエストを達成するたびにここへ足されます(第9章の9-2)。
  - **死んでも減りません**。死亡のときに書き換えられるのは `HP` や持ち物であって、この値には触らないからです。セーブにも書かれるので、再接続しても残ります。
  - 実際の最大HPは「100 + `MaxHPBonus` + 持っているアイテムの効果」で、`item_effects.go` の `maxHPLocked` が計算します(第9章の9-4)。`omitempty` なので、0のあいだはセーブに出ません。
- `RoomID`:今いる部屋のID。
- `Inventory []string`
  - 持ち物のIDの一覧です。`[]string` は**スライス**(長さが変えられる配列)で「文字列の並び」です。
  - 例:`["item.olive_stake", "item.beeswax"]`。

```go
	Crew            int                     `json:"crew,omitempty"`
	CrewInitialized bool                    `json:"crew_initialized,omitempty"`
	IntroSeen       bool                    `json:"intro_seen,omitempty"`
	LastDeathSubject string                 `json:"last_death_subject,omitempty"`
	CombatTargetID  string                  `json:"combat_target_id,omitempty"`
```

- `Crew`:オデュッセイア編の「仲間の人数」。
- `CrewInitialized`:仲間を付与済みか。`bool` は `true`/`false` の型です。
- `IntroSeen`:導入の案内文をもう見たか。
- `LastDeathSubject`:**直前の死因になったもののID**(敵・部屋・アイテム)。死んだあとにモイライへ `TALK` すると、これを手がかりに神話のヒントを1回だけ返します(`notify.go`、第11章)。
- `CombatTargetID`:今戦っている敵のID。**空文字なら戦闘中ではない**。

```go
	FledFrom        map[string]bool         `json:"fled_from,omitempty"`
	Quests          map[string]*PlayerQuest `json:"quests,omitempty"`
	Endings         map[string]bool         `json:"endings,omitempty"`
	EnemyHP         map[string]int          `json:"enemy_hp,omitempty"`
```

- `FledFrom map[string]bool`
  - 「どの敵から逃げ切ったか」の記録。キーが敵のID、値が `true`。
  - 逃げ切った敵は、もう出口を塞がなくなります。
- `Quests map[string]*PlayerQuest`
  - クエストIDから、そのクエストの状態を引く辞書。値が `*PlayerQuest`(ポインタ)なので、取り出して書き換えれば辞書の中身も変わります。
- `Endings map[string]bool`
  - 到達したエンディングのID。
- `EnemyHP map[string]int`
  - **敵のIDから、このプレイヤーから見た敵の残りHPを引く**辞書。
  - 敵のHPが**プレイヤーごと**に別々なので、他の人が倒した敵が、自分の世界ではまだ生きている、という仕様になります。

```go
	lastRegen        time.Time
	guarding         bool // braced with DEFEND: the next counter-attack is halved (not saved)
	exiting          bool
}
```

- `lastRegen time.Time`
  - 最後にHP回復を計算した時刻です。
- `guarding bool`
  - `DEFEND` で身構えている最中か。`true` の間は**次の反撃が半分**になり、反撃を1回受けると `false` に戻ります(`combat.go` の `takeGuard`、第8章)。
- `exiting bool`
  - 退室処理の最中かどうかです。
- この3つは**名前が小文字で始まる**ので、JSONに変換されません(=セーブされません)。しかも `json:"..."` のタグもありません。メモリ上だけで使う一時的な状態です。

## 3-4 `hasItem` メソッド

```go
func (p *Player) hasItem(itemID string) bool {
	for _, id := range p.Inventory {
		if id == itemID {
			return true
		}
	}
	return false
}
```

- `func (p *Player) hasItem(itemID string) bool {`
  - これが**メソッド**です。`func` と関数名の間の `(p *Player)` が**レシーバ**で、「この関数は `Player` に属し、中では呼び出した本人を `p` と呼ぶ」という意味です。
  - `*` が付くのは、コピーではなく**本物の `Player`** を指すためです。
  - `(itemID string)` が引数、その後ろの `bool` が**戻り値の型**です。
  - 呼ぶときは `player.hasItem("item.olive_stake")` と書きます。
  - 名前が小文字始まりの `hasItem` は、パッケージの外からは見えません(使うのはこのプログラム内だけなので問題なし)。
- `for _, id := range p.Inventory {`
  - `range` は「スライスや辞書を**1つずつ順に取り出す**」ための書き方です。
  - スライスに `range` を使うと、毎回(**番号**, **値**)の2つが返ります。
  - `_`(アンダースコア)は「**この値は使わない**」という印です。番号は要らないので捨てています。Goは使わない変数があるとコンパイルエラーになるので、このように `_` で明示します。
  - `id` に持ち物のIDが1つずつ入ります。
- `if id == itemID {`
  - 探しているIDと一致したら、
- `return true`
  - 「持っている」と答えて、関数を**その場で終わります**。
- 最後の `return false`
  - ループを最後まで回っても見つからなかったので「持っていない」と答えます。

## 3-5 `meetsMythRequirement`

「**神話の前提条件**(攻撃や会話の前に必要なアイテム・クエスト)を満たしているか」を調べます。満たしていないと攻撃で即死するゲームの核の部分です。

```go
func (p *Player) meetsMythRequirement(npc *NPC) bool {
	if npc.MythRequirementItem != "" && !p.hasItem(npc.MythRequirementItem) {
		return false
	}
```

- 引数 `npc *NPC` は調べる相手のNPC(ポインタ)です。`NPC` 型は `world.go` にあります。
- `npc.MythRequirementItem != ""`
  - 必要アイテムの指定が**ある**か。空文字なら「要らない」という意味です。
- `&&`
  - 「かつ」です。左が偽ならGoは右を評価しません(短絡評価)。
- `!p.hasItem(...)`
  - `!` は「ではない」。つまり「そのアイテムを**持っていない**」。
- つまり「必要アイテムが指定されていて、しかも持っていないなら」`false`(満たしていない)を返します。

```go
	if npc.MythRequirementQuest != "" {
		state := p.Quests[npc.MythRequirementQuest]
		if state == nil || state.Status != "completed" {
			return false
		}
	}
	return true
}
```

- 必要クエストの指定があるとき:
- `state := p.Quests[npc.MythRequirementQuest]`
  - 辞書 `Quests` から、そのクエストの状態を引きます。**辞書に無いキーを引くとゼロ値**(ポインタなら `nil`)が返ります。
- `state == nil || state.Status != "completed"`
  - 「一度も受けていない(`nil`)、または達成していない」なら `false`。
  - `||` は「または」。左が真なら右は評価されないので、`nil` のときに `state.Status` を読んでエラーになることはありません。
- どの条件にも引っかからなければ `return true`(満たしている)。
- 前提条件が**何も指定されていないNPC**は、2つの `if` がどちらも飛ばされて `true` になります。

## 3-6 `hasMythRequirement`

```go
func (npc *NPC) hasMythRequirement() bool {
	return npc.MythRequirementItem != "" || npc.MythRequirementQuest != ""
}
```

- これも**メソッド**ですが、レシーバが `*NPC` です。**メソッドは、同じパッケージの中なら、別のファイルにある型にも定義できます**(`NPC` 型は `world.go` にあります)。
- 「そもそも前提条件が設定されているNPCか」を返します。アイテム条件かクエスト条件のどちらかがあれば `true`。
- 先ほどの `meetsMythRequirement` は「条件を満たしているか」、これは「条件が存在するか」で、役割が別です。

## 3-7 敵HPの取得と設定

```go
func (p *Player) enemyHP(npcID string, npc *NPC) int {
	if hp, ok := p.EnemyHP[npcID]; ok {
		return hp
	}
	return npc.HP
}
```

- `if hp, ok := p.EnemyHP[npcID]; ok {`
  - 辞書を引くときは `値, ok := 辞書[キー]` と**2つ受け取れます**。`ok` は「キーが**存在したか**」の `bool` です。
  - `if 準備文; 条件 {` という形です。`;` の前で変数を作り、後ろでそれを条件にしています。この `hp` と `ok` は、この `if` の中だけで使えます。
  - プレイヤー別のHPが**記録されていれば**それを返します。
- `return npc.HP`
  - 記録が無ければ、まだ一度も攻撃していないので**NPCの初期HP**を返します。

```go
func (p *Player) setEnemyHP(npcID string, hp int) {
	if p.EnemyHP == nil {
		p.EnemyHP = make(map[string]int)
	}
	p.EnemyHP[npcID] = hp
}
```

- この関数は戻り値がありません(`)` の後ろに型が無い)。
- `if p.EnemyHP == nil {`
  - **`nil` の辞書には書き込めません**(実行時にパニック=異常終了します)。読むのは安全ですが、書くのは危険です。
  - `EnemyHP` は最初は `nil` なので、書く前に確認します。
- `make(map[string]int)`
  - **空の辞書を作る**組み込み関数です。
- `p.EnemyHP[npcID] = hp`
  - 辞書に書き込みます。

## 3-8 HP自動回復

```go
const (
	regenInterval = 2 * time.Second
	regenAmount   = 1
)
```

- `const ( ... )`
  - **定数**(あとで変えられない値)をまとめて宣言します。
- `regenInterval = 2 * time.Second`
  - 回復の間隔は2秒。`time.Second` は「1秒」を表す定数で、掛け算で時間の長さを作れます。
- `regenAmount = 1`
  - 1回の回復量は1HPです。

```go
func (p *Player) regenLocked(now time.Time, bonus, maxHP int) {
```

- 名前の最後の `Locked` は、このプロジェクトの決まりで「**呼ぶ前にサーバーのロック(`s.mu`)を取っていること**」を意味します(並行処理の話は第7章で詳しく)。
- 引数 `now` は「今の時刻」です。時刻を引数で受け取る形にすると、テストで好きな時刻を渡せます。
- `bonus, maxHP int` は「`bonus int, maxHP int`」の省略形です。同じ型の引数は、型を最後に1回だけ書けます。
  - `bonus`:**1回の回復で追加するHP**。ヘラの祝福や、持っているアイテムの効果の合計です(マイナスもあり得ます)。
  - `maxHP`:**このプレイヤーの今の最大HP**。以前は定数 `maxPlayerHP`(100)を直接使っていましたが、最大HPがクエストやアイテムで変わるようになったので、呼ぶ側が計算して渡します。

```go
	if p.HP > maxHP {
		p.HP = maxHP // an item that raised max HP was dropped
	}
	if p.HP >= maxHP || p.lastRegen.IsZero() {
		p.lastRegen = now
		return
	}
```

- 最初の `if` は**はみ出しの修正**です。最大HPを上げるアイテムを持っていて、HPがそのぶん多かったとき、アイテムを置くと最大HPが下がります。そのときHPが最大HPを超えたままにならないよう、ここで切り詰めます。
- 2つ目の `if`:「すでに満タン」または「`lastRegen` が一度も設定されていない(`IsZero()`)」とき:
  - 基準時刻を今に更新して、何もせず `return` で終わります。
  - 満タンの間は基準時刻を更新し続けるので、「ダメージを受けた瞬間」から回復の時間が数え始められます。

```go
	ticks := int(now.Sub(p.lastRegen) / regenInterval)
	if ticks <= 0 {
		return
	}
```

- `now.Sub(p.lastRegen)`
  - 「今 − 前回の時刻」で、経過時間(`time.Duration`)を出します。
- `/ regenInterval`
  - 経過時間を2秒で割ります。たとえば5秒経っていたら「2.5」ですが、整数の割り算になるので「**2回分**」になります。
- `int(...)`
  - 型を `int`(整数)に変換します。Goでは型の変換は**必ず明示的に**書きます。
- `ticks <= 0`
  - まだ2秒経っていないので、何もせず終わります。

```go
	p.HP += ticks * max(0, regenAmount+bonus)
	if p.HP >= maxHP {
		p.HP = maxHP
		p.lastRegen = now
		return
	}
	p.lastRegen = p.lastRegen.Add(time.Duration(ticks) * regenInterval)
}
```

- `p.HP += ticks * max(0, regenAmount+bonus)`
  - 経過した回数分だけ回復します(`+=` は「足して代入」)。
  - 1回の回復量は `regenAmount + bonus` ですが、`max(0, ...)` で**0より小さくならない**ようにしています。悪いアイテム(回復速度マイナス)を持っていても、HPが**減る**ことはなく、回復が止まるだけです。`max` は2つの値の大きいほうを返す、Goの組み込み関数です。
- 上限(`maxHP`)を超えたら切り詰めて、基準時刻を今にして終わります。
- 最後の行は、満タンにならなかったときの基準時刻の更新です。
  - **`now` にしない**のがポイントです。もし `now` にすると、5秒経過で2回回復したとき、端数の1秒が捨てられます。
  - 代わりに「前回の基準 + 回復した回数分の時間」だけ進めるので、端数の1秒が次回に持ち越されます。
  - `time.Duration(ticks)` は `int` を時間の型に変換しています。

### この章のまとめ
- HP回復には**タイマーもgoroutineも使っていません**。コマンドが来たときに「前回から何回分経ったか」をまとめて計算します。人が何もしていない間は、計算コストがゼロです。

---

# 第4章 `cmd/server/locale.go`(62行)

**役割**:日本語・英語の切り替えです。

## 4-1 パッケージとimport

```go
package main

import (
	"fmt"
	"net"
	"strings"
)
```

- `fmt`:文字列の整形や出力(`Fprintln`)に使います。
- `net`:`net.Conn` の型を使います。
- `strings`:大文字・小文字の変換に使います。

## 4-2 `LocalizedText`

```go
const defaultLocale = "en"

type LocalizedText map[string]string
```

- `defaultLocale`:対応する言語が無いときに使う言語は英語。
- `type LocalizedText map[string]string`
  - **既存の型に別名(新しい型名)を付けて、独自のメソッドを定義できる**書き方です。
  - 中身は「文字列→文字列の辞書」で、`{"en": "Hello", "ja": "こんにちは"}` のような**言語コードから本文を引く辞書**です。
  - 専用の型にしたのは、`Get` や `Format` というメソッドを付けるためです。

```go
func (t LocalizedText) Get(locale string) string {
	if s, ok := t[locale]; ok && s != "" {
		return s
	}
	return t[defaultLocale]
}
```

- レシーバが `(t LocalizedText)` です。辞書型は**もともと参照のような型**なので、`*` を付けずに受け取って構いません。
- `if s, ok := t[locale]; ok && s != "" {`
  - その言語の文字列があり(`ok`)、かつ空でないなら、
- `return s`
  - それを返します。
- `return t[defaultLocale]`
  - 無ければ**英語を返す**(英語も無ければ空文字)。
- 日本語版が未翻訳でも、英語が出るので画面が空になりません。

## 4-3 対応言語とクライアントの言語

```go
var supportedLocales = map[string]bool{"en": true, "ja": true}
```

- `var 名前 = 値` で、パッケージ全体で使える変数を作ります。
- `map[string]bool{"en": true, "ja": true}` は**辞書のリテラル**(中身を書いて作る形)です。
- `supportedLocales["ja"]` は `true`、`supportedLocales["fr"]` は(無いキーなのでゼロ値の)`false`。**「このコードは対応言語か」を調べる定番の書き方**です。

```go
func clientLocale(conn net.Conn) string {
	if client, ok := conn.(*serverClient); ok && client.locale != "" {
		return client.locale
	}
	return defaultLocale
}
```

- 引数 `conn net.Conn` は、通信の口です。`net.Conn` は**インターフェース**(「読み書きできるもの」という約束事)で、中身がどんな型でも渡せます。
- `conn.(*serverClient)`
  - **型アサーション**:「この `conn` の正体は `*serverClient` のはずだ」と言って、取り出します。
  - `値, ok :=` の形で受け取ると、違っていても**パニックにならず `ok` が `false`** になります。
- 正体が `*serverClient` で、かつ言語が設定されていれば、その言語を返します。そうでなければ英語。
  - テストでは、`serverClient` ではない偽の接続を渡すことがあるので、このように安全な書き方をしています。

## 4-4 LOOKで返す部屋データ

```go
type roomView struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}
```

- `LOOK` の応答に入れる、**1つの言語に決めた部屋の表示用データ**です。
- `Room`(第2章)との違い:`Name` と `Description` が `LocalizedText`(全言語入り)ではなく、ただの `string`(選んだ言語の1つ)です。クライアントに全言語を送る必要は無いので、変換してから送ります。

```go
func newRoomView(room *Room, locale string) roomView {
	exits := make(map[string]string, len(room.Exits))
	for dir, dest := range room.Exits {
		exits[dir] = dest
	}
	return roomView{ID: room.ID, Name: room.Name.Get(locale), Description: room.Description.Get(locale), Exits: exits}
}
```

- `make(map[string]string, len(room.Exits))`
  - 辞書を作ります。2つ目の引数は**あらかじめ確保する大きさ**(省略可。速度のための工夫です)。
- `for dir, dest := range room.Exits {`
  - 辞書に `range` を使うと、(**キー**, **値**)が1組ずつ取り出されます(順序はランダム)。
- `exits[dir] = dest`
  - 出口の辞書を**コピー**します。元の `room.Exits` をそのまま渡すと、送信側が書き換えたとき世界のデータが壊れかねないので、コピーして渡します。
- `return roomView{ID: room.ID, ...}`
  - 構造体を作って返します。`フィールド名: 値` の形でフィールドを指定します。
  - `room.Name.Get(locale)` で、選ばれた言語の文字列を取り出しています。

## 4-5 `LANG` コマンド

```go
func handleLang(s *Server, conn net.Conn, name *string, parts []string) bool {
```

- コマンドを処理する関数(**ハンドラ**)の基本の形です。全ハンドラがこの同じ引数・戻り値の形をしています。
  - `s *Server`:サーバー本体
  - `conn net.Conn`:この人との通信の口
  - `name *string`:この接続のプレイヤー名への**ポインタ**(CONNECT前は空文字)
  - `parts []string`:コマンドを空白で区切った単語の並び(例:`["LANG","ja"]`)
  - 戻り値の `bool`:**この接続を終了するか**(`true`なら終了)

```go
	if !requireExactArgs(conn, parts, 2) {
		return false
	}
```

- `requireExactArgs` は(`server.go` で定義)引数の数をちょうどに検査する関数で、違えば `ERR 400 BAD_REQUEST` を送って `false` を返します。
- `!` を付けて「検査に**失敗したら**」`return false`(接続は続けたまま、このコマンドの処理だけ終了)です。

```go
	if *name != "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
```

- `*name`
  - `name` はポインタなので、**ポインタの先の値**を読むには `*` を付けます(間接参照)。`""` でない = すでに `CONNECT` 済み。
  - `LANG` は**CONNECTの前にしか使えない**ので、済みならエラーです。
- `fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")`
  - `conn` に1行(末尾に改行付き)を書き込みます。これで相手に応答が届きます。

```go
	code := strings.ToLower(parts[1])
	if !supportedLocales[code] {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
```

- `parts[1]`
  - スライスの**添字**は0から始まります。`parts[0]` が `"LANG"`、`parts[1]` が言語コード。
  - 前の検査で単語がちょうど2つと確認済みなので、`parts[1]` は安全に読めます。
- `strings.ToLower(...)` は小文字化(`JA` も受け付けるため)。
- `!supportedLocales[code]`
  - 対応していない言語(辞書に無い→`false`)ならエラー。

```go
	if client, ok := conn.(*serverClient); ok {
		client.locale = code
	}
	fmt.Fprintln(conn, "OK lang="+code)
	return false
}
```

- 型アサーションで `*serverClient` を取り出し、**その接続の言語を保存**します(以降の応答はその言語になる)。
- `"OK lang="+code`
  - `+` で文字列をつなげます。
- `return false`
  - 接続は続けるので `false`。

---

# 第5章 `cmd/server/world.go`(247行)

**役割**:`data/world.json`(世界のデータ)を読み込む型と、読み込んだデータが**矛盾していないか検査する**処理です。サーバー起動時に1回だけ使われます。

## 5-1 import

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)
```

- `encoding/json`:JSONと構造体の相互変換。
- `fmt`:`fmt.Errorf` でエラーメッセージを作る。
- `os`:ファイルを読む(`os.ReadFile`)。
- `strings`:大文字小文字を無視した比較(`strings.EqualFold`)。

## 5-2 `Item`(アイテム)の型

```go
type Item struct {
	Name        LocalizedText `json:"name"`
	Description LocalizedText `json:"description"`
	RoomID      string        `json:"room_id"`
	Obtainable  bool          `json:"obtainable"`
	Renewable   bool          `json:"renewable,omitempty"`
	RewardOnly  bool          `json:"reward_only,omitempty"`
	Effects     []ItemEffect  `json:"effects,omitempty"`

	HomeRoomID string `json:"-"`
}
```

- `Name` / `Description`:名前と説明(言語別)。
- `RoomID`:**今、そのアイテムがある部屋のID**。誰かが拾うと `""` になる(どの部屋にも無い)。
- `Obtainable`:`TAKE` で取れるか。
- `Renewable`:`true` なら**取っても部屋から無くならない**(何度でも手に入る)。
- `RewardOnly`:`true` なら、エンディングの報酬としてだけ手に入る記念品(部屋には置かれない)。
- `Effects []ItemEffect`
  - **持っている間だけ効く効果**の一覧です。`ItemEffect`(種類と値のペア)は `item_effects.go`(9-4)で定義します。
  - 全アイテムに設定されていて、例えば `{"effect": "damage_bonus", "value": 3}` なら「与えるダメージ+3」、`{"effect": "max_hp", "value": -5}` なら「最大HP-5」です。**値がマイナスなら悪い効果**です。
  - `omitempty` なので、効果の無いアイテムはJSONに書かなくて構いません。
- 空行を挟んで `HomeRoomID string`:**元の置き場所**。死んで持ち物を失ったときに、ここへ戻す。
- `` `json:"-"` ``:タグの `-` は「**このフィールドはJSONに出し入れしない**」という意味。`world.json` には書かれておらず、読み込み後にプログラムが設定する(5-6参照)。

## 5-3 `NPC` の型

```go
type NPC struct {
	Name                 LocalizedText   `json:"name"`
	Description          LocalizedText   `json:"description"`
	Role                 string          `json:"role"`
	RoomID               string          `json:"room_id"`
	HP                   int             `json:"hp"`
	Dialogue             []LocalizedText `json:"dialogue"`
	DialogueCleared      []LocalizedText `json:"dialogue_cleared,omitempty"` // said instead once the room's enemies are beaten
```

- `Role`:役割。`"enemy"`(戦える敵)、`"quest_giver"`、`"dialogue"` のどれか。
- `RoomID`:いる部屋。
- `HP`:敵の初期HP。
- `Dialogue []LocalizedText`:**台詞のリスト**。`[]LocalizedText` は「LocalizedTextのスライス」。`TALK` では先頭(`Dialogue[0]`)を返し、ガイドNPCは続きも順に送る。
- `DialogueCleared []LocalizedText`
  - **その部屋の敵を全部倒した後に話す台詞**です。設定されていて、かつ部屋が「突破済み」のプレイヤーには、`Dialogue` の代わりにこちらを返します(`defeat.go` の `dialogueFor`、第8章の8-7)。
  - 敵を倒す前と後で、ピネウス・メデイア・ヘレネー・囚われの水夫・ペネロペイアの言葉が変わります。

```go
	MythRequirementItem  string          `json:"myth_requirement_item,omitempty"`
	MythRequirementQuest string          `json:"myth_requirement_quest,omitempty"`
	FleeAccurate         bool            `json:"flee_accurate,omitempty"`
	FleeSucceedsOnce     bool            `json:"flee_succeeds_once,omitempty"`
	Unwinnable           bool            `json:"unwinnable,omitempty"`
	Mighty               bool            `json:"mighty,omitempty"` // a non-enemy so powerful that attacking it is instant death
	CrewLossOnAttack     int             `json:"crew_loss_on_attack,omitempty"`
	Guide                bool            `json:"guide,omitempty"`
	Ending               *Ending         `json:"ending,omitempty"`
}
```

- `MythRequirementItem` / `MythRequirementQuest`:この敵を攻撃(や会話)する前に**必要なアイテム/達成済みクエスト**。無いまま挑むと即死(ゲームの核)。
- `FleeAccurate`:`FLEE` が**必ず成功**する。
- `FleeSucceedsOnce`:`FLEE` が**最初の1回だけ成功**する。
- `Unwinnable`:**勝てない敵**。攻撃すると仲間が減るだけ。
- `Mighty`:**強すぎて手を出せない敵以外のNPC**(神・魔女など)。`true` のNPCを攻撃すると、反撃を待たずに**一撃で死にます**(第8章の8-5)。敵以外のNPCにしか意味がありません。
- `CrewLossOnAttack`:勝てない敵を攻撃したとき減る仲間の数。
- `Guide`:新人向けの案内役か。
- `Ending *Ending`:このNPCと話すと迎えられるエンディング。ポインタなので、**エンディングが無いNPCは `nil`**。

## 5-4 クエストの型

```go
type QuestObjective struct {
	Type     string `json:"type"`
	TargetID string `json:"target_id"`
	Count    int    `json:"count"`
}

type QuestReward struct {
	HP int `json:"hp"`
}

type Quest struct {
	Name        LocalizedText  `json:"name"`
	Description LocalizedText  `json:"description"`
	GiverNPCID  string         `json:"giver_npc_id"`
	Objective   QuestObjective `json:"objective"`
	Reward      QuestReward    `json:"reward"`
}
```

- `QuestObjective`(目的):
  - `Type`:`"collect_item"`(アイテムを集める)か `"defeat_npc"`(敵を倒す)。
  - `TargetID`:集めるアイテム/倒す敵のID。
  - `Count`:必要な数。
- `QuestReward`(報酬):`HP` は報酬の大きさを表す数です。**HPを直接回復するのではなく、最大HPの上昇量の元になります**(上昇量は `HP / 5`、最低1。第9章の9-2)。JSONの名前が `hp` のままなのは、もとの設計の名残です。
- `Quest`:`GiverNPCID` は依頼者NPCのID。`Objective` と `Reward` は**構造体を丸ごと中に持つ**(ポインタではない)。

## 5-5 `World`(世界全体)

```go
type World struct {
	StartRoomID string            `json:"start_room_id"`
	Rooms       map[string]*Room  `json:"rooms"`
	Items       map[string]*Item  `json:"items"`
	NPCs        map[string]*NPC   `json:"npcs"`
	Quests      map[string]*Quest `json:"quests"`
	// Hints maps what killed a player (an NPC, room or item ID) to the Moirai's hint about it.
	Hints map[string]LocalizedText `json:"hints,omitempty"`
}
```

- `StartRoomID`:新規プレイヤーが始まる部屋。
- 4つの辞書:**IDから実体を引く**。`Rooms["loc.hall_of_fates"]` で運命の間の `*Room` が得られる。値が**ポインタ**なので、取り出して書き換えれば辞書の中身も変わる(アイテムの `RoomID` を書き換える処理で使う)。
- `Hints map[string]LocalizedText`
  - **死因になったもののID(敵・部屋・アイテム)から、モイライが教えるヒントの文を引く**辞書です。死んだあとにモイライへ `TALK` したときだけ使います(第11章)。
  - `world.json` では `"hints"` に書きます。IDが実在しないと起動時に弾かれます(5-7)。
- 世界データ(`world.json`)の全体が、この1つの構造体になる。

## 5-6 `loadWorld`

```go
func loadWorld(path string) (*World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read world data: %w", err)
	}
```

- `(*World, error)`:戻り値が2つ。**成功すれば世界とnil、失敗すればnilとエラー**。
- `os.ReadFile(path)`:ファイル全体を `[]byte` として読む。
- `fmt.Errorf("...: %w", err)`:新しいエラーを作る。`%w` を使うと**元のエラーを包んで**保持でき、あとで `errors.Is` で原因を調べられる。
- 失敗時は `nil, エラー` を返して関数を終了。

```go
	var world World
	if err := json.Unmarshal(data, &world); err != nil {
		return nil, fmt.Errorf("decode world data: %w", err)
	}
```

- `var world World`:中身が全部ゼロ値の `World` を作る。
- `json.Unmarshal(data, &world)`:JSONを読み取って `world` に詰める。**`&world` は「`world` のアドレス」**で、これを渡すことで関数が `world` を書き換えられる。
- `if err := ...; err != nil`:「変数を作ってすぐ検査」の形。この `err` は `if` の中だけで有効。

```go
	for _, item := range world.Items {
		if item != nil {
			item.HomeRoomID = item.RoomID
		}
	}
```

- 全アイテムについて、`HomeRoomID` に**読み込み直後の `RoomID` を控えておく**(元の置き場所)。
- `item != nil`:JSONに `"item.x": null` と書かれていた場合に備えた安全確認。

```go
	if err := world.validate(); err != nil {
		return nil, err
	}
	return &world, nil
}
```

- `validate()`(次節)でデータの矛盾を検査。問題があればエラーで終了。
- `&world`:`world` のポインタを返す。ローカル変数のアドレスを返しても、Goは必要なら変数をヒープに確保してくれるので**安全**。

## 5-7 `validate`— 矛盾を起動時に見つける

```go
func (w *World) validate() error {
	if w == nil {
		return fmt.Errorf("world is null")
	}
	if w.Rooms[w.StartRoomID] == nil {
		return fmt.Errorf("start room %q does not exist", w.StartRoomID)
	}
```

- レシーバ `(w *World)`:世界に対する検査。戻り値は `error` だけ(問題なければ `nil`)。
- `w == nil`:JSONが `null` だった場合。
- `w.Rooms[w.StartRoomID] == nil`:**開始部屋が存在するか**。辞書に無いキーは `nil` になる性質を利用。
- `%q`:文字列をダブルクォート付きで埋め込むフォーマット指定(`"loc.xxx"` のように表示される)。

### 部屋の検査

```go
	for id, room := range w.Rooms {
		if id == "" {
			return fmt.Errorf("room ID is empty")
		}
		if room == nil {
			return fmt.Errorf("room %q is null", id)
		}
		if room.ID != id {
			return fmt.Errorf("room %q has ID %q", id, room.ID)
		}
```

- 辞書のキー `id` と、部屋の中の `room.ID` が**一致するか**まで確認する。

```go
		for dir, dest := range room.Exits {
			if w.Rooms[dest] == nil {
				return fmt.Errorf("room %q exit %q points to unknown room %q", id, dir, dest)
			}
		}
```

- **出口の行き先の部屋が実在するか**。タイプミスで存在しない部屋につないでいたら、ここで見つかる。

```go
		for dir, secret := range room.SecretExits {
			if w.Rooms[secret.Room] == nil {
				return fmt.Errorf("room %q secret exit %q points to unknown room %q", id, dir, secret.Room)
			}
			if _, clash := room.Exits[dir]; clash {
				return fmt.Errorf("room %q secret exit %q duplicates a normal exit", id, dir)
			}
		}
```

- **隠し出口の検査**です。普通の出口と同じく、行き先が実在するかを確かめます。
- `if _, clash := room.Exits[dir]; clash {`
  - 辞書を引いて、**値は要らない**(`_` で捨てる)けれど「キーが**あったか**」だけを `clash` に受け取る書き方です。
  - 同じ方角に、普通の出口と隠し出口の**両方があるとエラー**にします。どちらが優先か分からなくなるのを、起動の時点で防ぎます。

```go
		if h := room.Hazard; h != nil {
			switch h.Type {
			case "lethal":
			case "item_gate":
				if w.Items[h.RequiredItemID] == nil {
					return fmt.Errorf("room %q hazard points to unknown item %q", id, h.RequiredItemID)
				}
			case "crew_gate":
				if h.CrewLoss < 1 {
					return fmt.Errorf("room %q crew_gate hazard has invalid crew_loss %d", id, h.CrewLoss)
				}
			case "crew_cost":
				if h.CrewLoss < 1 {
					return fmt.Errorf("room %q crew_cost hazard has invalid crew_loss %d", id, h.CrewLoss)
				}
			default:
				return fmt.Errorf("room %q has unknown hazard type %q", id, h.Type)
			}
		}
	}
```

- `if h := room.Hazard; h != nil`:危険がある部屋だけ、`h` に取り出して調べる。
- `switch h.Type { case ...: }`:`switch` は値で分岐する。Goの `case` は**自動で抜ける**(他言語のような `break` は不要)。
- `case "lethal":` の中身が空:入ると死ぬだけなので、追加の設定項目は無い=検査も不要。
- `item_gate`:必要アイテムが実在するか。
- `crew_gate` / `crew_cost`:仲間が減る数が1以上か。`%d` は整数の埋め込み。
- `default:`:どれにも当てはまらない=**知らない種類**のハザードはエラー。

### アイテムの検査

```go
	for id, item := range w.Items {
		if id == "" { ... }
		if item == nil { ... }
		if item.RewardOnly {
			if item.RoomID != "" || item.Obtainable {
				return fmt.Errorf("reward-only item %q must have no room and must not be obtainable", id)
			}
		} else if w.Rooms[item.RoomID] == nil {
			return fmt.Errorf("item %q points to unknown room %q", id, item.RoomID)
		}
	}
```

- (`{ ... }` は、直前と同じ形の検査を省略して書いた印です。)
- **報酬専用アイテム**:部屋が空で、拾えない設定でなければならない(エンディングの報酬としてだけ入手するため)。
- **それ以外**:置き場所の部屋が実在すること。
- `if A { } else if B { }`:Goの `else if` の書き方。

### NPCの検査

```go
	for id, npc := range w.NPCs {
		...
		if w.Rooms[npc.RoomID] == nil {
			return fmt.Errorf("NPC %q points to unknown room %q", id, npc.RoomID)
		}
		if npc.MythRequirementItem != "" && w.Items[npc.MythRequirementItem] == nil {
			return fmt.Errorf("NPC %q myth requirement points to unknown item %q", id, npc.MythRequirementItem)
		}
		if npc.MythRequirementQuest != "" && w.Quests[npc.MythRequirementQuest] == nil {
			return fmt.Errorf("NPC %q myth requirement points to unknown quest %q", id, npc.MythRequirementQuest)
		}
		if npc.Unwinnable && npc.CrewLossOnAttack < 1 {
			return fmt.Errorf("NPC %q is unwinnable but has invalid crew_loss_on_attack %d", id, npc.CrewLossOnAttack)
		}
	}
```

- NPCのいる部屋、神話条件のアイテム/クエストが**実在するか**。
- 「勝てない敵」なのに仲間の減少数が0以下だと意味がないので、設定ミスとして弾く。

### クエストの検査

```go
	for id, quest := range w.Quests {
		...
		if w.NPCs[quest.GiverNPCID] == nil { ... }
		if quest.Objective.Count < 1 { ... }
		switch quest.Objective.Type {
		case "collect_item":
			if w.Items[quest.Objective.TargetID] == nil { ... }
		case "defeat_npc":
			if w.NPCs[quest.Objective.TargetID] == nil { ... }
		default:
			return fmt.Errorf("quest %q has unknown objective type %q", id, quest.Objective.Type)
		}
	}
	if err := w.validateItemEffects(); err != nil {
		return err
	}
	return w.validateEndings()
}
```

- 依頼者NPCが実在するか。目標の個数が1以上か。
- 目的の種類によって、**対象がアイテムかNPCか**を使い分けて実在確認する。
- `if err := w.validateItemEffects(); err != nil { return err }`
  - **アイテムの効果の検査**(9-4)です。知らない種類の効果や、値が0の効果を弾きます。エラーがあればここで返します。
  - `if 準備文; 条件 {` の形で、`err` を作ると同時に判定しています。
- 最後の `return w.validateEndings()`:エンディング定義の検査(第9章)の結果をそのまま返す。ここまで問題が無ければ `nil`。

### ヒントの検査

```go
	for subject := range w.Hints {
		if w.Rooms[subject] == nil && w.Items[subject] == nil && w.NPCs[subject] == nil {
			return fmt.Errorf("hint is about unknown room, item or NPC %q", subject)
		}
	}
```

- `for subject := range w.Hints`:辞書をrangeするとき、**キーだけ**を受け取れます(値は要らないので省略)。
- ヒントの対象が、部屋・アイテム・NPCの**どれにも当てはまらなければ**エラーです。この検査は部屋のループのあとにあります。

> **この関数のねらい**:遊んでいる最中に「存在しない部屋に移動してクラッシュ」といった事故が起きないよう、**起動の時点で全部の参照を確かめる**。

## 5-8 `resolveNPCInRoom`— 名前やIDからNPCを特定する

```go
func (w *World) resolveNPCInRoom(roomID, query, locale string) string {
	if npc := w.NPCs[query]; npc != nil && npc.RoomID == roomID {
		return query
	}
```

- `(roomID, query, locale string)`:同じ型の引数は**型を1回だけ書いて**まとめられる。
- `query` は、プレイヤーが打った文字列(`npc.polyphemus` や `Polyphemus`)。
- まず **IDそのもの**として解釈し、その部屋にそのNPCがいればIDを返す。

```go
	npcID := ""
	for id, npc := range w.NPCs {
		if npc != nil && npc.RoomID == roomID && strings.EqualFold(npc.Name.Get(locale), query) && (npcID == "" || id < npcID) {
			npcID = id
		}
	}
	return npcID
}
```

- 次に、**名前**で探す。
- `strings.EqualFold(a, b)`:大文字・小文字を区別せず比較する。
- `(npcID == "" || id < npcID)`:同じ名前のNPCが複数見つかったら、**ID(文字列順)が一番小さいもの**を採用する。辞書の `range` の順序はランダムなので、こう決めないと結果が毎回変わってしまう。
- 見つからなければ `""`(呼び出し側が「NPCがいない」エラーにする)。

## 5-9 `questByGiver`

```go
func (w *World) questByGiver(npcID string) (string, *Quest) {
	questID := ""
	for id, quest := range w.Quests {
		if quest != nil && quest.GiverNPCID == npcID && (questID == "" || id < questID) {
			questID = id
		}
	}
	if questID == "" {
		return "", nil
	}
	return questID, w.Quests[questID]
}
```

- 戻り値が `(string, *Quest)` の2つ:クエストのIDと、クエスト本体。
- 依頼者がそのNPCのクエストを探す(複数あればID最小)。
- 見つからなければ `"", nil`。呼ぶ側は `if quest == nil` で判定する。

## 5-10 アイテムの見え方(`Item` のメソッド)

```go
func (item *Item) availableTo(player *Player, id string) bool {
	return item.Obtainable && item.visibleTo(player, id, player.RoomID)
}

func (item *Item) visibleTo(player *Player, id, roomID string) bool {
	if item.RoomID != roomID {
		return false
	}
	return !(item.Renewable && player.hasItem(id))
}
```

- `availableTo`:**取れるか** = 拾える設定で、かつ今いる部屋で見える。
- `visibleTo`:
  - アイテムが**その部屋にある**か(`RoomID` が違えば見えない)。
  - 「再生アイテム(`Renewable`)で、しかもすでに持っている」なら**見えない**。(再生アイテムは取っても部屋に残るので、持っている人にまで見せると二重に取れてしまうため。)
  - `!( ... )`:丸括弧全体を反転。

---

# 第6章 `cmd/server/client_conn.go`(168行)

**役割**:接続1本ごとの**送信用のキュー**です。「サーバー全体の処理を止めずに、遅い相手にも安全にメッセージを送る」ための仕組みです。この章は並行処理の考え方が入るので、少しゆっくり読んでください。

## 6-0 なぜキューが必要か

もしハンドラが直接 `conn.Write(...)` すると、相手の回線が遅い場合に**書き込みが終わるまでその場で止まります**。ハンドラはサーバーのロック(`s.mu`)を持ったまま書くことが多いので、**1人の遅い回線のせいで全員の操作が止まります**。

そこで「書きたい内容をいったん**キュー(チャネル)**に積み、実際の書き込みは専用の別goroutineに任せる」形にします。

```
ハンドラ ──積む──▶ [ キュー(最大64) ] ──▶ writeLoop(専用goroutine) ──▶ ネットワーク
```

## 6-1 import と定数

```go
package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var errSendQueueFull = errors.New("client send queue full")

const defaultWriteTimeout = 10 * time.Second
```

- `errors`:エラー値を作る(`errors.New`)。
- `io`:`io.ErrShortWrite`(「全部書けなかった」を表す標準のエラー)を使う。
- `sync`:`Mutex`、`Once`。 `sync/atomic`:ロック無しで安全に読み書きする部品。
- `errSendQueueFull`:「キューが満杯」を表す**目印用のエラー**。変数にしておくと、呼ぶ側が `errors.Is(err, errSendQueueFull)` のように判定できる。
- `defaultWriteTimeout`:1回の書き込みの制限時間(10秒)。これを超えたら「相手が詰まっている」とみなす。

## 6-2 送信メッセージの型

```go
type outboundMessage struct {
	data []byte
	done chan writeResult
}

type writeResult struct {
	n   int
	err error
}
```

- `outboundMessage`:キューに積む1通ぶん。
  - `data []byte`:送る中身(バイト列)。
  - `done chan writeResult`:**書き終わったことを知らせるための返信用チャネル**。`chan writeResult` は「`writeResult` を運ぶ管」。`nil` なら「完了通知は要らない」。
- `writeResult`:書き込みの結果。`n` は書けたバイト数、`err` はエラー。

## 6-3 `serverClient`

```go
type serverClient struct {
	net.Conn
	out          chan outboundMessage
	done         chan struct{}
	writeTimeout time.Duration
	closeOnce    sync.Once
	locale       string

	remote string
	ctx    atomic.Pointer[clientContext]
}
```

- `net.Conn`:型名だけを書く = **埋め込み**。`serverClient` は `net.Conn` の機能(`Read`、`Close`、`SetWriteDeadline`…)をすべて**そのまま使える**。必要なものだけ自分で書き直せる(後の `Write` と `Close`)。
- `out`:送信キュー。
- `done chan struct{}`:**終了の合図専用のチャネル**。`struct{}` は「中身が無い型」。値を運ぶ目的ではなく、`close(done)` したときに**待っている全員に一斉に終了を知らせる**ために使う。
- `writeTimeout`:書き込みの制限時間。
- `closeOnce sync.Once`:「**1回だけ実行する**」ための部品。`Close` が複数のgoroutineから呼ばれても、実際の終了処理を1回だけにできる。
- `locale`:この接続の言語(`LANG` で設定)。
- `remote`:相手のアドレス(ログ用)。
- `ctx atomic.Pointer[clientContext]`:**いま処理中の「プレイヤー名とコマンド名」**。`atomic.Pointer` はロックなしで安全に読み書きできるポインタ。`[clientContext]` は「`clientContext` を指す」という指定(ジェネリクス)。

## 6-4 `newServerClient`

```go
func newServerClient(conn net.Conn) *serverClient {
	return newServerClientWithWriteTimeout(conn, defaultWriteTimeout)
}

func newServerClientWithWriteTimeout(conn net.Conn, timeout time.Duration) *serverClient {
	client := &serverClient{
		Conn:         conn,
		remote:       remoteOf(conn),
		out:          make(chan outboundMessage, 64),
		done:         make(chan struct{}),
		writeTimeout: timeout,
	}
	go client.writeLoop()
	return client
}
```

- 1つ目は**通常用**(10秒)。2つ目は**秒数を指定できる版**(テストで短い時間を使うために分けてある)。
- `&serverClient{ ... }`:構造体を作ってそのポインタを得る。 `Conn: conn` のように、埋め込んだフィールドは**型名がフィールド名**になる。
- `make(chan outboundMessage, 64)`:**容量64のチャネル**。64通まではためておける。満杯だと送信側が待たされる(または失敗する)。
- `make(chan struct{})`:容量なし(0)のチャネル。 `close` を合図にするだけなので問題ない。
- **`go client.writeLoop()`**:送信専用のgoroutineを起動。これが生きている間、キューを監視し続ける。

## 6-5 `writeLoop`— 実際に書く係

```go
func (client *serverClient) writeLoop() {
	for {
		select {
		case message := <-client.out:
			...
		case <-client.done:
			return
		}
	}
}
```

- 無限ループの中で `select`。`select` は**複数のチャネル操作のうち、動かせるものを1つ実行**する(どれも動かせないときは待つ)。
- `case message := <-client.out:`:キューから1通取り出せたら実行。
- `case <-client.done:`:`done` が閉じられたら終了(`return`)。

```go
			err := client.Conn.SetWriteDeadline(time.Now().Add(client.writeTimeout))
			n := 0
			if err == nil {
				n, err = client.Conn.Write(message.data)
			}
			if err == nil && n != len(message.data) {
				err = io.ErrShortWrite
			}
```

- `SetWriteDeadline(今 + 10秒)`:**この書き込みは10秒以内に終わらせる**という締め切り。守れなければ `Write` がエラーを返す。
- `client.Conn.Write(...)`:ネットワークへ実際に書く。**埋め込みの元の `Write`** を呼ぶため `client.Conn.Write` と書く(`client.Write` だと自分で上書きした方=キューに積む方が呼ばれて、ループになってしまう)。
- `n, err = ...`:ここの `=` は新しい変数を作らず、**すでにある `n` と `err` に代入**する。
- 書けたバイト数が少なければ `io.ErrShortWrite`。

```go
			if err == nil {
				logOutbound(client, message.data)
			}
			if message.done != nil {
				message.done <- writeResult{n: n, err: err}
			}
			if err != nil {
				client.Close()
				return
			}
```

- 成功したら送った内容をログに記録(`logOutbound`は第13章)。
- 完了通知用のチャネルがあれば、結果を送る(待っているハンドラが目を覚ます)。このチャネルは容量1で作られているので、**受け取り手がいなくても詰まらない**。
- エラーなら、接続を閉じて `writeLoop` 自体も終了。

## 6-6 `Write`— `net.Conn` の `Write` を上書き

```go
func (client *serverClient) Write(data []byte) (int, error) {
	message := outboundMessage{
		data: append([]byte(nil), data...),
		done: make(chan writeResult, 1),
	}
```

- ハンドラの `fmt.Fprintln(conn, "ERR 400 ...")` は、内部でこの `Write` を呼ぶ。つまり **エラー応答も自動的にキュー経由**になる。
- `append([]byte(nil), data...)`:`data` を**コピー**する定番の書き方。呼び出し側が後でバッファを再利用しても、キューの中身が変わらない。
- `make(chan writeResult, 1)`:返信用チャネル(容量1)。

```go
	select {
	case client.out <- message:
	case <-client.done:
		return 0, net.ErrClosed
	}
	select {
	case result := <-message.done:
		return result.n, result.err
	case <-client.done:
		return 0, net.ErrClosed
	}
}
```

- 1つ目の `select`:キューに積む。満杯なら**空くまで待つ**。待っている間に接続が閉じたら `net.ErrClosed` で失敗。
- 2つ目の `select`:`writeLoop` が書き終わるのを待つ。終わったら結果を返す。
- この `Write` は**書き終わるまで戻らない**(同期的)。ロックを持っていない場面(エラー応答など)で使う。

## 6-7 `enqueueResponse` と `waitResponse`

```go
func (client *serverClient) enqueueResponse(line string) (<-chan writeResult, error) {
	message := outboundMessage{
		data: []byte(line + "\n"),
		done: make(chan writeResult, 1),
	}
	select {
	case client.out <- message:
		return message.done, nil
	case <-client.done:
		return nil, net.ErrClosed
	default:
		return nil, errSendQueueFull
	}
}
```

- 戻り値の `<-chan writeResult` は「**受信専用**のチャネル」。呼び出し側は読むだけで、書けない。
- `select` に **`default:`** がある:どのcaseも今すぐ動かせないなら、**待たずに `default` を実行**する。ここでは「キューが満杯ならすぐ `errSendQueueFull`」。
- **ここが重要**:送信キューへ積む段階では、ネットワークへの書き込み完了を待たない。キューが満杯なら接続を閉じてエラーを返す。実際の書き込みを待つ`waitResponse`は、サーバーのロックを外してから呼ぶ。
- 呼ぶ側は `response, err := client.enqueueResponse("OK ...")` と書き、後で `waitResponse(response)` で完了を待つ。

```go
func (client *serverClient) waitResponse(done <-chan writeResult) error {
	select {
	case result := <-done:
		return result.err
	case <-client.done:
		return net.ErrClosed
	}
}
```

- 積んだメッセージが**実際に書かれた**(か、接続が閉じた)のを待つ。ハンドラはこれを**ロックを外したあと**に呼ぶ。

## 6-8 `enqueueEvent`— 通知(EVT)の送信

```go
func (client *serverClient) enqueueEvent(line string) {
	select {
	case <-client.done:
		return
	default:
	}
	select {
	case client.out <- outboundMessage{data: []byte(line + "\n")}:
	case <-client.done:
	default:
		client.Close()
	}
}
```

- 1つ目の `select`:すでに閉じた接続なら何もしない(`default:` が空 = 「閉じてなければ次へ進む」)。
- 2つ目の `select`:キューに積む。**満杯(`default`)なら `client.Close()`**:遅くて追いつけないクライアントは**切り捨てる**。
  - 応答(`enqueueResponse`)は失敗を返すだけで切らないが、通知は「いくらでも溜まる可能性がある」ので、溜まりすぎた相手は切る。
- `done` を `nil` のまま渡している(完了通知なし)。結果を待たない一方通行。

## 6-9 `Close`

```go
func (client *serverClient) Close() error {
	var err error
	client.closeOnce.Do(func() {
		close(client.done)
		err = client.Conn.Close()
	})
	return err
}
```

- `closeOnce.Do(func() { ... })`:渡した関数を**最初の1回だけ**実行(2回目以降は何もしない)。 `writeLoop`、ハンドラ、`enqueueEvent` など、どこからでも `Close` を呼べて安全。
- `func() { ... }` は**無名関数**(名前のない関数)。その場で作って渡している。
- `close(client.done)`:`done` を閉じる。これで `select` で `<-client.done` を待っていた全員(`writeLoop`、`Write`、`waitResponse`…)が**一斉に目を覚ます**。
- `client.Conn.Close()`:本物の接続を閉じる。

## 6-10 コンテキスト(ログ用)

```go
type clientContext struct {
	player  string
	command string
}

func (client *serverClient) setContext(player, command string) {
	client.ctx.Store(&clientContext{player: player, command: command})
}

func (client *serverClient) context() (player, command string) {
	if c := client.ctx.Load(); c != nil {
		return c.player, c.command
	}
	return "", ""
}
```

- 「いま誰のどのコマンドを処理しているか」を保存・取得する。
- `Store`/`Load` はロック不要で安全。`writeLoop` は別goroutineなので、`handleClient` が書いた値を**ロックなしで読めること**に意味がある(ログに「どのコマンドへの応答か」を付けるため)。
- `(player, command string)` は**名前付き戻り値**(説明を兼ねている)。

```go
func remoteOf(conn net.Conn) string {
	if addr := conn.RemoteAddr(); addr != nil {
		return addr.String()
	}
	return ""
}
```

- 相手のアドレス(`"192.168.1.5:50123"`)を文字列で返す。取れないときは空文字。

# 第7章 `cmd/server/server.go`(849行)

**役割**:サーバー本体です。「サーバー全体の状態(`Server`)」「接続・退室の処理」「基本コマンド(CONNECT / LOOK / MOVE / WHO / QUIT / TAKE / DROP / INVENTORY / TALK / STATUS)」「接続ごとのメインループ(`handleClient`)」が入っています。

840行と長いので、7-1〜7-9 を前半、7-10 以降を後半として読みます。

## 7-1 import と定数

```go
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)
```

- `bufio`:行単位で読む `Scanner` を使う。
- `encoding/json`:応答のJSONを作る(`json.Marshal`)。
- `errors`:`errors.New` / `errors.Is`。
- `sort`:並べ替え(`sort.Strings`)。
- `sync`:`Mutex`。
- `time`:時刻。
- `unicode`:文字の種類判定(制御文字か)。 `unicode/utf8`:文字列が正しいUTF-8か。

```go
var errNameInUse = errors.New("player name in use")

const defaultStartRoomID = "loc.hall_of_fates"
const maxProtocolLineBytes = bufio.MaxScanTokenSize - 1
```

- `errNameInUse`:「名前が使用中」を表す目印のエラー。
- `defaultStartRoomID`:世界データが読めなかったとき用の、開始部屋の初期値。
- `maxProtocolLineBytes`:**1行の最大バイト数**。 `bufio.MaxScanTokenSize` は64KB(65536)で、`Scanner` が1行として受け付ける上限。そのぴったり1つ下にしてあるので、**この長さ以内なら必ずクライアントの `Scanner` で読める**。

```go
const (
	maxPlayerHP      = 100
	respawnHP        = 20
	combatMinDamage  = 8
	combatMaxDamage  = 14
	counterMinDamage = 7
	counterMaxDamage = 14

	odysseyStartRoomID = "loc.ody_troy_shore"
	startingCrew       = 12
)
```

- ゲームバランスの数値。
  - `maxPlayerHP`:**最大HPの基本値**100(全員が同じ値から始まる)。クエストで増えた分とアイテムの効果は、ここへ足して計算します(`maxHPLocked`、9-4)。名前は `max...` ですが、「上限そのもの」ではなく「出発点」です。
  - `respawnHP`:復活時のHP20。
  - 与ダメージ8〜14、反撃7〜14。
  - オデュッセイア編の開始部屋と、最初の仲間の人数12人。
- 数字をコードの中に直書きせず、名前を付けて1か所にまとめている。

## 7-2 `Server` 構造体

```go
type Server struct {
	mu            sync.Mutex
	ioMu          sync.Mutex
	players       map[string]*Player
	clients       map[string]*serverClient
	groups        map[string]*Group
	groupByPlayer map[string]string
	unsavedTakes  map[string]map[string]string
	nextGroupID   uint64
	abuse         *abuseMonitor
	saveDir       string
	world         *World
}
```

- `mu sync.Mutex`:**サーバーの状態を守る鍵**。 `Lock()` した人だけが中に入れる。ほぼすべてのハンドラが、状態を触る間これを持つ。
- `ioMu sync.Mutex`:**ファイルの読み書き専用の鍵**。遅いファイル操作の間にゲーム処理を止めないよう、`mu` とは別にしてある。
- `players`:接続中のプレイヤー(名前→状態)。
- `clients`:接続中の通信口(名前→`serverClient`)。通知(EVT)を送るときに引く。
- `groups`:グループID→グループ。
- `groupByPlayer`:プレイヤー名→所属グループID。
- `unsavedTakes`:**「取ったがまだ保存していないアイテム」の記録**。外側のキーがプレイヤー名、内側が「アイテムID→取った部屋」。異常切断のときにアイテムを元の部屋に戻すために使う。
  - `map[string]map[string]string`:辞書の中に辞書が入った形。
- `nextGroupID uint64`:グループ番号の採番用。 `uint64` は0以上の大きな整数。
- `abuse *abuseMonitor`:接続の連打を監視する(第13章)。
- `saveDir`:セーブ先のフォルダ名。
- `world`:読み込んだ世界データ。

## 7-3 `NewServer` と `newServer`

```go
func NewServer() *Server {
	s := newServer("saves")
	world, err := loadWorld("data/world.json")
	if err != nil {
		fatal("load_world_failed", err)
	}
	s.world = world
	if err := s.restoreItemLocations(); err != nil {
		fatal("restore_item_locations_failed", err)
	}
	if err := s.restoreItemOwnership(); err != nil {
		fatal("restore_item_ownership_failed", err)
	}
	return s
}
```

- `newServer("saves")`で空のサーバーを作る(下で説明)。
- `loadWorld("data/world.json")`で世界を読み、サーバーに持たせる。失敗は `fatal`(ログを出して終了)。
- `restoreItemLocations()`:**前回までのアイテムの置き場所**をセーブから復元(第12章)。
- `restoreItemOwnership()`:**プレイヤーが持っているアイテム**を、部屋から取り除く(第12章)。
- どれかが失敗したら起動しない。壊れたデータで動き出すより、すぐ止まる方が安全。
- **相対パス**(`"saves"`、`"data/world.json"`)なので、**リポジトリのルートで起動しないと動かない**。

```go
func newServer(saveDir string) *Server {
	return &Server{
		players:       make(map[string]*Player),
		clients:       make(map[string]*serverClient),
		groups:        make(map[string]*Group),
		groupByPlayer: make(map[string]string),
		unsavedTakes:  make(map[string]map[string]string),
		abuse:         newAbuseMonitor(),
		saveDir:       saveDir,
	}
}
```

- 辞書は `make` で**空の辞書を作っておく**(`nil` の辞書には書き込めないため)。
- `NewServer`(大文字)と `newServer`(小文字)は別の関数。小文字版は**セーブ先を引数で受け取る**ので、テストで一時フォルダを渡して使える。

## 7-4 `connectPlayer`— ログインの本体

3段階に分けて、ロックを長く持たないようにしています。

```go
func (s *Server) connectPlayer(name string) error {
	s.mu.Lock()
	if _, exists := s.players[name]; exists {
		s.mu.Unlock()
		return errNameInUse
	}
	s.mu.Unlock()
```

- **第1段階:名前の重複確認**。
- `s.mu.Lock()`:鍵を取る。他のgoroutineが持っていたら、**手放されるまでここで待つ**。
- `if _, exists := s.players[name]; exists`:辞書にその名前があるかだけを知りたいので、値は `_` で捨てる。
- 使用中なら**鍵を返して**から(`Unlock`)、エラーを返す。 **`return` の前に必ず `Unlock` する**のを忘れると、鍵が戻らずサーバー全体が止まる。

```go
	s.ioMu.Lock()
	players, err := s.loadPlayers()
	var player *Player
	if err == nil {
		player = players[name]
		switch {
		case player == nil:
			startRoom := defaultStartRoomID
			if s.world != nil {
				startRoom = s.world.StartRoomID
			}
			player = &Player{Name: name, HP: 100, RoomID: startRoom}
			players[name] = player
			err = s.writePlayers(players)
		case s.world != nil && s.world.Rooms[player.RoomID] == nil:
			err = fmt.Errorf("saved player %q has unknown room %q", name, player.RoomID)
		}
	}
	s.ioMu.Unlock()
	if err != nil {
		return err
	}
```

- **第2段階:セーブファイルを調べる**(ファイル用の鍵 `ioMu` を取る)。
- `s.loadPlayers()`:セーブファイルから全プレイヤーを読む。
- `var player *Player`:ここでは `nil`(まだ誰も指していない)。
- `switch { case 条件: ... }`:`switch` の後ろに何も書かない形。 **上から順に条件を見て、最初に真になったcaseだけ実行**。
  - `player == nil`:セーブに無い = **初めての人**。開始部屋に、HP100の新しい `Player` を作って保存する。 `&Player{...}` は「作ってポインタを得る」。
  - 2つ目:セーブはあるが、**その部屋が今の世界に存在しない**(`world.json` が変わった)なら、エラーにする。
- `ioMu` を解除し、エラーがあればここで返す。

```go
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.players[name]; exists {
		return errNameInUse
	}
	player.lastRegen = time.Now()
	s.players[name] = player
	return nil
}
```

- **第3段階:登録**。
- `defer s.mu.Unlock()`:関数を抜けるとき自動で鍵を返す。以降どこで `return` してもよい。
- **もう一度、重複を確認する**。第1段階とここまでの間は鍵を外していたので、別の接続が同じ名前で先に入った可能性がある。 (「**二重チェック**」と呼ばれる定番の書き方)
- `lastRegen`を接続時刻に設定し、再接続した人も最初の更新から接続後の待機時間をHP回復へ算入します。接続していない間の時間は回復に加算しません。
- 問題なければ `s.players` に登録して成功(`nil`)。

## 7-5 `saveAndRemovePlayer`— 保存して退室

```go
func (s *Server) saveAndRemovePlayer(name string) error {
	s.mu.Lock()
	player := s.players[name]
	if player == nil {
		s.mu.Unlock()
		return fmt.Errorf("player %q is not connected", name)
	}
	if player.exiting {
		s.mu.Unlock()
		return fmt.Errorf("player %q is already exiting", name)
	}
	player.exiting = true
	snapshot := *player
	snapshot.Inventory = append([]string(nil), player.Inventory...)
	s.mu.Unlock()
```

- 接続中でなければエラー。すでに退室処理中(`exiting`)でもエラー(二重に保存しないため)。
- `player.exiting = true`:退室中の印を付ける。
- `snapshot := *player`:`*player` は**ポインタの先の中身**。それを代入するので、 `snapshot` は**コピー**になる。
- ただし構造体のコピーは**スライスや辞書の中身までは複製しない**(同じものを共有する)。そこで持ち物のスライスだけ `append([]string(nil), ...)` で別に複製している。
- 鍵を外す。ここから先は、ファイルへの保存(遅い処理)を**ゲーム状態の鍵を持たずに**行う。

```go
	s.ioMu.Lock()
	err := s.savePlayer(&snapshot)
	s.ioMu.Unlock()
	s.mu.Lock()
	if err != nil {
		player.exiting = false
		s.mu.Unlock()
		return err
	}

	s.removePlayerLocked(name, false)
	s.mu.Unlock()
	return nil
}
```

- コピー(`snapshot`)を保存する。
- 保存に失敗したら、退室中の印を戻してエラーを返す(まだ接続中のまま)。
- 成功したら `removePlayerLocked` で状態から消す(次節)。

## 7-6 `removePlayerLocked`

```go
func (s *Server) removePlayerLocked(name string, restoreUnsavedTakes bool) {
	player := s.players[name]
	if player == nil {
		return
	}
```

- `Locked`:呼ぶ前に `s.mu` を持っていること。
- 第2引数 `restoreUnsavedTakes`:**未保存で取ったアイテムを部屋に戻すか**。

```go
	if restoreUnsavedTakes && s.world != nil {
		inventory := make(map[string]struct{}, len(player.Inventory))
		for _, itemID := range player.Inventory {
			inventory[itemID] = struct{}{}
		}
		for itemID, roomID := range s.unsavedTakes[name] {
			if _, held := inventory[itemID]; !held || s.world.Rooms[roomID] == nil {
				continue
			}
			if item := s.world.Items[itemID]; item != nil && item.RoomID == "" {
				item.RoomID = roomID
			}
		}
	}
```

- 保存に失敗して強制的に消すとき、その人が取ったアイテムを**世界から失わせない**ための処理。
- `inventory`:持ち物を**集合**(set)にする。 `map[string]struct{}` は「キーだけを使う辞書」で、 `struct{}{}` は「中身のない値」。 `_, held := inventory[itemID]` で「持っているか」を素早く調べられる。
- 未保存の取得記録 `s.unsavedTakes[name]` を1つずつ見て、**まだ持っている**ものを元の部屋に戻す(`item.RoomID = roomID`)。
- `continue`:条件に合わないものは飛ばす。
- 戻すのは `item.RoomID == ""`(誰にも持たれていない状態)のときだけ。

```go
	roomID := player.RoomID
	delete(s.players, name)
	delete(s.clients, name)
	delete(s.unsavedTakes, name)
	s.removeGroupMemberLocked(name)
	s.clearGroupInvitesLocked(name)
	for otherName, other := range s.players {
		if other.RoomID == roomID {
			if client := s.clients[otherName]; client != nil {
				client.enqueueEvent("EVT ROOM PRESENCE LEAVE " + name)
			}
		}
	}
	s.broadcastPlayerCountLocked()
}
```

- `delete(辞書, キー)`:辞書から消す組み込み関数。
- グループからの脱退と、招待の取り消し。
- **同じ部屋の他の人**に `EVT ROOM PRESENCE LEAVE 名前` を送る。
- 全員に人数の更新を送る。

## 7-7 `broadcastPlayerCountLocked` と `playerForUpdateLocked`

```go
func (s *Server) broadcastPlayerCountLocked() {
	event := fmt.Sprintf("EVT STATS players=%d", len(s.players))
	for _, client := range s.clients {
		client.enqueueEvent(event)
	}
}
```

- `fmt.Sprintf`:書式付きの文字列を**作って返す**(画面に出さない)。 `%d` に `len(s.players)` が入る。
- 全クライアントに現在の人数を通知。

```go
func (s *Server) playerForUpdateLocked(name string) *Player {
	player := s.players[name]
	if player == nil || player.exiting {
		return nil
	}
	player.regenLocked(time.Now(), s.effectTotalLocked(player, blessingRegenBonus), s.maxHPLocked(player))
	return player
}
```

- 状態を変えるコマンドの先頭で使う。**退室中の人は `nil` 扱い**にして、 `regenLocked` で**HP自動回復を反映**してから返す。
- `regenLocked` に渡す2つの値は、**ヘラの祝福とアイテムの効果を合算した「1回あたりの追加回復」**(`effectTotalLocked`、9-4)と、**このプレイヤーの今の最大HP**(`maxHPLocked`)です。以前は祝福だけを渡していましたが、アイテム効果も加わりました。
- これを通すだけで、全コマンドでHP回復が効く。

## 7-8 引数チェックの補助

```go
func requireArgs(conn net.Conn, parts []string, min int) bool {
	if len(parts) < min {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	return true
}

func requireExactArgs(conn net.Conn, parts []string, n int) bool {
	if len(parts) != n {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	return true
}
```

- `len(parts)`:単語の数。コマンド名も1つに数える(`MOVE east` は2)。
- `requireArgs`:**最低でも min 個**。 `requireExactArgs`:**ちょうど n 個**。
- 足りなければ **この場でエラー応答を送って `false`** を返す。ハンドラは `if !requireArgs(...) { return false }` の形で使う。

## 7-9 コマンドの表

```go
type commandHandler func(s *Server, conn net.Conn, name *string, parts []string) (stop bool)

var commandHandlers = map[string]commandHandler{
	"LANG":      handleLang,
	"CONNECT":   handleConnect,
	...
	"QUESTS":    handleQuests,
}
```

- `type commandHandler func(...) (stop bool)`:**関数の型に名前を付ける**。「この4つの引数を取り、`bool` を返す関数」を `commandHandler` と呼ぶ。 `(stop bool)` は戻り値の意味を示す名前(`true` なら接続終了)。
- `commandHandlers`:**コマンド名→担当関数**の辞書。Goでは関数も値なので、辞書に入れられる。
- 新しいコマンドを足すときは、関数を書いてここに1行足すだけ。

## 7-10 すべてのハンドラに共通する骨格

以降のハンドラは、ほぼ次の形をしています。**最初にここで一度だけ詳しく説明し、後ろでは「いつもの骨格」と呼びます。**

```go
func handleXxx(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) { return false }     // ① 引数の数
	if *name == "" {                                          // ② CONNECT済みか
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	s.mu.Lock()                                               // ③ 鍵を取る
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil {                      // ④ 状態の確認
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	/* ⑤ 実際の処理 */
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK ...")         // ⑥ 応答を送信キューへ
	if err == nil { /* ⑦ 成功したら、他の人への通知など */ }
	s.mu.Unlock()                                             // ⑧ 鍵を返す
	if err != nil { return true }                             // ⑨ 積めなかった(満杯/切断)=接続終了
	return client.waitResponse(response) != nil               // ⑩ 書き終わるのを待つ(鍵の外で)
}
```

- ①②:入力の検査。 `*name` は接続のプレイヤー名(空 = まだCONNECTしていない)。
- ③⑧:鍵の取得と返却。**エラーで早く抜けるときも、必ず `Unlock` してから `return`** している(各所の `s.mu.Unlock(); fmt.Fprintln(...); return false` の3行セットがそれ)。
- ④:内部状態がおかしいときは `ERR 500 STATE_ERROR`。
- ⑥:応答を**キューに積む**(待たない)。 `conn.(*serverClient)` で型アサーションし、送信キュー付きの接続を取り出す。
- ⑦:**`err == nil` のときだけ**、他の人への通知などをする(応答が積めなかったのに通知だけ出すとおかしくなるため)。
- ⑨:積めない = 接続がおかしい。 `true`(終了)を返す。
- ⑩:**鍵を返したあと**に書き終わりを待つ。鍵を持ったまま待つと、他の人の操作が止まる。 `!= nil` は「エラーがあれば `true`(終了)」。
- **応答を先に積み、そのあとで通知を積む**:同じキューなので、クライアントには必ず「応答 → 通知」の順で届く。

以降、この骨格に当てはまる部分は説明を省略し、**違うところ**だけ書きます。

## 7-11 `handleConnect`

```go
func handleConnect(s *Server, conn net.Conn, name *string, parts []string) bool {
	client := conn.(*serverClient)
	if len(parts) != 2 || *name != "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
```

- `CONNECT 名前` は単語2つが必須。すでに接続済み(`*name != ""`)ならエラー。

```go
	requestedName := parts[1]
	if len("EVT ROOM PRESENCE ENTER ")+len(requestedName) > maxProtocolLineBytes ||
		!utf8.ValidString(requestedName) ||
		strings.IndexFunc(requestedName, unicode.IsControl) >= 0 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
```

- 名前の検査(3つのどれかに当てはまればエラー)。
  1. `len(...)`:`"EVT ROOM PRESENCE ENTER 名前"` が**1行の上限を超える**ほど長い。
  2. `!utf8.ValidString(...)`:正しいUTF-8の文字列ではない。
  3. `strings.IndexFunc(s, unicode.IsControl) >= 0`:**制御文字**(改行・タブ・ESCなど)を含む。 `IndexFunc` は「条件に合う最初の文字の位置」を返し、無ければ `-1`。
- 名前は他の人への通知文(`EVT ... 名前`)にそのまま入る。**改行入りの名前で偽の通知行を作らせない**ための対策。
- `||` の途中で行が折り返されているのは、見やすくするためだけ。

```go
	if err := s.connectPlayer(requestedName); errors.Is(err, errNameInUse) {
		fmt.Fprintln(conn, "ERR 201 NAME_IN_USE")
		return false
	} else if err != nil {
		logger.Error("connect_player_failed", "player", requestedName, "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
```

- 7-4の `connectPlayer` を呼ぶ。
- `errors.Is(err, errNameInUse)`:エラーの**種類の判定**。「名前使用中」なら `201`。
- それ以外のエラー(ファイル破損など)はログに記録して `500`。
- `if A { } else if B { }` で、 `err` は両方の枝で使える。

```go
	*name = requestedName
	client.setContext(requestedName, "CONNECT")
	s.mu.Lock()
	response, err := client.enqueueResponse("OK connected")
```

- `*name = requestedName`:**ポインタの先に書き込む**。これで呼び出し元 `handleClient` の変数 `name` が書き換わり、以降「接続済み」になる。
- ログ用にコンテキストを更新。
- 鍵を取り、`OK connected` をキューに積む。

```go
	if err == nil {
		s.clients[requestedName] = client
		roomID := s.players[requestedName].RoomID
		for otherName, other := range s.players {
			if otherName != requestedName && other.RoomID == roomID {
				if recipient := s.clients[otherName]; recipient != nil {
					recipient.enqueueEvent("EVT ROOM PRESENCE ENTER " + requestedName)
				}
			}
		}
		s.broadcastPlayerCountLocked()
```

- 積めたら **ここで初めて `s.clients` に登録**。応答の**後に**登録するので、自分宛ての通知が `OK connected` より先に届くことがない。
- 同じ部屋の他の人に `ENTER` を通知。全員に人数を通知。

```go
		if player := s.players[requestedName]; player != nil {

			if !player.IntroSeen {
				player.IntroSeen = true
				s.sendGuideLocked(requestedName, 0)
			}
			s.announceQuestGiversLocked(player)
		}
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}
```

- **初めての人**には、案内役の導入文を送る(`IntroSeen` を `true` にして二度と送らない)。
- その部屋にクエストの依頼者がいれば案内(第9章)。
- あとはいつもの骨格(Unlock → 待つ)。

## 7-12 `handleLook`

骨格に加えて、次のことをします。

```go
	roomID := player.RoomID
	room := newRoomView(s.world.Rooms[roomID], clientLocale(conn))
```

- 今の部屋の情報を、**その接続の言語**で表示用の形にする(第4章)。

```go
	room.Exits = s.world.Rooms[roomID].exitsFor(player)
```

- `newRoomView` が作った出口は「普通の出口だけ」です。ここで**このプレイヤーに見える出口**(隠し出口を開けた人は、それも含む)で上書きします(`exitsFor`、2-6)。
- 同じ部屋を見ても、**最終エンディングを見た人にだけ運命の間の南の出口が見える**のは、この1行のためです。

```go
	players := make([]string, 0)
	for playerName, other := range s.players {
		if other.RoomID == roomID {
			players = append(players, playerName)
		}
	}
	items := make([]string, 0)
	for itemID, item := range s.world.Items {
		if item != nil && item.visibleTo(player, itemID, roomID) {
			items = append(items, itemID)
		}
	}
	npcs := make([]string, 0)
	for npcID, npc := range s.world.NPCs {
		if npc != nil && npc.RoomID == roomID {
			npcs = append(npcs, npcID)
		}
	}
```

- 同じ部屋の **プレイヤー名・見えるアイテムのID・NPCのID** を集める。
- NPCは、**倒した敵も含めて**全員を集めます。「いなくなった」のではなく「倒れている」ので、IDは残したまま、別の欄(下の `Defeated`)で倒したことを伝えます。
- `make([]string, 0)`:**長さ0の空のスライス**を作る。 `var x []string`(nil)でも動くが、JSONにしたとき `null` ではなく `[]` になるように、わざわざ空スライスにしている。

```go
	sort.Strings(players)
	sort.Strings(items)
	sort.Strings(npcs)
```

- 辞書の `range` は順序がバラバラ。 **毎回同じ並びで返す**ために並べ替える。

```go
	data, err := json.Marshal(struct {
		Room    roomView `json:"room"`
		Players []string `json:"players"`
		Items   []string `json:"items"`
		NPCs    []string `json:"npcs"`
		// Defeated lists the NPCs here this player has beaten; it is left out when there are none.
		Defeated []string `json:"defeated,omitempty"`
	}{room, players, items, npcs, s.defeatedNPCsLocked(player, roomID)})
```

- **無名の構造体**(名前を付けない型)をその場で定義して、すぐ値を入れてJSONにする。その場限りの形なので、型に名前を付けるまでもない。
- `}{room, players, items, npcs, s.defeatedNPCsLocked(player, roomID)}`:フィールドの**定義順**に値を渡している。最後の値は、**このプレイヤーが倒したNPCのID一覧**です(`defeat.go`、8-7)。
- `Defeated []string` の `json:"defeated,omitempty"`
  - RFCの `LOOK` には無い**追加の欄**です。GUIが倒した敵を倒れた絵で描くために使います。
  - `omitempty` なので、**倒した敵がいなければJSONに出ません**。この欄を知らないクライアントは無視すればよいので、RFC準拠のクライアントとも互換です。
- `json.Marshal`:構造体 → JSONのバイト列。

```go
	if len("OK ")+len(data) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
```

- 応答が**1行の上限を超える**なら `500`(プロトコル違反の長さを送らない)。
- `string(data)`:バイト列を文字列に変換。

## 7-13 `handleMove`

```go
	room := s.world.Rooms[player.RoomID]
	if room == nil { ...500... }
	destination, ok := room.exitsFor(player)[strings.ToLower(parts[1])]
	if !ok {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 301 NO_EXIT")
		return false
	}
	if s.world.Rooms[destination] == nil { ...500... }
```

- `room.exitsFor(player)[方向]` で行き先を引く。方向は**小文字にしてから**引く(`NORTH` でも通る)。
  - 出口の辞書は、`LOOK` と同じ `exitsFor`(2-6)から取ります。**まだ開いていない隠し出口には進めません**(`ERR 301 NO_EXIT` になる)。
- `destination, ok := ...`:辞書の2値形式。 `ok` が `false` = その方向に出口が無い → `301 NO_EXIT`。
- 行き先の部屋が世界に無ければ(データ異常)`500`。

```go
	if blockerID, blocker := s.blockingEnemyLocked(player, player.RoomID); blocker != nil {
		encounterRoomID := player.RoomID
		locale := clientLocale(conn)
		client := conn.(*serverClient)
		response, err := client.enqueueResponse("OK room=" + destination)
		if err == nil {
			s.respawnPlayerLocked(player, *name, "slip_past", blockerID, blocker.Name.Get(locale))
			s.broadcastFlavorLocked(encounterRoomID, flavor{key: "slip_past", player: *name, npc: blocker})
		}
		s.mu.Unlock()
		if err != nil {
			return true
		}
		return client.waitResponse(response) != nil
	}
```

- **生きている敵が出口を塞いでいる**(`blockingEnemyLocked`、第8章)と、移動しようとした人は**すり抜けようとして殺される**。
- `blockerID, blocker := ...`:戻り値は「塞いでいる敵のID」と「その敵のデータ」。 `blocker != nil` なら敵がいる。IDは、あとで**死因の記録**(`LastDeathSubject`、第3章)として `respawnPlayerLocked` に渡します。
- 応答は `OK room=行き先` だが、**実際には移動せず**、 `respawnPlayerLocked` で死亡・復活する。実況(`broadcastFlavorLocked`)も部屋に送る。
- ここで `return` するので、下の通常移動には進まない。

```go
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK room=" + destination)
	if err == nil {
		oldRoomID := player.RoomID
		player.RoomID = destination
		logger.Info("player_moved", "player", *name, "from", oldRoomID, "to", destination)
		if oldRoomID != destination {
			player.CombatTargetID = ""
			player.guarding = false
			for playerName, current := range s.players {
				recipient := s.clients[playerName]
				if recipient == nil {
					continue
				}
				switch current.RoomID {
				case oldRoomID:
					recipient.enqueueEvent("EVT ROOM PRESENCE LEAVE " + *name)
				case destination:
					recipient.enqueueEvent("EVT ROOM PRESENCE ENTER " + *name)
				}
			}
		}
```

- **通常の移動**。応答を積み、成功したら `player.RoomID` を書き換える。
- 部屋が変わったときは戦闘対象と防御状態を解除します。移動失敗や同じ部屋への移動では解除せず、過去の逃走成功の記録も保持します。
- 全プレイヤーを1周して、 **その人のいる部屋が旧部屋なら `LEAVE`、新部屋なら `ENTER`** を送る。 `switch current.RoomID { case 旧: ... case 新: ... }` は、値を比べて分岐する書き方。
- 自分自身も `players` に含まれるが、自分は今 `destination` にいるので、自分には `ENTER` が届く。

```go
		s.initializeCrewLocked(player, oldRoomID)
		if event := s.applyRoomHazardLocked(player, *name, s.world.Rooms[destination], clientLocale(conn)); event != nil {
			s.broadcastFlavorLocked(destination, *event)
		}
		s.announceQuestGiversLocked(player)
	}
	s.mu.Unlock()
	...
```

- オデュッセイア編の入り口なら仲間12人を与える(第9章)。
- 到着した部屋の**危険の判定**(第8章)。起きたことがあれば実況する。 `event` は `*flavor`(ポインタ)なので、中身を取り出して `*event` と渡す。
- クエスト依頼者の案内。
- あとはいつもの骨格。

## 7-14 `handleWho`

```go
func handleWho(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	client := conn.(*serverClient)
	s.mu.Lock()
	response, err := client.enqueueResponse(fmt.Sprintf("OK players=%d", len(s.players)))
	s.mu.Unlock()
	...
}
```

- 接続中の人数を返すだけ。 **CONNECT前でも使える**ので、名前の検査をしていない。

## 7-15 `handleQuit`

```go
func handleQuit(s *Server, conn net.Conn, name *string, parts []string) bool {
	if len(parts) != 1 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if *name != "" {
		if err := s.saveAndRemovePlayer(*name); err != nil {
			logger.Error("save_player_failed", "player", *name, "when", "quit", "error", err.Error())
			fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
			return false
		}
		logger.Info("player_quit", "player", *name)
		*name = ""
	}
	fmt.Fprintln(conn, "OK bye")
	return true
}
```

- 認証済みなら、保存して退室(7-5)。 **保存に失敗したら `ERR 500` を返して切断しない**(保存できないまま消すとデータが失われるため)。
- 成功したら `*name = ""` で「未接続」に戻す。 **こうしないと、接続を閉じるときの `defer`(7-21)がもう一度保存しようとしてしまう**。
- `OK bye` を送って `true`(接続終了)。未認証のまま `QUIT` しても `OK bye` で終了できる。

## 7-16 `handleTake`— アイテムを拾う

```go
	if !requireArgs(conn, parts, 2) {
		return false
	}
	...
	query := strings.Join(parts[1:], " ")
```

- `TAKE` の後ろは**複数の単語**でもよい(`TAKE Golden Fleece`)。
- `parts[1:]`:**スライスの切り出し**。「1番目から最後まで」。 `strings.Join(それ, " ")` で空白でつなぎ直し、1つの検索語 `query` にする。

```go
	locale := clientLocale(conn)
	itemID := ""
	if item := s.world.Items[query]; item != nil && item.availableTo(player, query) {
		itemID = query
	} else {
		for id, item := range s.world.Items {
			if item != nil && item.availableTo(player, id) && strings.EqualFold(item.Name.Get(locale), query) && (itemID == "" || id < itemID) {
				itemID = id
			}
		}
	}
	if itemID == "" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 ITEM_NOT_FOUND")
		return false
	}
```

- まず**IDそのもの**として探し、取れる状態ならそれを採用。
- ダメなら**名前**で全アイテムを探す(大文字小文字無視、複数あればID最小)。
- 見つからなければ `404 ITEM_NOT_FOUND`。
- `if 準備文; 条件 { } else { }`:`item` は `else` の中でも見える(が、ここでは使っていない)。

```go
	if !s.world.Items[itemID].Renewable {

		if s.unsavedTakes[*name] == nil {
			s.unsavedTakes[*name] = make(map[string]string)
		}
		s.unsavedTakes[*name][itemID] = player.RoomID
		s.world.Items[itemID].RoomID = ""
	}
```

- **再生しない**アイテム(普通のアイテム)の場合:
  - 「この人が、この部屋から、このアイテムを取った(まだ保存していない)」と `unsavedTakes` に記録。 内側の辞書が `nil` なら先に `make` で作る(`nil` の辞書には書けないため)。
  - **`RoomID = ""`**:アイテムを世界から取り除く = 他の人は取れなくなる(共有世界)。
- 再生アイテム(`Renewable`)は、部屋に残るので何もしない。

```go
	player.Inventory = append(player.Inventory, itemID)
	logger.Info("item_taken", "player", *name, "item", itemID, "room", player.RoomID, "renewable", s.world.Items[itemID].Renewable)
	s.checkQuestObjectiveLocked(player, "collect_item", itemID)
	takenInRoomID := player.RoomID
```

- `append`:スライスの末尾に追加(持ち物に入れる)。
- ログを出す。
- **クエストの自動判定**:「アイテムを集める」クエストを受けていれば進行する(第9章)。
- `takenInRoomID`:このあと死ぬ(部屋が変わる)可能性があるので、取った部屋を控えておく。

```go
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK taken=" + itemID)
	if err == nil {
		if event := s.applyTakeConsequencesLocked(player, *name, itemID); event != nil {
			s.broadcastFlavorLocked(takenInRoomID, *event)
		}
	}
	s.mu.Unlock()
	...
```

- 応答 `OK taken=ID`。
- 成功したら、**拾った結果の特殊効果**(ロトスの実を拾うと死ぬ、など。第9章)を適用し、起きたら取った部屋に実況する。

## 7-17 `handleDrop`— アイテムを置く(一番複雑)

「保存が途中で失敗しても、ファイルとメモリが食い違わない」ことを最優先に作られています。

```go
	index := -1
	for i, itemID := range player.Inventory {
		if itemID == query {
			index = i
			break
		}
	}
	if index == -1 {
		locale := clientLocale(conn)
		for i, itemID := range player.Inventory {
			if item := s.world.Items[itemID]; item != nil && strings.EqualFold(item.Name.Get(locale), query) && (index == -1 || itemID < player.Inventory[index]) {
				index = i
			}
		}
	}
	if index == -1 {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 ITEM_NOT_IN_INVENTORY")
		return false
	}
```

- 持ち物の中から探す。 `index := -1` は「まだ見つからない」の印(添字は0以上なので `-1` は使われない)。
- `for i, itemID := range ...`:ここでは**番号 `i`** も使う(後で削除する位置が必要)。
- `break`:ループをその場で抜ける。
- まずID一致、次に名前一致。見つからなければ `404 ITEM_NOT_IN_INVENTORY`。

```go
	itemID := player.Inventory[index]
	item := s.world.Items[itemID]
	if item == nil || (!item.Renewable && item.RoomID != "") {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
```

- 置くアイテムを確定。
- 異常検知:世界にそのアイテムが無い、または**普通のアイテムなのにすでにどこかの部屋にもある**(同じアイテムが2か所にある)なら `500`。

```go
	snapshot := *player
	snapshot.Inventory = append(make([]string, 0, len(player.Inventory)-1), player.Inventory[:index]...)
	snapshot.Inventory = append(snapshot.Inventory, player.Inventory[index+1:]...)
```

- 本物の `player` は**まだ変えず**、コピー(`snapshot`)の持ち物から該当を除いたものを作る。
- `player.Inventory[:index]`:先頭から `index` の手前まで。 `[index+1:]`:`index` の次から最後まで。この2つをつなぐと、 **`index` 番目だけが抜けたスライス**になる。
- `...`(`append(a, b...)`):スライス `b` の**中身を1つずつ展開**して渡す記法。
- `make([]string, 0, 容量)`:長さ0・容量を指定してスライスを作る(再確保を減らす)。

```go
	s.ioMu.Lock()
	var err error
	if item.Renewable {

		err = s.savePlayer(&snapshot)
	} else {
		var locations map[string]string
		locations, err = s.loadItemLocations()
		if err == nil {
			previousRoom, hadPreviousRoom := locations[itemID]
			locations[itemID] = player.RoomID
			err = s.writeItemLocations(locations)
			if err == nil {
				err = s.savePlayer(&snapshot)
				if err != nil {
					if hadPreviousRoom {
						locations[itemID] = previousRoom
					} else {
						delete(locations, itemID)
					}
					if rollbackErr := s.writeItemLocations(locations); rollbackErr != nil {
						logger.Error("restore_item_location_failed", "item", itemID, "error", rollbackErr.Error())
					}
				}
			}
		}
	}
	s.ioMu.Unlock()
```

- ファイル操作は `ioMu` の鍵の中で行う。
- **再生アイテム**:プレイヤーを保存するだけ。
- **普通のアイテム**:
  1. `itemdata.json`(アイテム位置)を読む。
  2. `previousRoom, hadPreviousRoom := locations[itemID]`:**元の記録**を控える(あとで戻すため)。
  3. 位置を「今の部屋」に更新して書く。
  4. プレイヤーの保存を試みる。
  5. **プレイヤーの保存に失敗したら、アイテム位置のファイルを元に戻す(ロールバック)**。元の記録が無かったなら `delete` で消す。
- `rollbackErr`:ロールバック自体が失敗した場合は、ログに残す(それ以上できることが無い)。

```go
	if err != nil {
		s.mu.Unlock()
		logger.Error("drop_item_failed", "player", *name, "item", itemID, "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}

	player.Inventory = snapshot.Inventory
	if !item.Renewable {
		item.RoomID = player.RoomID
	}
	logger.Info("item_dropped", "player", *name, "item", itemID, "room", player.RoomID)
	delete(s.unsavedTakes, *name)
```

- どこかで失敗していたら、**メモリは何も変えずに** `500` を返す。
- **全部成功してから初めて**、メモリ上の `player.Inventory` と `item.RoomID` を更新する。
- `delete(s.unsavedTakes, *name)`:保存されたので未保存の記録を消す。

```go
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK dropped=" + itemID)
	if err == nil {
		if event := s.applyDropConsequencesLocked(player, *name, itemID); event != nil {
			s.broadcastFlavorLocked(player.RoomID, *event)
		}
	}
	s.mu.Unlock()
	...
```

- 応答 `OK dropped=ID`。置いた結果の特殊効果(風の革袋を早く手放すと仲間が減る、など)を適用。

## 7-18 `handleInventory`

```go
	items := append(make([]string, 0, len(player.Inventory)), player.Inventory...)
	sort.Strings(items)
	data, err := json.Marshal(items)
	if err != nil || len("OK ")+len(data) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
```

- 持ち物を**コピー**してからソート(元の並びを変えないため)。
- `json.Marshal(items)` で `["item.a","item.b"]` の形にする。
- `err != nil || 長すぎる` を**1つの条件**にまとめている。
- 応答は `OK ["item.a","item.b"]`。

## 7-19 `handleTalk`

```go
	npcID := s.world.resolveNPCInRoom(player.RoomID, query, locale)
	if npcID == "" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 NPC_NOT_FOUND")
		return false
	}
	npc := s.world.NPCs[npcID]
```

- 同じ部屋のNPCを特定(第5章)。

```go
	if npc.Role != "enemy" && npc.hasMythRequirement() && !player.meetsMythRequirement(npc) {
		encounterRoomID := player.RoomID
		client := conn.(*serverClient)
		response, err := client.enqueueResponse("OK dead")
		if err == nil {
			s.respawnPlayerLocked(player, *name, "talk_unprepared", npcID, npc.Name.Get(locale))
			s.broadcastFlavorLocked(encounterRoomID, flavor{key: "talk_unprepared", player: *name, npc: npc})
		}
		s.mu.Unlock()
		...
	}
```

- **敵ではないNPC**に神話の前提条件があり、満たしていないときは、**話しかけただけで死亡**する。応答は `OK dead`。

```go
	dialogue := ""
	if lines := s.dialogueFor(player, npcID, npc); len(lines) > 0 {
		dialogue = lines[0].Get(locale)
	}
	if dialogue == "" || strings.TrimSpace(dialogue) == "" || !utf8.ValidString(dialogue) ||
		strings.IndexFunc(dialogue, unicode.IsControl) >= 0 || len("OK ")+len(dialogue) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
```

- 台詞の先頭をその言語で取り出す。
  - 台詞のリストは `s.dialogueFor(player, npcID, npc)` が選びます(`defeat.go`、8-7)。**その部屋の敵を全部倒したプレイヤーには `DialogueCleared`、そうでなければ `Dialogue`** です。同じNPCでも、敵を倒す前と後で、話す内容が変わります。
  - `if lines := ...; len(lines) > 0 {` は、`if 準備文; 条件 {` の形です。`lines` はこの `if` の中だけで使えます。
- **台詞が使えない場合は `500`**:空・空白だけ・不正なUTF-8・制御文字を含む・長すぎる。データの不備でプロトコルが壊れないための安全装置。

```go
	response, err := client.enqueueResponse("OK " + dialogue)
	if err == nil {
		logger.Info("npc_interaction", "player", *name, "npc", npcID, "room", player.RoomID)
		if npc.Guide {

			s.sendGuideLocked(*name, 1)
			s.sendDeathHintLocked(player, *name)
		}
		s.sendQuestHintLocked(player, npcID)
		s.talkEndingLocked(player, npc)
	}
```

- 応答は `OK 台詞`。
- そのあと:
  - **案内役**なら台詞の続き(2つ目以降)を通知として順に送る(第11章)。続けて、直前に死んでいれば**死因にちなんだ神話のヒント**を1回だけ送る(`sendDeathHintLocked`、第11章)。
  - **依頼者**なら、クエストの案内や進捗を通知(第9章)。
  - **エンディング**のあるNPCなら、エンディングの判定(第9章)。

## 7-20 `handleStatus`

```go
	player.regenLocked(time.Now(), s.effectTotalLocked(player, blessingRegenBonus), s.maxHPLocked(player))
	status := "healthy"
	if player.CombatTargetID != "" {
		status = "combat"
	}
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		MaxHP  int    `json:"max_hp"`
		Status string `json:"status"`
	}{player.HP, s.maxHPLocked(player), status})
```

- ここだけ `playerForUpdateLocked` ではなく、 `s.players[*name]` で取って**直接** `regenLocked` を呼ぶ(`STATUS` は退室中でも答えてよいため)。
- 戦闘中(`CombatTargetID` が空でない)なら `"combat"`、そうでなければ `"healthy"`。
- `max_hp` には定数ではなく **`s.maxHPLocked(player)`**(今の最大HP)を返します。クエストを達成した人は100より大きくなります(例:`{"hp":103,"max_hp":103,...}`)。
- 応答は `OK {"hp":80,"max_hp":100,"status":"healthy"}`。

## 7-21 `handleClient`— 接続ごとのメインループ

**1人の接続につき1つ、この関数が動き続けます。** `main.go` の `go server.handleClient(conn)` で起動されたものです。

```go
func (s *Server) handleClient(rawConn net.Conn) {
	conn := newServerClient(rawConn)
	var name string
	lastPlayer := ""
	openedAt := time.Now()
```

- 引数 `rawConn` は、 `Accept` で得た生の接続。
- `newServerClient(rawConn)` で**送信キュー付きの接続に包む**(第6章)。以降は `conn` を使う。
- `var name string`:この接続のプレイヤー名。 **最初は空(= 未認証)**。ハンドラに `&name`(ポインタ)で渡し、 `CONNECT` 成功で書き換えてもらう。
- `lastPlayer`:ログ用に、最後に分かったプレイヤー名を残す(`QUIT` で `name` が空に戻ってもログに名前を残すため)。
- `openedAt`:接続した時刻(接続時間のログ用)。

```go
	logger.Info("connection_open", "remote", conn.remote)
	s.abuse.noteConnection(hostOf(conn.remote), openedAt)
```

- 接続開始をログに記録。
- 同じ相手(IP)からの**接続の連打**を監視に記録(第13章)。 `hostOf` はポート番号を除いてIPだけにする。

```go
	defer func() {
		if name != "" {
			if err := s.saveAndRemovePlayer(name); err != nil {
				logger.Error("save_player_failed", "player", name, "when", "disconnect", "error", err.Error())
				s.mu.Lock()
				s.removePlayerLocked(name, true)
				s.mu.Unlock()
			}
		}
		logger.Info("connection_close", "remote", conn.remote, "player", lastPlayer, "duration_ms", time.Since(openedAt).Milliseconds())
		conn.Close()
	}()
```

- `defer func() { ... }()`:**無名関数を作って、その場で `()` を付けて予約**する形。この `handleClient` を**どう抜けても**(`QUIT`、切断、エラー)必ず実行される後片付け。
- 名前が残っている(= `QUIT` せずに切れた)なら、保存して退室させる。
- **保存に失敗したときの最終手段**:鍵を取って `removePlayerLocked(name, true)`(第2引数 `true` = 未保存で取ったアイテムを部屋に戻す)で、強制的に消す。
- 接続終了のログ(接続時間 `time.Since(openedAt).Milliseconds()` ミリ秒)。
- `conn.Close()`:接続を閉じる。

```go
	if _, err := fmt.Fprintln(conn, "OK hello proto=1"); err != nil {
		return
	}
```

- **接続直後にサーバーから送る決まり**のあいさつ(RFC)。 `Fprintln` は(書いたバイト数, エラー)を返すが、バイト数は使わないので `_`。
- 送れなかったら、すぐ終了(`defer` が後片付け)。

```go
	scanner := bufio.NewScanner(conn)
	var flood floodTracker

	for scanner.Scan() {
		parts := parseCommandParts(scanner.Text())
		if len(parts) == 0 {
			continue
		}
```

- `bufio.NewScanner(conn)`:接続から**1行ずつ**読む道具。
- `flood`:この接続のコマンド連打を数える(第13章)。
- `for scanner.Scan() {`:**1行読めるたびに繰り返し、接続が切れる/エラーになると終わる**。 `Scan()` は次の行が来るまで待つ。
- `scanner.Text()`:読んだ1行(改行を除いた文字列)。
- `parseCommandParts`(第10章)で単語に分ける。空行(単語が0個)は無視。

```go
		command := strings.ToUpper(parts[0])
		conn.setContext(name, command)
		logger.Info("command", "remote", conn.remote, "player", name, "command", command, "args", clip(strings.Join(parts[1:], " ")))
		if count, warn := flood.record(time.Now()); warn {
			logger.Warn("abuse_command_flood", "remote", conn.remote, "player", name, "commands", count, "window_ms", commandWindow.Milliseconds())
		}
```

- コマンド名を**大文字にする**(`look` でも `LOOK` でも通る)。
- ログ用のコンテキストを更新し、受け取ったコマンドを記録(引数は `clip` で長さを制限)。
- `flood.record` が「短時間に多すぎる」を判定。 **警告ログを出すだけで、コマンドは普通に処理する**(拒否はしない)。

```go
		handler, ok := commandHandlers[command]
		if !ok {
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
			continue
		}
		stop := handler(s, conn, &name, parts)
		conn.setContext(name, command)
		if name != "" {
			lastPlayer = name
		}
		if stop {
			return
		}
	}
```

- **7-9の辞書**から担当の関数を引く。無いコマンドは `400`。
- `handler(s, conn, &name, parts)`:**関数を呼ぶ**。 `&name` で `name` のポインタを渡す。
- 実行後、ログ用コンテキストを更新(`CONNECT` で名前が入ったかもしれないため)。
- `stop` が `true` なら `return`(→ `defer` の後片付けが走って接続終了)。

```go
	if err := scanner.Err(); err != nil {
		logger.Warn("connection_read_error", "remote", conn.remote, "player", name, "error", err.Error())
	}
}
```

- ループが終わった理由がエラー(1行が長すぎる、通信エラーなど)なら、警告ログに残す。正常な切断(`io.EOF`)は `Err()` が `nil` を返す。
- 関数の終わりで `defer` が実行される。

> これで **「接続 → 1行読む → 担当関数を呼ぶ → 応答を返す」** の流れの全体が分かりました。第8章以降は、この「担当関数」の中身(戦闘、クエストなど)です。

---

# 第8章 戦闘・危険・死亡(`hazard.go`, `combat.go`, `hardcore.go`)

「神話に逆らうと死ぬ」というゲームの核の部分です。短い順に `hazard.go` → `hardcore.go` → `combat.go` の順で読みます。

## 8-1 `hazard.go`(57行)— 部屋の危険と、出口を塞ぐ敵

```go
package main

func (s *Server) applyRoomHazardLocked(player *Player, name string, room *Room, locale string) *flavor {
	if room == nil || room.Hazard == nil {
		return nil
	}
	hazard := room.Hazard
	roomName := room.Name.Get(locale)
```

- import が無いファイル(他のファイルで定義された関数を使うだけ)。
- 「部屋に入った人に、その部屋の危険を適用する」。戻り値の `*flavor` は**実況用のデータ**で、何も起きなければ `nil`。
- 部屋が無い、または危険が無ければ即 `nil`。
- `roomName`:死亡メッセージに入れる部屋の名前(その言語で)。

```go
	switch hazard.Type {
	case "lethal":
		if description := room.Description.Get(locale); description != "" {
			s.sendPlayerEventLocked(name, "DEATH", description)
		}
		s.respawnPlayerLocked(player, name, "hazard_lethal", room.ID, roomName)
		return &flavor{key: "hazard_lethal", player: name, room: room}
```

- **`lethal`**:入った時点で必ず死亡。部屋の説明文があれば、本人の言語で `EVT PLAYER DEATH 説明文` を先に送信キューへ積む。
- `respawnPlayerLocked`(8-4)は持ち物などの死亡処理と復活を行い、死因と復活先を知らせる。説明文がある即死部屋では、個人向けの死亡通知が2件になる。
- `&flavor{...}`:実況データを作ってポインタを返す。

```go
	case "item_gate":
		if player.hasItem(hazard.RequiredItemID) {
			return nil
		}
		s.respawnPlayerLocked(player, name, "hazard_item", room.ID, roomName)
		return &flavor{key: "hazard_item", player: name, room: room}
```

- **`item_gate`**:必要アイテムを**持っていれば何も起きない**(`nil`)。持っていなければ死亡。

```go
	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name, "hazard_crew", room.ID, roomName)
			return &flavor{key: "hazard_crew_dead", player: name, room: room}
		}
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return &flavor{key: "hazard_crew_loss", player: name, room: room, n: lost}
```

- **`crew_gate`**:通過に必要な人数(`MinPartyTotal`)があり、 **`Crew + 1`(仲間+自分)** が足りなければ全滅=死亡。
- 足りていれば仲間が `CrewLoss` 人減る。 `spendCrewLocked`(第9章)は**実際に減った人数**を返す。それを `n: lost` で実況に入れる。

```go
	case "crew_cost":
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return &flavor{key: "hazard_crew_loss", player: name, room: room, n: lost}
	}
	return nil
}
```

- **`crew_cost`**:条件なしで仲間が減る。
- どのcaseにも当たらなければ最後の `return nil`。

```go
func (s *Server) blockingEnemyLocked(player *Player, roomID string) (string, *NPC) {
	blockID := ""
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID || npc.Role != "enemy" || player.enemyHP(id, npc) <= 0 {
			continue
		}
		if player.FledFrom[id] {
			continue
		}
		if blockID == "" || id < blockID {
			blockID = id
		}
	}
	if blockID == "" {
		return "", nil
	}
	return blockID, s.world.NPCs[blockID]
}
```

- 「**この人の出口を塞いでいる敵**」を探す。
- 次のどれかに当てはまるNPCは**飛ばす**(`continue`):
  - 同じ部屋にいない、または敵ではない。
  - **この人にとってすでに倒されている**(`enemyHP <= 0`)。
- さらに `FledFrom[id]`(**逃げ切った敵**)も飛ばす。 `FledFrom` が `nil` の辞書でも、**読むのは安全**なので問題ない。
- 残った中でID最小の1体を返す。 戻り値は (ID, NPC)。いなければ `"", nil`。
- `MOVE` と `FLEE` で使われる。

## 8-2 `hardcore.go`(158行)— 仲間の効果と、死亡のペナルティ

```go
const (
	allyDamageBonus = 5

	allyCounterReductionPercent = 20

	maxAllyBonusCount = 3
)
```

- 同じ部屋の仲間1人につき:与ダメージ **+5**、受けるダメージ **20%軽減**。効果が付く仲間は**最大3人**まで。

```go
func (s *Server) alliesInRoomLocked(name string) []string {
	player := s.players[name]
	group := s.groups[s.groupByPlayer[name]]
	if player == nil || group == nil {
		return nil
	}
	var allies []string
	for memberName := range group.Members {
		member := s.players[memberName]
		if memberName != name && member != nil && !member.exiting && member.RoomID == player.RoomID {
			allies = append(allies, memberName)
		}
	}
	return allies
}
```

- 「**同じグループで、同じ部屋にいる仲間**」の名前の一覧。
- `s.groups[s.groupByPlayer[name]]`:名前→グループID→グループ、と**2段で引く**。グループに入っていなければ `groupByPlayer[name]` が `""` になり、 `groups[""]` は `nil`。
- `var allies []string`:空のまま(`nil`)のスライス。 `append` すると自動的に作られる。
- `for memberName := range group.Members`:**辞書を `range` で1つだけ受け取ると、キーだけ**が取れる(集合の要素を回す定番の形)。
- 自分自身、接続していない人、退室中の人、別の部屋の人は除く。

```go
func allyBonusCount(allies []string) int {
	if len(allies) > maxAllyBonusCount {
		return maxAllyBonusCount
	}
	return len(allies)
}
```

- 人数の**上限を3に切り詰める**。

```go
type deathOutcome string

const (
	outcomeNothingLost deathOutcome = ""
	outcomeLost        deathOutcome = "lost"
	outcomeKept        deathOutcome = "kept"
)
```

- `type deathOutcome string`:「文字列だけど、**死亡の結果を表す専用の型**」。 他の文字列と取り違える事故を型で防ぐ。
- 3つの値:何も失わなかった(持ち物が無い)、**失った**、**仲間のおかげで守られた**。

```go
func (s *Server) applyDeathPenaltyLocked(player *Player, name string) deathOutcome {
	protected := len(s.alliesInRoomLocked(name)) > 0
	player.EnemyHP = nil
	player.FledFrom = nil
	if s.world == nil {
		return outcomeNothingLost
	}
```

- 同室に仲間が1人でもいれば `protected`(持ち物を守れる)。
- `EnemyHP = nil`:**敵のHPの記録を消す** = 傷つけた敵が全回復する。 `FledFrom = nil`:逃げ切った記録も消える。

```go
	var kept, lost []string
	for _, itemID := range player.Inventory {
		if item := s.world.Items[itemID]; item == nil || item.RewardOnly {
			kept = append(kept, itemID)
		} else {
			lost = append(lost, itemID)
		}
	}
	if len(lost) == 0 {
		return outcomeNothingLost
	}
	if protected {
		return outcomeKept
	}
```

- 持ち物を「残るもの(`kept`)」と「失うもの(`lost`)」に分ける。 **`RewardOnly`(エンディングの記念品)は残る**。
- `var kept, lost []string`:同じ型の変数を2つまとめて宣言。
- 失うものが無ければ `outcomeNothingLost`。仲間に守られていれば何も変えずに `outcomeKept`。

```go
	if err := s.returnItemsHomeLocked(name, lost); err != nil {
		logger.Error("return_lost_items_failed", "player", name, "error", err.Error())
		return outcomeNothingLost
	}
	player.Inventory = kept
	return outcomeLost
}
```

- 保存済みの所有記録を更新してから、再生しないアイテムを元の置き場所へ戻し、持ち物を「残るもの」だけにします。保存前に別の人が取得すると、旧所有者のセーブと二重所有になるため、この順序にします。
- 保存失敗時はエラーをログに残し、持ち物と未保存の取得記録を保持します。仲間による保護を示す`outcomeKept`は返しません。
- `outcomeLost` を返す。

```go
func (s *Server) returnItemsHomeLocked(name string, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}
	lost := make(map[string]bool, len(itemIDs))
	returned := make(map[string]string)
	for _, itemID := range itemIDs {
		lost[itemID] = true
		item := s.world.Items[itemID]
		if item.Renewable {
			continue
		}
		roomID := item.HomeRoomID
		if roomID == "" {
			roomID = s.world.StartRoomID
		}
		returned[itemID] = roomID
	}

	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	locations, err := s.loadItemLocations()
	if err != nil {
		return err
	}
	previousLocations := make(map[string]string, len(locations))
	for itemID, roomID := range locations {
		previousLocations[itemID] = roomID
	}
	for itemID, roomID := range returned {
		locations[itemID] = roomID
	}
	if len(returned) > 0 {
		if err := s.writeItemLocations(locations); err != nil {
			return err
		}
	}

	// Remove the old saved ownership before another player can take these items.
	// Keep the rest of the saved state rather than saving the pre-respawn HP and room.
	players, err := s.loadPlayers()
	if err == nil {
		if saved := players[name]; saved != nil {
			var inventory []string
			for _, itemID := range saved.Inventory {
				if !lost[itemID] {
					inventory = append(inventory, itemID)
				}
			}
			saved.Inventory = inventory
			err = s.writePlayers(players)
		}
	}
	if err != nil {
		if len(returned) > 0 {
			if rollbackErr := s.writeItemLocations(previousLocations); rollbackErr != nil {
				logger.Error("restore_item_location_failed", "player", name, "error", rollbackErr.Error())
			}
		}
		return err
	}

	for itemID, roomID := range returned {
		s.world.Items[itemID].RoomID = roomID
		delete(s.unsavedTakes[name], itemID)
	}
	return nil
}
```

- 失う全アイテムのIDと、再生しないアイテムの帰還先を準備します。元の部屋が無ければ開始部屋を使います。
- ファイル用のロックを取り、元の位置情報をコピーしてから帰還先を保存します。続いて旧所有者の保存済み持ち物から失ったIDだけを除きます。保存済みのHP・部屋・クエスト進捗は上書きしません。
- プレイヤー保存が失敗したら、アイテム位置を以前の状態へ戻す処理を試みます。その復元も失敗した場合はログに残します。再起動時は保存済み所有者の記録が部屋の位置情報に優先します。
- 保存が成功してから、メモリ上の部屋へ返却し、未保存の取得記録から除きます。呼び出し元は共通ロックを保持しているため、途中で別の人が取得できません。

```go
func (s *Server) shareVictoryLocked(attacker string, allies []string, npcID string, npc *NPC) {
	for _, allyName := range allies {
		ally := s.players[allyName]
		if ally == nil || ally.exiting || ally.enemyHP(npcID, npc) <= 0 {
			continue
		}
		ally.setEnemyHP(npcID, 0)
		logger.Info("victory_shared", "player", allyName, "ally_of", attacker, "npc", npcID)
		ally.CombatTargetID = ""
		s.checkQuestObjectiveLocked(ally, "defeat_npc", npcID)
		locale := s.localeOfLocked(allyName)
		s.sendPlayerEventLocked(allyName, "TEAM", LocalizedText{
			"en": "Your ally %s defeated %s, and you share the victory.",
			"ja": "仲間の%sが%sを打ち倒した。その手柄はあなたにも与えられる。",
		}.Format(locale, attacker, npc.Name.Get(locale)))
	}
}
```

- 誰かが敵を倒したとき、**同室の仲間にも撃破を共有**する。
- すでに倒していた仲間は飛ばす。
- 仲間の `EnemyHP` を0にして戦闘を解除し、**その仲間のクエストも進める**(`defeat_npc`)。
- `LocalizedText{ "en": ..., "ja": ... }.Format(locale, ...)`:**その場で言語別の文字列の辞書を作り、すぐ `Format`** する形。 `%s` が `Format` に渡した引数で順に置き換わる。
- `sendPlayerEventLocked(名前, "TEAM", 文)`:その人だけに通知を送る(第11章)。

## 8-3 `combat.go`(325行) 前半 — 乱数・防御・反撃の軽減

```go
import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"
)

var randDamage = func(min, max int) int {
	return min + rand.IntN(max-min+1)
}
```

- `math/rand/v2`:乱数。
- **`var randDamage = func(...) {...}`**:関数を**変数に入れている**。普通の `func randDamage(...)` ではなくこうする理由は、**テストのときに固定値を返す関数に差し替えられる**から。
- `rand.IntN(n)`:0以上n未満の乱数。 `min + rand.IntN(max-min+1)` で、 **min以上max以下**の乱数になる(8〜14なら `8 + IntN(7)`)。

```go
// guardReductionPercent is how much a DEFEND stance cuts the next counter-attack.
const guardReductionPercent = 50

func takeGuard(player *Player) int {
	if !player.guarding {
		return 0
	}
	player.guarding = false
	return guardReductionPercent
}
```

- `guardReductionPercent = 50`:`DEFEND` で身構えたとき、次の反撃を**何%減らすか**。
- `takeGuard`:**身構えの効果を「1回だけ」取り出す**関数です。
  - 身構えていなければ(`guarding` が `false`)、軽減は0%。
  - 身構えていれば、`guarding` を**`false` に戻してから**50を返します。つまり、この関数を呼んだ時点で身構えは**使い切り**になります。反撃を1回受けるたびにこの関数を呼ぶので、「次の反撃だけ半分」になります。
  - 「状態を変えながら値を返す」関数なので、**同じ反撃の中で2回呼ばない**ことが大事です。

```go
// reduceCounter lowers a counter-attack by percent (capped), never below 1 damage.
// A negative percent (a bad item) makes the counter-attack hurt more, up to maxCounterIncrease.
func reduceCounter(counter, percent int) int {
	percent = max(-maxCounterIncrease, min(percent, maxCounterReduction))
	if percent == 0 {
		return counter
	}
	return max(1, counter*(100-percent)/100)
}
```

- **反撃のダメージを、軽減率(%)で減らす**関数です。反撃の計算はすべてここを通ります。
- `percent = max(-maxCounterIncrease, min(percent, maxCounterReduction))`
  - `min(percent, maxCounterReduction)`:軽減は**最大80%**まで(`maxCounterReduction`、9-3)。仲間3人(60%)に祝福(20%)を重ねても、ダメージが0になって無敵になることはありません。
  - `max(-maxCounterIncrease, ...)`:軽減が**マイナス**(悪いアイテムの効果)のときは、**増えるのは最大50%まで**(`maxCounterIncrease`)。
  - 2つを組み合わせて、`percent` を「-50〜80」の範囲に収めます。`min`/`max` は、Goの組み込み関数(2つの値の小さいほう/大きいほう)です。
- `if percent == 0 { return counter }`:軽減も増加も無ければ、そのまま返します。
- `counter*(100-percent)/100`
  - 軽減率を掛けた後のダメージです。例:反撃10で20%軽減なら `10*(100-20)/100 = 8`。**マイナスの軽減**(`percent = -20`)なら `10*(120)/100 = 12` と、**増えます**。
  - 整数どうしの割り算なので端数は切り捨てです。
- `max(1, ...)`:**最低でも1ダメージ**は受けます。

> **ねらい**:反撃を減らす要素(仲間・祝福・アイテム・身構え)と、増やす要素(悪いアイテム)が、**全部「%の足し算」**で1つの数にまとまり、最後にこの関数で1回だけ計算されます。要素が増えても、計算の場所は1か所のままです。

## 8-4 `respawnPlayerLocked`(死亡と復活)

```go
func (s *Server) respawnPlayerLocked(player *Player, name, cause, subject string, args ...any) {
	player.LastDeathSubject = subject
	outcome := s.applyDeathPenaltyLocked(player, name)
	s.notifyDeathLocked(name, cause, outcome, args...)
	oldRoomID := player.RoomID
	logger.Info("player_died", "player", name, "cause", cause, "room", oldRoomID, "belongings", string(outcome))
```

- `name, cause, subject string`:3つの文字列引数をまとめて書いています。
  - `cause`:**死因の種類**(`"attack_counter"`、`"hazard_lethal"` など)。メッセージの文面を選ぶのに使います(第11章)。
  - `subject`:**死因になったもののID**(敵・部屋・アイテム)。次の行で `LastDeathSubject` に記録され、モイライが神話のヒントを返すときの手がかりになります。
- `args ...any`:**可変長引数**。 `...any` は「任意の型の値を何個でも」。 死因ごとに必要な情報(敵の名前など)が違うので、このように受け取る。渡すときも `args...` と展開して `notifyDeathLocked` に渡す。
- 順番:① 死因を記録 → ② 持ち物のペナルティ(8-2) → ③ 本人へ死亡メッセージを通知 → ④ ログ。
- 死んでも**書き換えない**もの:`MaxHPBonus`(クエストで増えた最大HP、第3章)、エンディング、クエストの達成記録。死んで失うのは、HPと持ち物と、敵に関する記録(傷つけた敵の `EnemyHP`、逃げ切った記録の `FledFrom`)だけです。

```go
	destination := defaultStartRoomID
	if s.world != nil {
		destination = s.world.StartRoomID
	}
	player.HP = respawnHP
	player.guarding = false
	player.CombatTargetID = ""
	player.RoomID = destination
	if oldRoomID == destination {
		return
	}
```

- 復活先は開始部屋。HPは20、戦闘状態と身構えを解除。
- すでに開始部屋で死んだ(部屋が変わらない)なら、入退室の通知は要らないので終了。

```go
	for playerName, current := range s.players {
		recipient := s.clients[playerName]
		if recipient == nil {
			continue
		}
		switch current.RoomID {
		case oldRoomID:
			recipient.enqueueEvent("EVT ROOM PRESENCE LEAVE " + name)
		case destination:
			recipient.enqueueEvent("EVT ROOM PRESENCE ENTER " + name)
		}
	}
}
```

- 7-13と同じ入退室の通知。死んだ部屋にいる人に `LEAVE`、復活した部屋にいる人に `ENTER`。

## 8-5 `combatResult` と `handleAttack`

```go
type combatResult struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}
```

- `ATTACK` の応答JSONの形(RFCで決まっている)。`Status` に入る値は次のとおりです。

| `status` | 意味 |
|---|---|
| `combat` | 戦闘が続いている(敵は生きている) |
| `victory` | 敵を倒した |
| `dead` | 反撃や神話の罠で、自分が死んだ |
| `overwhelmed` | 勝てない敵に押し返された(仲間が減る) |
| `wounded` | 敵以外のNPCを傷つけた(まだ死んでいない) |
| `murder` | 敵以外のNPCを殺して、自分が死んだ |
| `smitten` | 強いNPCに手を出して、一撃で殺された |

- 最初の4つがもとの設計で、下の3つは**敵以外を攻撃できる**ようにしたとき(ケース3・4)に増えました。

```go
func handleAttack(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 2) { return false }
	if *name == "" { ... }
	query := strings.Join(parts[1:], " ")
	locale := clientLocale(conn)

	s.mu.Lock()
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil || s.world.Rooms[player.RoomID] == nil { ...500... }
	npcID := s.world.resolveNPCInRoom(player.RoomID, query, locale)
	if npcID == "" { ...404 NPC_NOT_FOUND... }
	npc := s.world.NPCs[npcID]
	enemyHP := player.enemyHP(npcID, npc)
	if npc.Role == "enemy" && enemyHP <= 0 {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 405 NPC_NOT_HOSTILE")
		return false
	}
	encounterRoomID := player.RoomID
```

- いつもの骨格(7-10)の前半。NPCを特定する。
- `if npc.Role == "enemy" && enemyHP <= 0`
  - **この人にとってすでに倒されている敵**だけを `405 NPC_NOT_HOSTILE` で断ります。
  - 以前は「敵でないNPC」もここで断っていましたが、今は**敵以外も攻撃できる**ので、条件を「敵で、かつ倒し済み」だけにしました(RFCとの違いとして、READMEに書いてあります)。
- `encounterRoomID`:戦った部屋を控える(死んで部屋が変わっても、**戦った部屋に実況を送る**ため)。

```go
	var event flavor
	var result combatResult

	switch {
```

- あとで埋める2つの変数:実況データと応答の中身。
- `switch {` で、上から順に条件を見る。**最初に当てはまったケースだけ**が実行されます。ケースの並びが優先順位です。

### ケース1:勝てない敵

```go
	case npc.Unwinnable:
		lost := spendCrewLocked(player, npc.CrewLossOnAttack)
		player.CombatTargetID = ""
		event = flavor{key: "attack_unwinnable", player: *name, npc: npc, n: lost}
		result = combatResult{player.HP, enemyHP, 0, "overwhelmed"}
```

- **勝てない敵**(`Unwinnable`):HPは減らず、 **仲間が減る**だけ。戦闘状態は解除。
- `combatResult{player.HP, enemyHP, 0, "overwhelmed"}`:フィールドを**定義順に並べて**値を入れる書き方。

### ケース2:神話の前提を満たしていない(即死)

```go
	case npc.hasMythRequirement() && !player.meetsMythRequirement(npc):
		s.respawnPlayerLocked(player, *name, "attack_unprepared", npcID, npc.Name.Get(locale))
		event = flavor{key: "attack_unprepared", player: *name, npc: npc}
		result = combatResult{0, enemyHP, 0, "dead"}
```

- 必要なアイテムやクエストが無いのに攻撃すると、**即死**。ゲームの核心。
- `respawnPlayerLocked` の4つ目の引数 `npcID` が**死因のID**(8-4)です。
- 応答の `attacker_hp` は0(死んだので)。復活後のHP20とは別に、結果として0を返している。

### ケース3:強いNPCに手を出した(一撃で死ぬ)

```go
	case npc.Role != "enemy" && (npc.Mighty || npc.Guide):
		// Gods, sorcerers and the like answer a blow with death before it lands.
		s.respawnPlayerLocked(player, *name, "attack_mighty", npcID, npc.Name.Get(locale))
		event = flavor{key: "attack_mighty", player: *name, npc: npc}
		result = combatResult{0, enemyHP, 0, "smitten"}
```

- `npc.Role != "enemy" && (npc.Mighty || npc.Guide)`
  - **敵ではなく**、かつ**`Mighty`(神・魔女など)か`Guide`(モイライ)**のNPCを攻撃した場合です。`&&` は「かつ」、`||` は「または」で、括弧で優先順位をはっきりさせています。
  - メデイア・キルケー・アテナ・アキレウス・モイライなど10人が対象です(`world.json` の `"mighty": true`)。
- **反撃を待たず、その場で死にます**。ダメージの計算も要りません。`result` の `Damage` は0、`status` は `"smitten"`(打ちのめされた)です。
- 勝てない戦いを挑ませないための、神話らしい設計です。ゲームを進めるうえで話しかけるべき相手を、うっかり殺せないようにもなっています。

### ケース4:一般人を攻撃した

```go
	case npc.Role != "enemy":
		// An ordinary person cannot fight back, but killing them ends the game for you.
		damage := max(1, randDamage(combatMinDamage, combatMaxDamage)+s.effectTotalLocked(player, blessingDamageBonus))
		enemyHP = max(0, enemyHP-damage)
		player.setEnemyHP(npcID, enemyHP)
		if enemyHP == 0 {
			s.respawnPlayerLocked(player, *name, "attack_murder", npcID, npc.Name.Get(locale))
			event = flavor{key: "attack_murder", player: *name, npc: npc}
			result = combatResult{0, 0, damage, "murder"}
		} else {
			event = flavor{key: "attack_wounded", player: *name, npc: npc, n: damage}
			result = combatResult{player.HP, enemyHP, damage, "wounded"}
		}
```

- ここまで来る `npc.Role != "enemy"` は、**敵ではなく、強くもない人**(村人、王様、依頼者など)です。
- `damage := max(1, 8〜14の乱数 + 祝福・アイテムの与ダメージ加算)`
  - 敵への攻撃と同じ計算ですが、**仲間の加勢は付きません**(一般人を仲間と一緒に殴らせない)。`effectTotalLocked` は祝福とアイテムの効果の合計(9-4)。
  - `max(1, ...)`:悪いアイテムでダメージがマイナスになっても、**最低1**は与えます。
- `enemyHP = max(0, enemyHP-damage)`:HPを減らし、**0未満にはしません**。`player.setEnemyHP` で、この人専用の記録に保存します(第3章)。
- **HPが0になった(殺した)**:`respawnPlayerLocked` で**自分が死亡・復活**します。死因は `"attack_murder"`。メッセージは「運命の女神たちが、その報いにあなたの糸を断ち切った」です。これが「殺した瞬間ゲームオーバー」の実体です。
- **まだ生きている**:`"wounded"` を返して終わりです。**反撃はしません**(`player.HP` は減りません)。戦闘状態(`CombatTargetID`)にもしないので、`FLEE` の対象にもなりません。
- 死ぬと `EnemyHP` が消える(8-2)ので、殴った一般人は**元気な状態に戻ります**。

### ケース5:通常の戦闘

```go
	default:

		allies := s.alliesInRoomLocked(*name)
		bonus := allyBonusCount(allies)
		damage := max(1, randDamage(combatMinDamage, combatMaxDamage)+bonus*allyDamageBonus+s.effectTotalLocked(player, blessingDamageBonus))
		enemyHP -= damage
		if enemyHP < 0 {
			enemyHP = 0
		}
		player.setEnemyHP(npcID, enemyHP)
```

- 同室の仲間の数(最大3)で `bonus` を決める。
- 与ダメージ = 8〜14の乱数 + 仲間1人につき5 + **祝福とアイテムの加算**(アポロンの祝福は+3、アイテムは±)。 `max(1, ...)` で**最低1**。
- 敵のHPから引き、**0未満にならないように**切り詰める。 `-=` は「引いて代入」。
- **このプレイヤー専用のHP記録**に保存(`setEnemyHP`)。

```go
		if enemyHP == 0 {
			player.CombatTargetID = ""
			s.checkQuestObjectiveLocked(player, "defeat_npc", npcID)
			s.shareVictoryLocked(*name, allies, npcID, npc)
			event = flavor{key: "attack_defeat", player: *name, npc: npc}
			result = combatResult{player.HP, 0, damage, "victory"}
```

- **倒した場合**:戦闘解除、 「敵を倒す」クエストを進行(クエストが完了すれば最大HPが上がる、9-2)、仲間に撃破を共有。 `status` は `"victory"`。

```go
		} else {
			player.CombatTargetID = npcID
			counter := randDamage(counterMinDamage, counterMaxDamage)
			counter = reduceCounter(counter, bonus*allyCounterReductionPercent+s.effectTotalLocked(player, blessingCounterReduction)+takeGuard(player))
			player.HP -= counter
```

- **倒せなかった場合**:戦闘中の相手を記録し、反撃を受ける。
- 反撃のダメージは7〜14。そこから、次の3つの**軽減率(%)を足した値**で `reduceCounter`(8-3)にかけます。
  - `bonus*allyCounterReductionPercent`:仲間1人につき20%。
  - `s.effectTotalLocked(player, blessingCounterReduction)`:アテナの祝福(20%)と、持っているアイテムの効果(防具は+、悪い物は-)の合計。
  - `takeGuard(player)`:`DEFEND` の身構えがあれば50%(使い切り)。
- このため、**仲間・祝福・アイテム・身構えの全部が、1つの式にまとまっています**。

```go
			if player.HP <= 0 {
				s.respawnPlayerLocked(player, *name, "attack_counter", npcID, npc.Name.Get(locale))
				event = flavor{key: "attack_struck_down", player: *name, npc: npc}
				result = combatResult{0, enemyHP, damage, "dead"}
			} else {
				event = flavor{key: "attack_hit", player: *name, npc: npc, n: damage, m: counter}
				result = combatResult{player.HP, enemyHP, damage, "combat"}
			}
		}
	}
```

- HPが尽きたら死亡・復活、 `status` は `"dead"`。
- 生きていれば `status` は `"combat"`。実況には与えたダメージ `n` と受けたダメージ `m` を入れる。

```go
	logger.Info("combat_attack", "player", *name, "npc", npcID, "status", result.Status, "damage", result.Damage, "attacker_hp", result.AttackerHP, "target_hp", result.TargetHP)
	data, err := json.Marshal(result)
	if err != nil { ...500... }
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil {
		s.broadcastFlavorLocked(encounterRoomID, event)
	}
	s.mu.Unlock()
	...
```

- ケースが終わったら、 結果をログに残し、JSONにして応答する。
- 成功したら、 **戦った部屋の全員**に実況を送る。

> **ターンの流れ(READMEの「Combat System」)**:`ATTACK` 1回が1ラウンドで、**プレイヤーが先に攻撃**します。敵が生き残れば、**同じラウンドの中で**すぐ反撃します。別の先手判定は無く、`DEFEND`・`FLEE` は「攻撃のかわりに選ぶ行動」です。

## 8-6 `handleFlee`(逃げる)と `handleDefend`(身構える)

```go
	targetID := player.CombatTargetID
	npc := s.world.NPCs[targetID]
	if targetID == "" || npc == nil || npc.RoomID != player.RoomID || npc.Role != "enemy" || player.enemyHP(targetID, npc) <= 0 {

		player.CombatTargetID = ""
		player.guarding = false
		targetID, npc = s.blockingEnemyLocked(player, player.RoomID)
		if npc == nil {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 407 NOT_IN_COMBAT")
			return false
		}
	}
	encounterRoomID := player.RoomID
```

- 戦闘中の敵を逃げる対象にする。戦闘中でなくても、 **出口を塞がれている**なら、その敵から逃げられる。
- 保存された戦闘対象が別の部屋・非敵・撃破済み・存在しないNPCなら、その対象と防御状態を解除し、現在の部屋を塞ぐ敵を調べます。
- どちらでもなければ `407 NOT_IN_COMBAT`。
- `targetID, npc = ...`:すでにある変数に**2つ同時に代入**。

```go
	fleeSucceeds := npc.FleeAccurate || (npc.FleeSucceedsOnce && !player.FledFrom[targetID])
```

- **逃走の成否**:
  - `FleeAccurate`:その敵からは必ず成功。
  - `FleeSucceedsOnce`:**まだこの敵から逃げ切っていなければ**成功(1回だけ)。 `!player.FledFrom[targetID]` は「記録が無い or false」。
  - どちらでもなければ失敗。

```go
	var result string
	var event flavor
	if fleeSucceeds {
		result = "success"
		event = flavor{key: "flee_success", player: *name, npc: npc}
		player.CombatTargetID = ""
		player.guarding = false
		if player.FledFrom == nil {
			player.FledFrom = make(map[string]bool)
		}
		player.FledFrom[targetID] = true
	} else {
		counter := reduceCounter(randDamage(counterMinDamage, counterMaxDamage), s.effectTotalLocked(player, blessingCounterReduction)+takeGuard(player))
		player.HP -= counter
		if player.HP <= 0 {
			s.respawnPlayerLocked(player, *name, "flee_failed", targetID, npc.Name.Get(locale))
			result = "failure_dead"
			event = flavor{key: "flee_dead", player: *name, npc: npc}
		} else {
			result = "failure"
			event = flavor{key: "flee_hit", player: *name, npc: npc, n: counter}
		}
	}
```

- **成功**:戦闘解除、 `FledFrom[敵] = true` と記録(以後その敵は出口を塞がない)。 辞書が `nil` なら先に `make`。
- **失敗**:反撃を受ける。仲間の軽減は付かず、**祝福とアイテムの軽減率と、身構え**だけが効きます(`reduceCounter`、8-3)。HPが尽きれば死亡。

```go
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		Result string `json:"result"`
	}{player.HP, result})
```

- 応答は `OK {"hp":..,"result":"success|failure|failure_dead"}`。残りは骨格どおり。

### `handleDefend`— 身構える

```go
	npc := s.world.NPCs[player.CombatTargetID]
	if player.CombatTargetID == "" || npc == nil {
		player.CombatTargetID = ""
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 407 NOT_IN_COMBAT")
		return false
	}
	player.guarding = true
	encounterRoomID := player.RoomID
	logger.Info("combat_defend", "player", *name, "npc", player.CombatTargetID, "hp", player.HP)
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		Result string `json:"result"`
	}{player.HP, "braced"})
```

- `DEFEND` は**RFCに無い独自コマンド**です(RFCが例として挙げた追加コマンドの1つ)。
- **戦闘中でなければ** `407 NOT_IN_COMBAT`。戦闘中なら `player.guarding = true` にして、攻撃もダメージもせず、次の反撃で `takeGuard`(8-3)が50%を引きます。
- 応答は `OK {"hp":80,"result":"braced"}`。部屋の全員にも実況(`"defend"`)が届きます。
- 身構えは**1回の反撃で使い切り**です。移動したり、逃げたり、死んだりしても解除されます。

## 8-7 `defeat.go`(44行)— 倒した状態と、突破後のセリフ

**役割**:「このプレイヤーが敵を倒したか」「この部屋を突破したか」を調べ、その結果を `LOOK` の表示と、NPCのセリフの切り替えに使います。ファイル全体が1つの機能です。

```go
func (p *Player) hasDefeated(npcID string, npc *NPC) bool {
	return npc != nil && npc.Role == "enemy" && !npc.Unwinnable && p.enemyHP(npcID, npc) <= 0
}
```

- `hasDefeated`:**このプレイヤーが、その敵を倒したか**。次の4つが**全部**成り立つと `true` です。
  - `npc != nil`:NPCが実在する。
  - `npc.Role == "enemy"`:敵である。
  - `!npc.Unwinnable`:**勝てない敵ではない**。勝てない敵は「倒す」ことが無いので、対象から外します。
  - `p.enemyHP(npcID, npc) <= 0`:**このプレイヤーから見た残りHPが0以下**(第3章)。
- 倒した記録は、すでにある **プレイヤーごとの敵HP**(`EnemyHP`)を使っているだけです。新しいデータは何も増えていません。だから**死ぬと `EnemyHP` が消えて、敵が元に戻ります**(8-2)。

```go
func (s *Server) defeatedNPCsLocked(player *Player, roomID string) []string {
	var ids []string
	for id, npc := range s.world.NPCs {
		if npc != nil && npc.RoomID == roomID && player.hasDefeated(id, npc) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
```

- `defeatedNPCsLocked`:**その部屋にいる、このプレイヤーが倒したNPCのID一覧**です。`LOOK` の `defeated` に入れます(7-12)。
- `var ids []string`:空のスライスを `var` で作ります。**`nil` のスライス**で、`append` すると自動で増えます。倒した敵が1人もいなければ `nil` のままで、JSONにしたとき `omitempty` で消えます。
- `sort.Strings(ids)`:辞書の `range` は順番がバラバラなので、**毎回同じ並び**になるよう並べ替えます。

```go
func (s *Server) roomClearedLocked(player *Player, roomID string) bool {
	found := false
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID || npc.Role != "enemy" || npc.Unwinnable {
			continue
		}
		if !player.hasDefeated(id, npc) {
			return false
		}
		found = true
	}
	return found
}
```

- `roomClearedLocked`:**部屋を突破したか**。「倒せる敵が1人以上いて、**全員**倒した」ときだけ `true` です。
- 敵でない人、勝てない敵は `continue` で**数えません**。
- 1人でも倒していない敵がいれば、その場で `return false`(早期リターン)。
- `found` は「倒せる敵が1人でもいたか」の印です。**敵のいない部屋は `false`** にしたいので、最後に `return found` とします(敵がいないのに「突破済み」にならないように)。

```go
func (s *Server) dialogueFor(player *Player, npcID string, npc *NPC) []LocalizedText {
	if len(npc.DialogueCleared) > 0 && npc.Role != "enemy" && s.roomClearedLocked(player, npc.RoomID) {
		return npc.DialogueCleared
	}
	return npc.Dialogue
}
```

- `dialogueFor`:NPCが**このプレイヤーに話す台詞のリスト**を選びます。`TALK` が呼びます(7-19)。
- 次の3つが**全部**成り立つときだけ、突破後の台詞 `DialogueCleared` を返します。それ以外は、いつもの `Dialogue` です。
  - `len(npc.DialogueCleared) > 0`:突破後の台詞が**設定されている**。
  - `npc.Role != "enemy"`:話し相手が敵ではない(敵は突破後の台詞を持たない)。
  - `s.roomClearedLocked(player, npc.RoomID)`:そのNPCのいる部屋を**突破している**。
- 引数 `npcID` は今は使っていません(将来、NPCごとの条件を足せるように残してあります)。Goでは**使わない引数は許される**(使わない変数や import は許されないのと違う点です)。

> **ねらい**:「敵を倒すと、世界が少し変わる」を、新しいデータなしで実現しています。倒した状態は `EnemyHP`、突破の判定はこのファイルの関数、見せ方はGUI(15-5)、と役割が分かれています。

---

# 第9章 クエスト・エンディング・アイテム効果・オデュッセイア編(`quest.go`, `endings.go`, `item_effects.go`, `odyssey.go`)

## 9-1 `odyssey.go`(46行)— 仲間(crew)と特殊アイテム

一番短いので先に読みます。オデュッセイア編に固有の仕掛けです。

```go
package main

const (
	itemLotusFruit    = "item.lotus_fruit"
	itemSacredCattle  = "item.sacred_cattle"
	itemBagOfWinds    = "item.bag_of_winds"
	ithacaShoreRoomID = "loc.ody_ithaca_shore"

	windsCrewLoss = 3
)
```

- 特別な効果を持つアイテムと部屋のIDを、定数にしている(文字列を直接書くとタイプミスに気づきにくいため)。
  - ロトスの実、聖なる牛、風の革袋、イタケー海岸。
- `windsCrewLoss`:風の革袋を早く開けたとき減る仲間の人数(3人)。

```go
func (s *Server) initializeCrewLocked(player *Player, fromRoomID string) {
	if player.RoomID != odysseyStartRoomID || s.world == nil || fromRoomID != s.world.StartRoomID {
		return
	}
	player.Crew = startingCrew
	player.CrewInitialized = true
}
```

- 「**運命の間(ハブ)からトロイ海岸に入った瞬間**」に、仲間を12人にする。
- 条件のどれかが外れていれば何もしない(早期 `return`)。
  - 今いる部屋がオデュッセイア編の入口ではない。
  - 出発した部屋が開始部屋ではない(途中から入り直した場合などは、リセットされない)。
- `||` で3つの「やらない条件」を並べ、**当てはまったら即終了**する書き方。

```go
func spendCrewLocked(player *Player, amount int) int {
	if amount > player.Crew {
		amount = player.Crew
	}
	player.Crew -= amount
	return amount
}
```

- 仲間を `amount` 人減らす。 **人数が足りなければ、いる分だけ減らす**(マイナスにならない)。
- **実際に減った人数**を返す(実況文に「仲間を◯人失った」と入れるため)。

```go
func (s *Server) applyTakeConsequencesLocked(player *Player, name, itemID string) *flavor {
	switch itemID {
	case itemLotusFruit:
		s.respawnPlayerLocked(player, name, "lotus", itemLotusFruit)
		return &flavor{key: "lotus", player: name}
	case itemSacredCattle:
		s.respawnPlayerLocked(player, name, "cattle", itemSacredCattle)
		return &flavor{key: "cattle", player: name}
	}
	return nil
}
```

- **拾った結果の効果**:ロトスの実(故郷を忘れる)や聖なる牛(ヘリオスの怒り)は、**手に取った瞬間に死亡**する(神話の通り)。
- `respawnPlayerLocked(player, name, "lotus", itemLotusFruit)`:死因の種類 `"lotus"` と、**死因になったアイテムのID**(`itemLotusFruit`)を渡す(8-4)。IDは、あとでモイライがヒントを返す手がかりになる。 `args` は省略(可変長引数なので、渡さなくてよい)。
- `switch itemID { case 定数: }`:値で分岐。どれでもなければ最後の `return nil`(何も起きない)。

```go
func (s *Server) applyDropConsequencesLocked(player *Player, name, itemID string) *flavor {
	if itemID != itemBagOfWinds || player.RoomID == ithacaShoreRoomID {
		return nil
	}
	lost := spendCrewLocked(player, windsCrewLoss)
	return &flavor{key: "winds_opened", player: name, n: lost}
}
```

- **手放した結果の効果**:風の革袋を、 **イタケー海岸以外で**手放すと嵐が起きて仲間が3人減る。
- 革袋以外、またはイタケー海岸なら何も起きない(`nil`)。

## 9-2 `quest.go`(239行)— クエスト

### 全体の考え方

- 「**完了報告コマンドは無い**」。サーバーが、アイテムを拾ったとき・敵を倒したときに**自動で判定**して進行させる。
- クエストの状態は、プレイヤーごとに `Player.Quests[クエストID]` に持つ(`active` / `completed` と進捗)。

### 9-2-1 import

```go
import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
)
```

### 9-2-2 `checkQuestObjectiveLocked`— 自動判定の入口

```go
func (s *Server) checkQuestObjectiveLocked(player *Player, objType, targetID string) {
	if s.world == nil {
		return
	}
	for questID, quest := range s.world.Quests {
		if quest == nil || quest.Objective.Type != objType || quest.Objective.TargetID != targetID {
			continue
		}
		state := player.Quests[questID]
		if state == nil || state.Status != "active" {
			continue
		}
		s.advanceQuestLocked(player, questID, quest, state)
	}
}
```

- 呼ばれる場面:
  - `TAKE` でアイテムを拾ったとき → `("collect_item", アイテムID)`
  - 敵を倒したとき → `("defeat_npc", 敵のID)`
- 全クエストを見て、 **目的の種類と対象が一致する**ものだけを調べる(違えば `continue`)。
- さらに、その人が**そのクエストを受けていて、進行中(`active`)**であれば進行させる。まだ受けていない、すでに達成済みなら無視。

### 9-2-3 `advanceQuestLocked`— 進捗を進める

```go
func (s *Server) advanceQuestLocked(player *Player, questID string, quest *Quest, state *PlayerQuest) {
	locale := s.localeOfLocked(player.Name)
	state.Progress++
```

- `state.Progress++`:進捗を1増やす。 `++` は「1足す」。

```go
	if state.Progress < quest.Objective.Count {
		logger.Info("quest_progress", "player", player.Name, "quest", questID, "progress", state.Progress, "target", quest.Objective.Count)
	}
	if state.Progress < quest.Objective.Count {
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "Quest \"%s\" progress: %d/%d.",
			"ja": "クエスト「%s」の進捗: %d/%d。",
		}.Format(locale, quest.Name.Get(locale), state.Progress, quest.Objective.Count))
		return
	}
```

- まだ目標数に達していなければ、ログに残し、 **「進捗: 2/3」と本人に通知**して終わる。
- `\"`:文字列の中のダブルクォートを書くための書き方。

```go
	state.Status = "completed"
	logger.Info("quest_completed", "player", player.Name, "quest", questID, "max_hp_gain", questMaxHPGain(quest))
	gain := questMaxHPGain(quest)
	player.MaxHPBonus += gain
	player.HP = s.maxHPLocked(player) // finishing a quest also restores full health
	s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{ ... }.Format(locale, quest.Name.Get(locale), gain, s.maxHPLocked(player)))
}
```

- 達したら `"completed"` にして、 **最大HPを上げる**。
  - `questMaxHPGain(quest)` は、報酬の値の **1/5**(最低1)です(9-4)。たとえば報酬が15なら +3 です。
  - `player.MaxHPBonus += gain`:上がった分を**プレイヤーに記録**します(第3章)。死んでも失わず、セーブにも残ります。
  - `player.HP = s.maxHPLocked(player)`:**HPを新しい最大HPまで全回復**します。以前は「報酬のHPを足す(100まで)」でしたが、報酬の意味が「回復」から「最大HPの成長」に変わりました。
- 「クエスト達成! 報酬: 最大HP+3(現在の最大HP 103)、HPも全回復」と本人に通知。
- ログの項目名も `reward_hp` から **`max_hp_gain`** に変わりました(READMEのログの表と対応)。

### 9-2-4 依頼者の案内

```go
func (s *Server) questGiverNPCIDsLocked(roomID string) []string {
	var ids []string
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID {
			continue
		}
		if _, quest := s.world.questByGiver(id); quest != nil {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
```

- ある部屋にいる「**クエストの依頼者になっているNPC**」のID一覧(並べ替え済み)。

```go
func (s *Server) announceQuestGiversLocked(player *Player) {
	if s.world == nil {
		return
	}
	locale := s.localeOfLocked(player.Name)
	for _, npcID := range s.questGiverNPCIDsLocked(player.RoomID) {
		questID, quest := s.world.questByGiver(npcID)
		if player.Quests[questID] != nil {
			continue
		}
		npcName := s.world.NPCs[npcID].Name.Get(locale)
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "%s has a request for you: \"%s\". Type QUEST %s to hear it and accept.",
			"ja": "%sから依頼がある: 「%s」。QUEST %s と入力すると内容を聞いて受注できる。",
		}.Format(locale, npcName, quest.Name.Get(locale), npcName))
	}
}
```

- 部屋に入った(`CONNECT` / `MOVE`)とき、 **まだ受けていない依頼があれば案内**する。
- `player.Quests[questID] != nil`:すでに受けている(または達成済み)なら案内しない。

```go
func (s *Server) sendQuestHintLocked(player *Player, npcID string) {
	questID, quest := s.world.questByGiver(npcID)
	if quest == nil {
		return
	}
	...
	state := player.Quests[questID]
	switch {
	case state == nil:
		(「◯◯を頼みたい(報酬: 最大HP+◯)。QUEST でうけられる」を通知)
	case state.Status == "active":
		(「クエスト◯◯は進行中(2/3)」を通知)
	}
}
```

- `TALK` で依頼者と話したときの通知。
  - まだ受けていない → 依頼の案内。
  - 進行中 → 進捗を知らせる。
  - 達成済みなら、どのcaseにも当たらず何も出ない。

### 9-2-5 `handleQuest`— 依頼を受ける

```go
	npcID := s.world.resolveNPCInRoom(player.RoomID, query, locale)
	if npcID == "" { ...404 NPC_NOT_FOUND... }
	questID, quest := s.world.questByGiver(npcID)
	if quest == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 406 NO_QUEST_AVAILABLE")
		return false
	}
	if state := player.Quests[questID]; state != nil && state.Status == "completed" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 406 NO_QUEST_AVAILABLE")
		return false
	}
```

- 同室のNPCを特定し、そのNPCが出すクエストを探す。
- 無い、または**もう達成済み**なら `406 NO_QUEST_AVAILABLE`。

```go
	if player.Quests == nil {
		player.Quests = make(map[string]*PlayerQuest)
	}
	newlyAccepted := player.Quests[questID] == nil
	if newlyAccepted {
		player.Quests[questID] = &PlayerQuest{Status: "active"}
		logger.Info("quest_accepted", "player", *name, "quest", questID, "giver", quest.GiverNPCID)
	}
```

- 辞書が `nil` なら作る。
- **初めて受けるなら**、 `active` の状態を作る。すでに受けている(進行中の)場合は何も変えず、説明だけ繰り返す。
- `newlyAccepted`:新規に受けたかの `bool`。

```go
	data, err := json.Marshal(struct {
		QuestID     string `json:"quest_id"`
		Description string `json:"description"`
		Reward      int    `json:"reward"`
		Status      string `json:"status"`
	}{questID, quest.Description.Get(locale), questMaxHPGain(quest), "available"})
	...
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil && newlyAccepted && s.objectiveAlreadyMetLocked(player, quest) {

		s.advanceQuestLocked(player, questID, quest, player.Quests[questID])
	}
```

- 応答はRFCで決まった形 `{quest_id, description, reward, status}`。
  - ただし**`reward` は数値**で、中身は「最大HPの増える量」(`questMaxHPGain`)です。RFCの例では `"gold_coin"` のような文字列なので、**RFCとの違い**になります(READMEの「Protocol Implementation」の表に理由を書いてあります)。
- **受けた時点で条件をすでに満たしていた**(たとえばもう持っているアイテム)ら、その場で進行させる。

### 9-2-6 `handleQuests`— 一覧

```go
	ids := make([]string, 0, len(player.Quests))
	for id := range player.Quests {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	type questEntry struct {
		QuestID  string `json:"quest_id"`
		Status   string `json:"status"`
		Progress string `json:"progress"`
	}
```

- 受けたクエストのIDを集めて並べ替える。
- **関数の中で型を定義**している(`type questEntry struct`)。この関数でしか使わない型は、近くに置ける。

```go
	list := make([]questEntry, 0, len(ids))
	for _, id := range ids {
		state := player.Quests[id]
		if state == nil {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
			return false
		}
		count := 1
		if quest := s.world.Quests[id]; quest != nil {
			count = quest.Objective.Count
		}
		list = append(list, questEntry{id, state.Status, fmt.Sprintf("%d/%d", state.Progress, count)})
	}
```

- 各クエストについて `{quest_id, status, progress: "2/3"}` を作る。
- 状態が`nil`なら、共通ロックを解放して`ERR 500 STATE_ERROR`を返します。他のプレイヤーのコマンド処理を止めません。
- `fmt.Sprintf("%d/%d", 進捗, 目標)` で `"2/3"` の文字列に。
- 応答は `OK [{...},{...}]`。

### 9-2-7 `objectiveAlreadyMetLocked`

```go
func (s *Server) objectiveAlreadyMetLocked(player *Player, quest *Quest) bool {
	switch quest.Objective.Type {
	case "collect_item":
		return player.hasItem(quest.Objective.TargetID)
	case "defeat_npc":
		npc := s.world.NPCs[quest.Objective.TargetID]
		return npc != nil && player.enemyHP(quest.Objective.TargetID, npc) <= 0
	}
	return false
}
```

- 「受けた時点で、すでに目的を満たしているか」。
  - アイテム集め:もう持っている。
  - 敵を倒す:その敵がこの人にとってすでに倒されている(HPが0以下)。

## 9-3 `endings.go`(242行)— エンディングと祝福

### 9-3-1 `Ending` の型

```go
type Ending struct {
	ID              string          `json:"id"`
	Name            LocalizedText   `json:"name"`
	RequiresItems   []string        `json:"requires_items,omitempty"`
	RequiresQuests  []string        `json:"requires_quests,omitempty"`
	RequiresEndings []string        `json:"requires_endings,omitempty"`
	RewardItem      string          `json:"reward_item,omitempty"`
	Blessing        *Blessing       `json:"blessing,omitempty"`
	Hint            LocalizedText   `json:"hint"`
	Text            []LocalizedText `json:"text"`
}
```

- 各NPCに付けられるエンディングの定義。
- `RequiresItems` / `RequiresQuests` / `RequiresEndings`:迎えるために**持っているべきアイテム・達成すべきクエスト・先に迎えるべき別のエンディング**(最終エンディングは、3つの物語のエンディングを全部迎えることが条件)。
- `RewardItem`:報酬の記念品(アイテムID)。
- `Blessing *Blessing`:このエンディングを迎えたときに**永続で手に入る神の祝福**。ポインタなので、祝福の無いエンディング(最終エンディング)は `nil`。下の「祝福」を参照。
- `Hint`:条件が足りないときに表示する文。
- `Text []LocalizedText`:エンディングの本文(複数行)。

### 祝福 `Blessing` と効果の種類

```go
// Blessing is a permanent boon from the god of an arc, earned by reaching that arc's ending.
// It is derived from the endings a player has reached, so nothing extra is stored on the player.
type Blessing struct {
	God         LocalizedText `json:"god"`
	Name        LocalizedText `json:"name"`
	Description LocalizedText `json:"description"`
	Effect      string        `json:"effect"`
	Value       int           `json:"value"`
}

const (
	blessingCounterReduction = "counter_reduction" // enemy counter-attacks hurt Value% less
	blessingRegenBonus       = "regen_bonus"       // Value extra HP every regen tick
	blessingDamageBonus      = "damage_bonus"      // Value extra damage on every hit
	maxCounterIncrease       = 50                  // bad items never make counter-attacks hurt more than 50% extra
	maxCounterReduction      = 80                  // allies (3 x 20%) plus a blessing never reach 100%
)
```

- **祝福**は「各編の神が、その編をクリアした人にくれる、一生ものの効果」です。アテナ(オデュッセイア編)・ヘラ(アルゴ船編)・アポロン(トロイア編)の3つがあります。
- `God`:神の名前、`Name`:祝福の名前、`Description`:効果の説明(どれも言語別)。
- `Effect`:**効果の種類**を表す文字列で、下の3つの定数のどれかです。`Value` はその大きさです。
  - `counter_reduction`:敵の反撃のダメージを `Value`% **減らす**(アテナ:20)。
  - `regen_bonus`:HP自動回復の1回あたりを `Value` **増やす**(ヘラ:1。つまり回復が2倍)。
  - `damage_bonus`:自分の攻撃のダメージを `Value` **増やす**(アポロン:3)。
- 祝福は**プレイヤーに保存しません**。コメントのとおり、「どのエンディングを迎えたか」(`Endings`)から**毎回計算して求める**ので、持ち物のように失う心配も、データが食い違う心配もありません(下の `blessingTotalLocked`)。
- `maxCounterReduction = 80`:仲間3人(20%×3=60%)に祝福(20%)を足しても、**軽減の合計は80%まで**。反撃が0になって無敵にならないための上限です(8-3の `reduceCounter` が使います)。
- `maxCounterIncrease = 50`:逆に、悪いアイテムで反撃が強くなる分は**最大50%まで**。理不尽に死なないための下限です。
- 祝福の効果の種類(`blessing...` の3つ)は、アイテムの効果(9-4)も**同じ名前を共有**しています。そのため、反撃や与ダメージの計算では、祝福とアイテムの効果を**同じ足し算**にまとめられます。

### 9-3-2 `validateEndings`

`world.go` の `validate()` から最後に呼ばれる検査です。

```go
func (w *World) validateEndings() error {
	ids := make(map[string]string)
	for npcID, npc := range w.NPCs {
		if npc == nil || npc.Ending == nil {
			continue
		}
		e := npc.Ending
		if e.ID == "" {
			return fmt.Errorf("NPC %q has an ending with no id", npcID)
		}
		if other, dup := ids[e.ID]; dup {
			return fmt.Errorf("ending %q is defined by both %q and %q", e.ID, other, npcID)
		}
		ids[e.ID] = npcID
```

- エンディングを持つNPCだけ調べる。
- `ids`:エンディングID→NPCのID。 **同じエンディングIDが2回出たら、重複エラー**。

```go
		if len(e.Text) == 0 { ...本文が無い... }
		for _, itemID := range e.RequiresItems {
			if w.Items[itemID] == nil { ...存在しないアイテム... }
		}
		for _, questID := range e.RequiresQuests {
			if w.Quests[questID] == nil { ...存在しないクエスト... }
		}
		if b := e.Blessing; b != nil {
			switch b.Effect {
			case blessingCounterReduction, blessingRegenBonus, blessingDamageBonus:
			default:
				return fmt.Errorf("ending %q blessing has unknown effect %q", e.ID, b.Effect)
			}
			if b.Value < 1 || b.God["en"] == "" || b.Name["en"] == "" || b.Name["ja"] == "" || b.God["ja"] == "" {
				return fmt.Errorf("ending %q blessing needs a positive value and en/ja god and name", e.ID)
			}
		}
		if e.RewardItem != "" {
			item := w.Items[e.RewardItem]
			if item == nil || !item.RewardOnly {
				return fmt.Errorf("ending %q reward %q must be an existing reward_only item", e.ID, e.RewardItem)
			}
		}
	}
```

- 本文があるか、必要なアイテム・クエストが実在するか、 **祝福の定義が正しいか**、**報酬が「報酬専用アイテム」であるか**を検査。
- 祝福の検査:
  - `case a, b, c:`:`switch` の1つの `case` に**値を複数並べる**と「どれかに当てはまれば」の意味になります。3つの既知の効果のどれかなら何もせず(空の `case`)通り、それ以外は `default` でエラーにします。
  - `b.Value < 1`:効果の大きさは**1以上**でなければなりません(0やマイナスの祝福は意味が無い)。
  - `b.God["en"] == ""` など:**神と祝福の名前に、英語と日本語の両方**が必要です。表示の言語を切り替えたときに、名前が空にならないための検査です。

```go
	for _, npcID := range ids {
		for _, required := range w.NPCs[npcID].Ending.RequiresEndings {
			if _, ok := ids[required]; !ok {
				return fmt.Errorf("ending %q requires unknown ending %q", w.NPCs[npcID].Ending.ID, required)
			}
		}
	}
	return nil
}
```

- 全エンディングのIDが出そろってから、「**先に迎えるべきエンディング**」が実在するかを検査(1周目では、まだ出ていないIDを指す可能性があるため、2周目に分けている)。

### 9-3-3 検索の補助

```go
func (w *World) routeEndings() []*Ending {
	var list []*Ending
	for _, npc := range w.NPCs {
		if npc != nil && npc.Ending != nil && len(npc.Ending.RequiresEndings) == 0 {
			list = append(list, npc.Ending)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}
```

- 「**他のエンディングを前提としない**エンディング」=各物語の個別エンディングの一覧(最終エンディングは除く)。
- `sort.Slice(スライス, func(i, j int) bool {...})`:**比べ方を関数で指定して並べ替える**。 `i番目 < j番目` なら `true` を返す関数を渡す。ここではID順。

```go
func (w *World) endingByID(id string) *Ending {
	for _, npc := range w.NPCs {
		if npc != nil && npc.Ending != nil && npc.Ending.ID == id {
			return npc.Ending
		}
	}
	return nil
}
```

- IDからエンディングを探す。無ければ `nil`。

### 9-3-4 `missingForEndingLocked`— 足りない条件

```go
func (s *Server) missingForEndingLocked(player *Player, e *Ending, locale string) []string {
	var missing []string
	for _, itemID := range e.RequiresItems {
		if !player.hasItem(itemID) {
			missing = append(missing, LocalizedText{
				"en": "\"%s\" in your inventory",
				"ja": "所持品の「%s」",
			}.Format(locale, s.world.Items[itemID].Name.Get(locale)))
		}
	}
	for _, questID := range e.RequiresQuests {
		if state := player.Quests[questID]; state == nil || state.Status != "completed" {
			missing = append(missing, LocalizedText{
				"en": "the completed quest \"%s\"",
				"ja": "クエスト「%s」の達成",
			}.Format(locale, s.world.Quests[questID].Name.Get(locale)))
		}
	}
	for _, endingID := range e.RequiresEndings {
		if !player.Endings[endingID] {
			missing = append(missing, LocalizedText{
				"en": "the ending \"%s\"",
				"ja": "エンディング「%s」への到達",
			}.Format(locale, s.world.endingByID(endingID).Name.Get(locale)))
		}
	}
	return missing
}
```

- エンディングを迎えるのに**足りないもの**を、文章のリストにして返す。 `len(missing) == 0` なら条件を満たしている。
- 必要なアイテムの所持、クエストの`completed`、前提エンディングへの到達を順に調べる。クエストの記録が`nil`でも、`state == nil`の判定で未達成として扱う。
- (現在のゲームでは、足りない内容を**プレイヤーに教えず**、ただヒント文だけを出す仕様。リストの長さだけを判定に使っている)

### 祝福の合計 `blessingTotalLocked`

```go
// blessingTotalLocked adds up one effect over every blessing the player has earned.
func (s *Server) blessingTotalLocked(player *Player, effect string) int {
	if s.world == nil {
		return 0
	}
	total := 0
	for _, npc := range s.world.NPCs {
		if npc == nil || npc.Ending == nil || npc.Ending.Blessing == nil {
			continue
		}
		if b := npc.Ending.Blessing; b.Effect == effect && player.Endings[npc.Ending.ID] {
			total += b.Value
		}
	}
	return total
}
```

- **「ある種類の効果を、このプレイヤーが祝福から合計いくつ受けているか」**を返します。`effect` に `"damage_bonus"` を渡せば与ダメージの加算、`"counter_reduction"` なら反撃の軽減率が得られます。
- 世界の全NPCを調べ、エンディングと祝福を持つNPCのうち、
  - **その効果の種類が `effect` と同じで**(`b.Effect == effect`)、
  - **このプレイヤーがそのエンディングを迎えている**(`player.Endings[...]`)
  
  ものの `Value` を足し合わせます。
- 3つの編をすべてクリアしていれば、3つの祝福が全部効きます(種類が違うので、足し合わさずそれぞれ別の効果になります)。
- この関数は祝福だけの合計です。戦闘や回復の計算は、アイテム効果も足した **`effectTotalLocked`(9-4)** を呼びます。

### 9-3-5 通知の補助

```go
func (s *Server) sendEndingLocked(name, text string) {
	s.sendPlayerEventLocked(name, "ENDING", text)
}
```

- エンディング用の通知 `EVT PLAYER ENDING 文` を送る短い関数。

`sendEndingProgressLocked`:

```go
	routes := s.world.routeEndings()
	if len(routes) == 0 {
		return
	}
	locale := s.localeOfLocked(player.Name)
	done := 0
	parts := make([]string, 0, len(routes))
	for _, e := range routes {
		state := LocalizedText{"en": "not yet", "ja": "未達成"}.Get(locale)
		if player.Endings[e.ID] {
			done++
			state = LocalizedText{"en": "done", "ja": "達成"}.Get(locale)
		}
		parts = append(parts, e.Name.Get(locale)+" ("+state+")")
	}
	separator := ", "
	if locale == "ja" {
		separator = "、"
	}
	s.sendEndingLocked(player.Name, LocalizedText{
		"en": "Endings reached: %d/%d. %s",
		"ja": "到達したエンディング: %d/%d。%s",
	}.Format(locale, done, len(routes), strings.Join(parts, separator)))
```

- 「到達したエンディング: 1/3。◯◯(達成)、△△(未達成)、…」という**進捗の通知**を作る。
- `strings.Join(parts, separator)`:文字列のリストを区切り文字でつなぐ。区切りは日本語なら「、」、英語なら「, 」。

```go
func (s *Server) grantEndingRewardLocked(player *Player, e *Ending) {
	if e.RewardItem != "" && !player.hasItem(e.RewardItem) {
		player.Inventory = append(player.Inventory, e.RewardItem)
	}
}
```

- 報酬の記念品を、持っていなければ持ち物に加える(二重には渡さない)。

### 9-3-6 `talkEndingLocked`— 判定の本体

`TALK` のあと、そのNPCがエンディングを持っていれば呼ばれる。

```go
func (s *Server) talkEndingLocked(player *Player, npc *NPC) {
	e := npc.Ending
	if e == nil || s.world == nil {
		return
	}
	locale := s.localeOfLocked(player.Name)
	play := func() {
		for _, line := range e.Text {
			s.sendEndingLocked(player.Name, line.Get(locale))
		}
	}
```

- エンディングが無いNPCは即終了。
- **`play := func() { ... }`**:関数を変数に入れる。この関数は**外側の変数(`e`、`locale`、`player`、`s`)をそのまま使える**。このような「外側の変数を覚えて使う関数」を**クロージャ**と呼ぶ。本文を1行ずつ通知する関数。何度か呼ぶので、関数にまとめている。

```go
	if player.Endings[e.ID] {
		s.grantEndingRewardLocked(player, e)
		play()
		return
	}
```

- **すでに迎えたエンディング**:報酬を(まだなら)渡し、本文をもう一度流す。

```go
	if missing := s.missingForEndingLocked(player, e, locale); len(missing) > 0 {

		s.sendEndingLocked(player.Name, e.Hint.Get(locale))
		return
	}
```

- **条件が足りない**:ヒント文だけを送って終了(足りないものは教えない)。

```go
	if player.Endings == nil {
		player.Endings = make(map[string]bool)
	}
	player.Endings[e.ID] = true
	s.grantEndingRewardLocked(player, e)
	logger.Info("ending_reached", "player", player.Name, "ending", e.ID)
	s.sendEndingLocked(player.Name, LocalizedText{
		"en": "=== ENDING: %s ===",
		"ja": "=== エンディング: %s ===",
	}.Format(locale, e.Name.Get(locale)))
	play()
	if reward := s.world.Items[e.RewardItem]; reward != nil {
		s.sendEndingLocked(player.Name, LocalizedText{ ...「報酬を受け取った: ◯◯」... }.Format(locale, reward.Name.Get(locale)))
	}
	if b := e.Blessing; b != nil {
		s.sendEndingLocked(player.Name, LocalizedText{
			"en": "%s grants you her blessing, for good: %s. %s",
			"ja": "%sの祝福を受けた(永続): %s。%s",
		}.Format(locale, b.God.Get(locale), b.Name.Get(locale), b.Description.Get(locale)))
	}
	s.sendEndingProgressLocked(player)
}
```

- **初めて迎える**:記録して報酬を付与し、ログに残す。
- 「=== エンディング: ◯◯ ===」の見出し → 本文 → 報酬の案内 → **祝福の案内** → 全体の進捗、の順に通知する。
- `if b := e.Blessing; b != nil {`:祝福があるエンディングだけ、「◯◯の祝福を受けた(永続): △△。(効果の説明)」と知らせます。祝福そのものは**ここで何かを保存するのではなく**、`player.Endings[e.ID] = true` と記録した時点で、`blessingTotalLocked` が自動的に数えるようになります。
- 3つの編をすべて迎えた人がモイライに話しかけると、最終エンディング(`ending.final`)が流れ、これを迎えると**運命の間の南に隠し出口が開きます**(`room.go` の `exitsFor`、2-6)。

---

## 9-4 `item_effects.go`(77行)— アイテムの効果と最大HP

**役割**:持っているアイテムが与える効果(良いものも悪いものも)を集計し、**今の最大HP**を計算します。祝福(9-3)とは別の仕組みですが、足し算で1つにまとめて使います。

### 9-4-1 定数と型

```go
const (
	effectMaxHP = "max_hp" // Value more (or fewer) max HP while carried

	// A quest's reward HP is turned into a permanent max HP gain: reward / questMaxHPDivisor (at least 1).
	questMaxHPDivisor = 5
	// minPlayerMaxHP keeps a bad item from pushing max HP under the HP a player respawns with.
	minPlayerMaxHP = 20
)

type ItemEffect struct {
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}
```

- `effectMaxHP`:アイテムだけが持つ4つめの効果の種類(**最大HPの増減**)です。祝福の3種類(9-3)と合わせて、アイテムの効果は全部で4種類あります。
- `questMaxHPDivisor = 5`:クエストの報酬を**5で割った値**が、最大HPの上昇量です(報酬が10〜25なので、上昇は2〜5)。
- `minPlayerMaxHP = 20`:最大HPの**下限**。コメントのとおり、復活したときのHP(20)より小さくならないようにします(悪いアイテムを大量に持っても、復活した瞬間に最大HPを超えてしまうことがない)。
- `ItemEffect`:効果1つ分(種類と大きさ)。`world.json` の `"effects": [ {"effect": "damage_bonus", "value": 3} ]` の1項目が、これ1つに読み込まれます(5-2)。**`Value` はマイナスもあり**、マイナスなら悪い効果です。

### 9-4-2 検査

```go
func validItemEffect(effect string) bool {
	switch effect {
	case effectMaxHP, blessingCounterReduction, blessingRegenBonus, blessingDamageBonus:
		return true
	}
	return false
}

func (w *World) validateItemEffects() error {
	for id, item := range w.Items {
		if item == nil {
			continue
		}
		for _, e := range item.Effects {
			if !validItemEffect(e.Effect) {
				return fmt.Errorf("item %q has unknown effect %q", id, e.Effect)
			}
			if e.Value == 0 {
				return fmt.Errorf("item %q effect %q has value 0", id, e.Effect)
			}
		}
	}
	return nil
}
```

- `validItemEffect`:**知っている効果の種類か**を返す小さな関数。4つのどれかなら `true`。
- `validateItemEffects`:**全アイテムの効果を検査**します。`world.go` の `validate`(5-7)から呼ばれます。
  - 知らない種類(`"luck"` などのタイプミス)は、起動の時点でエラー。
  - 大きさが **0 の効果**もエラー(効果が無いので、書き間違いだと分かる)。
  - 祝福と違い、**マイナスは許します**(悪い効果のため)。

### 9-4-3 持ち物からの集計

```go
func (s *Server) itemEffectTotalLocked(player *Player, effect string) int {
	if s.world == nil {
		return 0
	}
	total := 0
	for _, itemID := range player.Inventory {
		if item := s.world.Items[itemID]; item != nil {
			for _, e := range item.Effects {
				if e.Effect == effect {
					total += e.Value
				}
			}
		}
	}
	return total
}
```

- **持っているアイテム全部**(`player.Inventory`)について、効果を1つずつ調べ、**種類が `effect` と同じものの値を足し合わせます**。
- 見るのは**持ち物だけ**です。部屋に落ちているアイテムの効果は、効きません(拾うと効く・置くと消える)。これが「持っている間だけ効く」の実体です。
- 良い効果(+)と悪い効果(-)は、**そのまま足し引き**されます。竜の歯(与ダメージ+3、被ダメージ+10%)と眠り薬(与ダメージ-2)を両方持てば、与ダメージの合計は `+3 + (-2) = +1` です。
- `if item := s.world.Items[itemID]; item != nil {`:辞書から引いた結果が `nil` でないときだけ使う、いつもの形です(持ち物のIDが世界に無いアイテムでも落ちません)。

### 9-4-4 祝福との合算と最大HP

```go
// effectTotalLocked is the player's total for one effect: god blessings plus carried items.
func (s *Server) effectTotalLocked(player *Player, effect string) int {
	return s.blessingTotalLocked(player, effect) + s.itemEffectTotalLocked(player, effect)
}

// maxHPLocked is the player's current max HP: the base, quest gains (kept after death) and carried items.
func (s *Server) maxHPLocked(player *Player) int {
	return max(minPlayerMaxHP, maxPlayerHP+player.MaxHPBonus+s.itemEffectTotalLocked(player, effectMaxHP))
}

// questMaxHPGain is how much finishing the quest gives.
func questMaxHPGain(quest *Quest) int {
	return max(1, quest.Reward.HP/questMaxHPDivisor)
}
```

- `effectTotalLocked`:**祝福の合計 + アイテムの合計**。戦闘(8-5、8-6)と回復(7-7)は、祝福だけでなくこの関数を呼びます。同じ名前(`damage_bonus` など)の効果を**1つの足し算に統一**してあるので、計算の側は「祝福かアイテムか」を気にしません。
- `maxHPLocked`:**今の最大HP**。次の3つを足して、**下限(20)より小さくならない**ようにします。
  - `maxPlayerHP`(100):基本の最大HP。
  - `player.MaxHPBonus`:クエストで増えた分(死んでも失わない)。
  - `itemEffectTotalLocked(player, effectMaxHP)`:持っているアイテムの最大HPの増減(置くと元に戻る)。
- `questMaxHPGain`:クエストを達成したときの**最大HPの上昇量**。`quest.Reward.HP / 5` ですが、整数の割り算で端数が切り捨てになるので、`max(1, ...)` で**最低1**にします(報酬が3でも、1は上がる)。

> **使われている場所のまとめ**:`maxHPLocked` は `STATUS`(7-20)、HP自動回復(7-7、3-8)、クエスト達成(9-2)から呼ばれます。`effectTotalLocked` は `ATTACK`(8-5)・`FLEE`(8-6)と、HP自動回復(7-7)から呼ばれます。

# 第10章 チャットとグループ(`chat.go`, `group.go`)

## 10-1 `chat.go`(99行)

```go
import (
	"fmt"
	"net"
	"strings"
	"unicode"
	"unicode/utf8"
)
```

### 10-1-1 `parseCommandParts`— 1行を単語に分ける

7-21の `handleClient` が、 **すべての行**に対して最初に呼ぶ関数です。

```go
func parseCommandParts(line string) []string {
	parts := strings.Fields(line)
	if len(parts) < 3 || !strings.EqualFold(parts[0], "CHAT") {
		return parts
	}
```

- `strings.Fields(line)`:**空白(連続した空白やタブも)で区切って**単語のスライスにする。
- 普通のコマンドは、これだけで終わり。
- `CHAT スコープ 本文` で単語が3つ以上のときだけ、特別な処理に進む(本文の空白をそのまま保つため)。

```go
	rest := strings.TrimLeftFunc(line, unicode.IsSpace)
	for range 2 {
		separator := strings.IndexFunc(rest, unicode.IsSpace)
		if separator < 0 {
			return parts
		}
		rest = strings.TrimLeftFunc(rest[separator:], unicode.IsSpace)
	}
	return []string{parts[0], parts[1], strings.TrimSuffix(rest, "\r")}
}
```

- `strings.TrimLeftFunc(line, unicode.IsSpace)`:先頭の空白を取り除く。
- **`for range 2 {`**:2回繰り返す(Go 1.22以降の書き方。回数だけ指定する)。
  - 1回ごとに「次の空白の位置」を探し、そこまでの単語(`CHAT`、`GLOBAL`)を読み飛ばす。
  - `rest[separator:]`:文字列の切り出し(`separator`の位置から最後まで)。
- 2回済んだあとの `rest` が**本文(元の空白を含む)**。
- `strings.TrimSuffix(rest, "\r")`:末尾の `\r`(Windowsの改行の一部)を取る。
- 3つの要素 `[CHAT, スコープ, 本文]` のスライスを返す。 `[]string{a, b, c}` はスライスのリテラル。

### 10-1-2 `handleChat`

```go
	scope := strings.ToUpper(parts[1])
	if scope != "GLOBAL" && scope != "ROOM" && scope != "GROUP" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	message := strings.Join(parts[2:], " ")
	if strings.TrimSpace(message) == "" || !utf8.ValidString(message) || strings.IndexFunc(message, unicode.IsControl) >= 0 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	event := "EVT " + scope + " CHAT " + *name + " " + message
	if len(event) > maxProtocolLineBytes {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
```

- スコープ(範囲)は `GLOBAL` / `ROOM` / `GROUP` のみ。
- 本文は、空白だけではなく、正しいUTF-8で、制御文字を含まない(改行を混ぜて偽の通知行を作らせない)。
- 送る通知 `EVT スコープ CHAT 名前 本文` が1行の上限を超えないか。

```go
	client := conn.(*serverClient)
	s.mu.Lock()
	player := s.players[*name]
	...
	var group *Group
	if scope == "GROUP" {
		id, member := s.groupByPlayer[*name]
		if !member {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 401 NOT_IN_GROUP")
			return false
		}
		group = s.groups[id]
		if group == nil { ...500... }
	}
```

- `GROUP` 宛てのときだけ、 **グループに入っているか**を確認(入っていなければ `401`)。

```go
	response, err := client.enqueueResponse("OK")
	if err == nil {
		switch scope {
		case "GLOBAL":
			for _, recipient := range s.clients {
				recipient.enqueueEvent(event)
			}
		case "ROOM":
			for name, other := range s.players {
				if other.RoomID == player.RoomID {
					if recipient := s.clients[name]; recipient != nil {
						recipient.enqueueEvent(event)
					}
				}
			}
		case "GROUP":
			s.broadcastGroupLocked(group, event)
		}
	}
```

- 先に `OK` を積み、成功したらスコープに応じて配る。
  - **GLOBAL**:接続中の全員。
  - **ROOM**:同じ部屋の全員。
  - **GROUP**:グループの全メンバー。
- **自分にも届く**(自分の発言も通知として戻ってくる)。
- 内側の `name` は、外側の `*name` とは別の変数(辞書の `range` のキー)。同じ名前を内側で作ると外側が**隠される**(シャドーイング)。この関数ではここ以降 `*name` を使っていないので問題ない。

## 10-2 `group.go`(223行)

### 10-2-1 グループの型

```go
type Group struct {
	ID      string
	Leader  string
	Members map[string]struct{}
	Invited map[string]struct{}
}

func newGroup(id, creator string) *Group {
	return &Group{
		ID:      id,
		Leader:  creator,
		Members: map[string]struct{}{creator: {}},
		Invited: make(map[string]struct{}),
	}
}
```

- `Members`、`Invited`:**集合**(`map[string]struct{}`)。名前を入れるだけ。
- `map[string]struct{}{creator: {}}`:最初のメンバー(作った人)だけが入った集合のリテラル。 `{}` は空の `struct{}` の値。
- JSONには出さない型(セーブされない)なので、 `json:` タグは無い。

### 10-2-2 `handleGroup`— 振り分け

```go
	if !requireArgs(conn, parts, 2) {
		return false
	}
	action := strings.ToUpper(parts[1])
	switch action {
	case "CREATE", "LEAVE":
		if !requireExactArgs(conn, parts, 2) {
			return false
		}
	case "INVITE", "JOIN":
		if !requireExactArgs(conn, parts, 3) {
			return false
		}
	default:
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
```

- `GROUP CREATE` のように、 **2番目の単語が操作の種類**。
- `case "CREATE", "LEAVE":`:1つのcaseに**複数の値**を書ける。
- 操作ごとに、単語の数をチェックする(`INVITE 名前` / `JOIN リーダー名` は3つ)。知らない操作は `400`。

```go
	switch action {
	case "CREATE":
		return handleGroupCreate(s, conn, *name)
	case "INVITE":
		return handleGroupInvite(s, conn, *name, parts[2])
	case "JOIN":
		return handleGroupJoin(s, conn, *name, parts[2])
	default:
		return handleGroupLeave(s, conn, *name)
	}
```

- 操作ごとの関数に振り分ける。

### 10-2-3 `handleGroupCreate`

```go
	if _, exists := s.groupByPlayer[name]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	id := fmt.Sprintf("group.%d", s.nextGroupID+1)
	response, err := client.enqueueResponse("OK group=" + id)
	if err == nil {
		s.nextGroupID++
		s.groups[id] = newGroup(id, name)
		s.groupByPlayer[name] = id
		s.clearGroupInvitesLocked(name)
	}
```

- すでにグループにいれば `402 ALREADY_IN_GROUP`。
- ID `group.1`、 `group.2`… を順に作る。
- **応答が積めたときだけ**グループを作る(`nextGroupID++` も、そのときだけ)。作ったあと、自分宛ての招待は消す(グループに入ったので)。

### 10-2-4 `handleGroupInvite`

```go
	id, member := s.groupByPlayer[name]
	group := s.groups[id]
	if !member { ...401 NOT_IN_GROUP... }
	if group == nil || s.players[name] == nil { ...500... }
	if _, exists := s.groupByPlayer[target]; exists { ...402 ALREADY_IN_GROUP... }
	invitee := s.clients[target]
	if s.players[target] == nil || invitee == nil { ...400 BAD_REQUEST... }
	response, err := client.enqueueResponse("OK")
	if err == nil {
		group.Invited[target] = struct{}{}
		invitee.enqueueEvent("EVT GROUP INVITE " + group.Leader)
	}
```

- 招待する人は**グループに入っている**必要がある(`401`)。
- 招待される相手が**すでに別のグループ**にいれば `402`。**接続していない**相手なら `400`。
- 招待を記録し、相手に `EVT GROUP INVITE リーダー名` を送る。

### 10-2-5 `handleGroupJoin`

```go
	id := s.groupByPlayer[leader]
	group := s.groups[id]
	if group == nil || group.Leader != leader { ...400... }
	if _, invited := group.Invited[name]; !invited { ...400... }
	response, err := client.enqueueResponse("OK group=" + id)
	if err == nil {
		group.Members[name] = struct{}{}
		s.groupByPlayer[name] = id
		s.clearGroupInvitesLocked(name)
		s.broadcastGroupLocked(group, "EVT GROUP JOIN "+name)
	}
```

- `JOIN リーダー名`:そのリーダーのグループがあり、 **自分が招待されている**ときだけ参加できる。
- 参加したら、メンバー全員に `EVT GROUP JOIN 名前`。

### 10-2-6 `handleGroupLeave` と補助

```go
func (s *Server) broadcastGroupLocked(group *Group, event string) {
	for member := range group.Members {
		if client := s.clients[member]; client != nil {
			client.enqueueEvent(event)
		}
	}
}

func (s *Server) clearGroupInvitesLocked(name string) {
	for _, group := range s.groups {
		delete(group.Invited, name)
	}
}
```

- `broadcastGroupLocked`:グループの全員に通知。
- `clearGroupInvitesLocked`:全グループから、その人への招待を消す。

```go
func (s *Server) removeGroupMemberLocked(name string) {
	id, member := s.groupByPlayer[name]
	if !member {
		return
	}
	delete(s.groupByPlayer, name)
	group := s.groups[id]
	if group == nil {
		return
	}
	delete(group.Members, name)
	if len(group.Members) == 0 {
		delete(s.groups, id)
		return
	}
	if group.Leader == name {
		group.Leader = ""
		for member := range group.Members {
			if group.Leader == "" || member < group.Leader {
				group.Leader = member
			}
		}
		clear(group.Invited)
	}
	s.broadcastGroupLocked(group, "EVT GROUP LEAVE "+name)
}
```

- グループを抜ける処理。 `LEAVE` と、切断・退室のとき(7-6)の両方から呼ばれる。
- グループに入っていなければ何もしない。
- 抜けて**メンバーが0人**になったら、グループごと削除。
- **リーダーが抜けた**ら、残りのメンバーのうち**名前の順で一番小さい人**を新リーダーに(辞書の順序は不定なので、決め方を固定している)。 `clear(辞書)` は辞書の中身を全部消す組み込み関数。このとき古い招待も無効にする。
- 残りのメンバーに `EVT GROUP LEAVE 名前` を通知。

---

# 第11章 通知文と実況文(`notify.go`, `flavor.go`)

プレイヤーに送る**物語の文章**を扱うファイルです。2種類あります。

| 種類 | 送り先 | 送る形式 | 例 |
|---|---|---|---|
| **個人向けの通知** | 本人だけ | `EVT PLAYER <種別> 文章` | 死亡、クエスト、案内、エンディング、仲間 |
| **実況(flavor)** | 同じ部屋の全員 | `EVT ROOM COMBAT 文章` | 「aliceはポリュペモスを打ち倒した」 |

## 11-1 `notify.go`(173行)— 個人向けの通知

### 11-1-1 `Format` と言語の取得

```go
func (t LocalizedText) Format(locale string, args ...any) string {
	return fmt.Sprintf(t.Get(locale), args...)
}
```

- `LocalizedText`(第4章)に **`Format` メソッドを追加**。
- 「その言語の文字列を選ぶ」→「`%s` や `%d` を引数で埋める」を1回で行う。
- `args ...any`:引数を何個でも受け取り、 `args...` でそのまま `Sprintf` に渡す。
- 使い方:`LocalizedText{"en": "Hello %s", "ja": "こんにちは%s"}.Format(locale, "alice")`

```go
func (s *Server) localeOfLocked(name string) string {
	if client := s.clients[name]; client != nil && client.locale != "" {
		return client.locale
	}
	return defaultLocale
}
```

- ある**プレイヤー名**から、その人の接続の言語を引く(通知を相手の言語で送るため)。

### 11-1-2 `sendPlayerEventLocked`

```go
func (s *Server) sendPlayerEventLocked(name, kind, text string) {
	client := s.clients[name]
	if client == nil {
		return
	}
	text = strings.NewReplacer("\r", " ", "\n", " ").Replace(text)
	client.enqueueEvent("EVT PLAYER " + kind + " " + text)
}
```

- その人だけに `EVT PLAYER 種別 文章` を送る。
- **`strings.NewReplacer("\r", " ", "\n", " ").Replace(text)`**:文章の中の改行を空白に置き換える。プロトコルは「1行1メッセージ」なので、 **文章に改行が混じると、2つのメッセージに分かれて壊れる**。
- `NewReplacer(置換前, 置換後, 置換前, 置換後…)` で複数の置き換えを一度に行う。

### 11-1-3 死亡メッセージ

```go
var deathTexts = map[string]LocalizedText{

	"attack_counter": {
		"en": "You were struck down by %s and your HP ran out.",
		"ja": "%sに打ち倒され、HPが尽きた。",
	},
	...
}
```

- **死因 → 言語別の文章テンプレート**の辞書。キーは `respawnPlayerLocked` に渡した `cause`(`"attack_counter"`、`"slip_past"`、`"hazard_lethal"`…)。
- 値の中で `{...}` を省略して、 `"attack_counter": { "en": ..., "ja": ... }` と書ける(要素の型が決まっているため、 `LocalizedText{...}` の型名を省略できる)。
- `%s` には敵の名前や部屋の名前が入る。
- **ヒントや答えを含めない**ことが方針(「何に殺されたか」だけを言う)。
- 敵以外への攻撃で増えた死因もここにあります。
  - `"attack_mighty"`:「◯◯に手を上げた。一撃が届く前に、あなたは打ち殺された。」(強いNPCへの攻撃、8-5のケース3)
  - `"attack_murder"`:「◯◯を手にかけた。運命の女神たちは、その報いにあなたの糸を断ち切った。」(一般人を殺した、ケース4)
- 死因を増やしたら、**`deathTexts` に文面を足す**だけで、メッセージが出るようになります(`notify_test.go` が、全死因の文面が英語・日本語とも空でなく、書式が壊れていないことを確認しています)。

```go
var respawnTail = LocalizedText{
	"en": "You awaken in %s with %d HP.",
	"ja": "%sで目を覚ました。HPは%dに減っている。",
}

var outcomeTexts = map[deathOutcome]LocalizedText{
	outcomeLost: {...},
	outcomeKept: {...},
}

var enemiesRecoverTail = LocalizedText{...}
```

- 死亡メッセージの**後ろにつなげる定型文**。
  - 復活した場所とHP。
  - 持ち物の結果(失った/守られた)。 `map[deathOutcome]LocalizedText`:**キーが `deathOutcome` 型**の辞書。
  - 「傷つけた敵はすべて全快した」。

### 11-1-4 `notifyDeathLocked`

```go
func (s *Server) notifyDeathLocked(name, cause string, outcome deathOutcome, args ...any) {
	locale := s.localeOfLocked(name)
	text, ok := deathTexts[cause]
	if !ok {
		return
	}
	message := strings.TrimSpace(text.Format(locale, args...))
	if s.world != nil {
		if room := s.world.Rooms[s.world.StartRoomID]; room != nil {
			message += " " + respawnTail.Format(locale, room.Name.Get(locale), respawnHP)
		}
	}
	if tail, ok := outcomeTexts[outcome]; ok {
		message += " " + tail.Get(locale)
	}
	message += " " + enemiesRecoverTail.Get(locale)
	s.sendPlayerEventLocked(name, "DEATH", message)
}
```

- 死んだ本人に、 **1通の長い文章**を作って送る。つなげる順番:
  1. 死因の文(`%s` を埋める)
  2. 「◯◯で目を覚ました。HPは20に…」
  3. 持ち物の結果(`outcomeNothingLost` は辞書に無いので何も足されない)
  4. 「傷つけた敵は全快した」
- `message += " " + ...`:文字列を後ろに足していく。
- 知らない死因(辞書に無い)なら、何も送らずに終了。

### 11-1-5 案内役 `guideNPC` と `sendGuideLocked`

```go
func (s *Server) guideNPC() *NPC {
	if s.world == nil {
		return nil
	}
	ids := make([]string, 0)
	for id, npc := range s.world.NPCs {
		if npc != nil && npc.Guide && npc.RoomID == s.world.StartRoomID {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Strings(ids)
	return s.world.NPCs[ids[0]]
}
```

- 開始部屋にいる「案内役」(`Guide` が `true` のNPC)を探す。複数いればID順で先頭。いなければ `nil`。

```go
func (s *Server) sendGuideLocked(name string, index int) {
	guide := s.guideNPC()
	if guide == nil {
		return
	}
	locale := s.localeOfLocked(name)
	for i := index; i < len(guide.Dialogue); i++ {
		if line := guide.Dialogue[i].Get(locale); strings.TrimSpace(line) != "" {
			s.sendPlayerEventLocked(name, "GUIDE", line)
		}
	}
}
```

- 案内役の台詞を**`index` 番目から最後まで**、1行ずつ `EVT PLAYER GUIDE ...` で送る。
  - 初めての接続:`index = 0` で全部。
  - 案内役に `TALK`:`index = 1`(最初の台詞は `TALK` の応答として返されているので、残りを通知で送る)。
- `for i := index; i < len(...); i++ {`:Goの**従来型の`for`**。初期化; 条件; 後処理。
- 空の台詞は送らない。

### 11-1-6 `sendDeathHintLocked`— モイライのヒント

```go
// sendDeathHintLocked gives the player the Moirai's hint about what last killed them, once.
// Without a recorded death (or a hint for it) she adds nothing.
func (s *Server) sendDeathHintLocked(player *Player, name string) {
	subject := player.LastDeathSubject
	if subject == "" || s.world == nil {
		return
	}
	player.LastDeathSubject = ""
	hint, ok := s.world.Hints[subject]
	if !ok {
		return
	}
	s.sendPlayerEventLocked(name, "HINT", hint.Get(s.localeOfLocked(name)))
}
```

- **死んだあとに案内役(モイライ)へ `TALK` すると、死因にちなんだ神話のヒントを1回だけ**送ります(`TALK` の処理から呼ばれる、7-19)。
- `subject := player.LastDeathSubject`:`respawnPlayerLocked` が記録した、死因のID(敵・部屋・アイテム。8-4)。**空なら、まだ死んでいない**ので何もしません。
- `player.LastDeathSubject = ""`:**読んだらすぐ空にします**。これで、同じ死因のヒントは1回しか出ません。もう一度ほしければ、もう一度死ぬ必要があります(「ヒントは頼めば1回だけ」という設計)。
- `hint, ok := s.world.Hints[subject]`:`world.json` の `"hints"` から、その死因のヒントを引きます(5-5)。**ヒントが書いていない死因なら**、`ok` が `false` で何も送りません。
- `EVT PLAYER HINT 文` として送ります。文は**答えを言わず、神話の言い回しで方向だけ示す**ものです(例:ポリュペモスに殺されたら「オデュッセウスは素手で一つ目の巨人に挑んだのではない。火で固めたオリーブの杭が、その役目を果たした。」)。ヒントは `world.json` の `hints` に、全部で22件あります。

## 11-2 `flavor.go`(139行)— 部屋への実況

### 11-2-1 `flavor` の型と文章表

```go
type flavor struct {
	key    string
	player string
	npc    *NPC
	room   *Room
	n      int
	m      int
}
```

- 実況1件の**材料**。小文字始まりの型・フィールドなので、このパッケージの中だけで使う。
  - `key`:文章の種類(`"attack_hit"` など)。
  - `player`:行動した人の名前。
  - `npc` / `room`:関係するNPC・部屋(無いときは `nil`)。
  - `n` / `m`:数字(ダメージ、減った仲間の数など)。

```go
var flavorTexts = map[string]LocalizedText{
	"attack_unwinnable": {
		"en": "{player} attacks {npc} and is driven back, losing {n} crew.",
		"ja": "{player}は{npc}に挑んだが押し返され、仲間を{n}人失った。",
	},
	...
}
```

- 実況文のテンプレート表。 `%s` ではなく **`{player}` `{npc}` `{room}` `{n}` `{m}`** という目印を使う。(`Sprintf` だと引数の**順番**に縛られるが、目印方式なら、言語によって語順が違っても書ける。)
- 敵以外への攻撃のための実況が3つ増えています(`flavorTexts` のキー)。
  - `"attack_mighty"`:「{player}は{npc}に手を上げ、一撃が届く前に打ち殺された。」
  - `"attack_wounded"`:「{player}は{npc}に{n}のダメージを与えた。」(一般人を傷つけた。反撃は無いので `{m}` は使わない)
  - `"attack_murder"`:「{player}は{npc}を手にかけ、運命の女神たちにその糸を断ち切られた。」

### 11-2-2 `text`

```go
func (f flavor) text(locale string) string {
	npcName, roomName := "", ""
	if f.npc != nil {
		npcName = f.npc.Name.Get(locale)
	}
	if f.room != nil {
		roomName = f.room.Name.Get(locale)
	}
	return strings.NewReplacer(
		"{player}", f.player,
		"{npc}", npcName,
		"{room}", roomName,
		"{n}", strconv.Itoa(f.n),
		"{m}", strconv.Itoa(f.m),
	).Replace(flavorTexts[f.key].Get(locale))
}
```

- レシーバが `(f flavor)`(ポインタではない)。中身を**読むだけ**で書き換えないので、コピーで構わない。
- `npcName, roomName := "", ""`:2つの変数を同時に初期化。
- `NPC`/`Room` が `nil` でなければ、その言語の名前を取り出す。
- `strconv.Itoa(n)`:整数を文字列に変換(`strconv` パッケージ。 import で書いてある)。
- `NewReplacer(...).Replace(テンプレート)`:テンプレート中の目印を一度に置き換える。

### 11-2-3 `broadcastFlavorLocked`

```go
func (s *Server) broadcastFlavorLocked(roomID string, f flavor) {
	for playerName, other := range s.players {
		if other.RoomID != roomID {
			continue
		}
		if recipient := s.clients[playerName]; recipient != nil {
			event := "EVT ROOM COMBAT " + f.text(s.localeOfLocked(playerName))
			if len(event) > maxProtocolLineBytes {
				event = event[:maxProtocolLineBytes-len("...")]
				for !utf8.ValidString(event) {
					event = event[:len(event)-1]
				}
				event += "..."
			}
			recipient.enqueueEvent(event)
		}
	}
}
```

- 指定の部屋にいる**全員**に、実況を送る。
- **受け取る人ごとに言語が違うかもしれない**ので、ループの中で `f.text(その人の言語)` を作る。日本語の人には日本語、英語の人には英語。
- 長いプレイヤー名などで1行が65,535バイトを超える場合は、UTF-8の文字境界で切り詰めて`...`を付けます。標準設定のCLIのScannerでも受信でき、短い実況文はそのまま送ります。

---

# 第12章 セーブ(`player_store.go`, `item_store.go`)

セーブは2つのJSONファイルです。

| ファイル | 中身 |
|---|---|
| `saves/playerdata.json` | 全プレイヤー(名前→`Player`) |
| `saves/itemdata.json` | 普通のアイテムの置き場所(アイテムID→部屋ID) |

## 12-1 `player_store.go`(114行)

```go
const playersSaveFile = "playerdata.json"

func (s *Server) playersSavePath() string {
	return filepath.Join(s.saveDir, playersSaveFile)
}
```

- `filepath.Join(a, b)`:OSに合った区切り文字でパスをつなぐ(`saves/playerdata.json`)。

### `loadPlayers`

```go
func (s *Server) loadPlayers() (map[string]*Player, error) {
	data, err := os.ReadFile(s.playersSavePath())
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]*Player), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read player state: %w", err)
	}
	if len(data) == 0 {
		return make(map[string]*Player), nil
	}
```

- ファイルを読む。
- **ファイルが無い**(`os.ErrNotExist`)のは、初回起動なので正常。空の辞書を返す。
- それ以外の読み込みエラーは本物のエラー。
- ファイルが**空**(0バイト)でも、空の辞書を返す。

```go
	var players map[string]*Player
	if err := json.Unmarshal(data, &players); err != nil {
		return nil, fmt.Errorf("decode player state: %w", err)
	}
	if players == nil {
		return nil, errors.New("invalid player state: expected object")
	}
	for name, player := range players {
		if player == nil || player.Name != name {
			return nil, fmt.Errorf("player name mismatch for %q", name)
		}
		for questID, state := range player.Quests {
			if state == nil {
				return nil, fmt.Errorf("saved player %q has null quest state for %q", name, questID)
			}
		}
	}
	return players, nil
}
```

- JSONを辞書に変換。
- ファイルの中身が `null` だと `players` が `nil` になるので、それは不正として弾く。
- 辞書のキーと、中の `Name` が**一致しているか**確認(手で書き換えて壊れたデータに気づくため)。
- クエストの各状態が`null`なら、プレイヤー名とクエストIDを示すエラーで読み込みを拒否します。セーブファイルは書き換えません。クエスト辞書自体の省略・`null`・空の辞書は受け付けます。

### `savePlayer` と `writePlayers`

```go
func (s *Server) savePlayer(player *Player) error {
	if player == nil {
		return errors.New("missing player state")
	}
	players, err := s.loadPlayers()
	if err != nil {
		return err
	}
	players[player.Name] = player
	return s.writePlayers(players)
}
```

- 1人ぶんを保存するときも、 **全員のファイルを読み → 1人ぶんを差し替え → 全部書き直す**。シンプルな作りで、人数が少ない前提。

```go
func (s *Server) writePlayers(players map[string]*Player) error {
	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		return fmt.Errorf("encode player state: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(s.saveDir, 0700); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}
```

- `json.MarshalIndent(players, "", "  ")`:**見やすい字下げ付きのJSON**にする。
- 最後に改行 `'\n'` を足す。 `'...'`(シングルクォート)は1文字(`rune`/`byte`)を表す。
- `os.MkdirAll(フォルダ, 0700)`:フォルダを(途中のフォルダごと)作る。無ければ作り、あれば何もしない。 `0700` は**権限(本人だけが読み書き可能)**。

```go
	temp, err := os.CreateTemp(s.saveDir, ".players-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary save: %w", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("write player state: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync player state: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close player state: %w", err)
	}
	if err := os.Rename(temp.Name(), s.playersSavePath()); err != nil {
		return fmt.Errorf("replace player state: %w", err)
	}
	return nil
}
```

- **本番のファイルに直接書かない**のがポイント(アトミックな書き込み)。
  1. `os.CreateTemp`:同じフォルダに**一時ファイル**(`.players-12345.tmp`)を作る。 `*` のところにランダムな数字が入る。
  2. 一時ファイルにデータを書く。
  3. `Sync()`:OSのキャッシュから**ディスクに確実に書き出す**。
  4. `Close()`:閉じる。
  5. **`os.Rename`**:一時ファイルを本番のファイル名に**入れ替える**。同じディスク内のrenameは一瞬で終わり、途中の状態が見えない。
- 書き込み中にサーバーが落ちても、**古いセーブか新しいセーブのどちらかが必ず残る**(壊れた中途半端なファイルにならない)。
- `defer os.Remove(temp.Name())`:失敗したときに一時ファイルが残らないよう掃除する。成功時はもう名前が変わっているので、 `Remove` は何もしない(エラーは無視される)。
- `defer temp.Close()`:二重に `Close` しても問題ない(2回目は無視される)。

## 12-2 `item_store.go`(86行)

プレイヤーのセーブと同じ作りで、データが「アイテムID→部屋ID」の辞書です。

```go
const itemsSaveFile = "itemdata.json"

func (s *Server) itemsSavePath() string {
	return filepath.Join(s.saveDir, itemsSaveFile)
}
```

- `itemsSaveFile`:アイテムの場所を保存するファイル名(`saves/itemdata.json`)。
- `filepath.Join(s.saveDir, itemsSaveFile)`:保存フォルダとファイル名を、**OSに合った区切り文字**でつなぎます(`/` を自分で書かない)。

### `loadItemLocations`— 読み込み

```go
func (s *Server) loadItemLocations() (map[string]string, error) {
	data, err := os.ReadFile(s.itemsSavePath())
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read item locations: %w", err)
	}
	var locations map[string]string
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("decode item locations: %w", err)
	}
	if locations == nil {
		return nil, errors.New("invalid item locations: expected object")
	}
	return locations, nil
}
```

- 「アイテムID → 今ある部屋のID」の辞書を、ファイルから読みます。`loadPlayers`(12-1)と**同じ作り**です。
- **ファイルが無い**(`os.ErrNotExist`)のは初回起動なので、エラーにせず空の辞書を返します。`errors.Is(err, 目印)` は、エラーが「その種類のエラー」かを調べる関数です。
- それ以外の読み込みエラーは本物の失敗として、`%w` で**元のエラーを包んで**返します(`%w` は、包んだエラーを後で `errors.Is` で取り出せるようにする書式です)。
- `json.Unmarshal(data, &locations)`:JSONを辞書に変換します。`&locations` は「この変数の置き場所」を渡す書き方で、関数が中に値を書き込めるようにします。
- ファイルの中身が `null` だと辞書が `nil` になるので、**「オブジェクトであるべき」エラー**として弾きます。

### `writeItemLocations`— 書き込み(途中で壊れない)

```go
func (s *Server) writeItemLocations(locations map[string]string) error {
	data, err := json.MarshalIndent(locations, "", "  ")
	if err != nil {
		return fmt.Errorf("encode item locations: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(s.saveDir, 0700); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}
	temp, err := os.CreateTemp(s.saveDir, ".items-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary item save: %w", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("write item locations: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync item locations: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close item locations: %w", err)
	}
	if err := os.Rename(temp.Name(), s.itemsSavePath()); err != nil {
		return fmt.Errorf("replace item locations: %w", err)
	}
	return nil
}
```

- `json.MarshalIndent(locations, "", "  ")`:辞書をJSONにします。**2スペースで字下げ**するので、人が読めます。`data = append(data, '\n')` で、最後に改行を1つ足します。
- `os.MkdirAll(s.saveDir, 0700)`:保存フォルダを(無ければ)作ります。`0700` は「所有者だけが読み書きできる」権限です。
- **書き込みの手順は `writePlayers`(12-1)と同じ**で、途中でサーバーが止まってもファイルが壊れないようにしています。
  1. `os.CreateTemp`:同じフォルダに**一時ファイル**を作る。
  2. `defer os.Remove(temp.Name())` と `defer temp.Close()`:関数を抜けるとき、**後始末**(一時ファイルの削除と閉じる処理)を必ず行う。`defer` は後ろから順に実行されます。
  3. `temp.Write` → `temp.Sync`(OSのバッファからディスクへ**確実に書き出す**)→ `temp.Close`。
  4. `os.Rename`:一時ファイルを本番の名前に**一瞬で入れ替える**。書き込み途中の壊れた状態を、読む側が見ることはありません。
- 成功すれば、`defer os.Remove` は「もう名前が変わって存在しない」ので何も起きません。

### `restoreItemLocations`(起動時に呼ばれる)

```go
func (s *Server) restoreItemLocations() error {
	locations, err := s.loadItemLocations()
	if err != nil {
		return err
	}
	for itemID, roomID := range locations {
		if s.world.Items[itemID] == nil {
			return fmt.Errorf("saved item %q does not exist", itemID)
		}
		if s.world.Rooms[roomID] == nil {
			return fmt.Errorf("saved item %q has unknown room %q", itemID, roomID)
		}
	}
	for itemID, roomID := range locations {
		if s.world.Items[itemID].Renewable {

			continue
		}
		s.world.Items[itemID].RoomID = roomID
	}
	return nil
}
```

- 1周目:セーブにあるアイテムと部屋が、今の世界に**実在するか**を全部検査(途中まで書き換えてからエラーにならないよう、先に全部調べる)。
- 2周目:実際に、アイテムの `RoomID` を**セーブ上の場所に更新**する。再生アイテムは動かないので対象外。

### `restoreItemOwnership`(`player_store.go` にある。起動時に呼ばれる)

```go
func (s *Server) restoreItemOwnership() error {
	players, err := s.loadPlayers()
	if err != nil {
		return err
	}
	owners := make(map[string]string)
	for name, player := range players {
		for _, itemID := range player.Inventory {
			if s.world.Items[itemID] == nil || s.world.Items[itemID].Renewable {

				continue
			}
			if owner, exists := owners[itemID]; exists {
				return fmt.Errorf("item %q appears in inventories of %q and %q", itemID, owner, name)
			}
			owners[itemID] = name
		}
	}
	for itemID := range owners {
		s.world.Items[itemID].RoomID = ""
	}
	return nil
}
```

- 誰かがすでに持っている(普通の)アイテムを、**部屋から取り除く**(`RoomID = ""`)。
- 同じアイテムが**2人の持ち物に入っていたらエラー**:データ破損の検知。(アイテムは世界に1つしか無いはず。)
- `owners`:アイテムID→持ち主の名前。

---

# 第13章 ログと不正利用の検知(`logging.go`)

## 13-1 import と定数

```go
import (
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	maxLoggedText = 300

	connectWindow        = 10 * time.Second
	maxConnectsPerWindow = 8

	commandWindow        = time.Second
	maxCommandsPerWindow = 20

	abuseWarnInterval = 5 * time.Second
)
```

- `log/slog`:Go標準の**構造化ログ**(キーと値のペアで記録し、JSONにできる)。
- `maxLoggedText`:ログに書く文字数の上限(300)。
- 接続の連打:**10秒に8回を超えたら**警告。
- コマンドの連打:**1秒に20回を超えたら**警告。
- 同じ相手への警告は、5秒に1回まで(ログが警告で埋まらないように)。

## 13-2 ロガーの準備

```go
var logger = newLogger(os.Stderr, slog.LevelInfo)

func newLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}
```

- `logger`:全ファイルから使えるグローバル変数。最初は「標準エラー出力へ、情報レベル以上を」。
- `io.Writer`:書き込める何か(ファイル、標準出力、テスト用のバッファ)を表す interface。テストでログを文字列に取り込める。
- `slog.NewJSONHandler`:ログを1行1件のJSONで出力する。

```go
func setupLogging() (closeFn func()) {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("TAP_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
```

- 環境変数 `TAP_LOG_LEVEL` でログの詳しさを切り替える(`debug` / `warn` / `error`。指定が無ければ `info`)。
- `os.Getenv("名前")`:環境変数を読む。
- 戻り値 `(closeFn func())`:名前付き戻り値。 **関数を返す**。

```go
	var out io.Writer = os.Stderr
	closeFn = func() {}
	if path := os.Getenv("TAP_LOG_FILE"); path != "" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			newLogger(os.Stderr, slog.LevelInfo).Error("open_log_file", "path", path, "error", err.Error())
		} else {
			out = io.MultiWriter(os.Stderr, file)
			closeFn = func() { file.Close() }
		}
	}
	logger = newLogger(out, level)
	slog.SetDefault(logger)
	return closeFn
}
```

- 環境変数 `TAP_LOG_FILE` にパスがあれば、そのファイルにも書き出す。
  - `os.O_CREATE|os.O_APPEND|os.O_WRONLY`:無ければ作る・**末尾に追記**・書き込み専用。 `|` でフラグをまとめる。
  - `0o600`:権限(本人だけ読み書き)。
  - `io.MultiWriter(a, b)`:**同じ内容を両方に書く**(画面とファイルの両方)。
- ファイルを開けなかったときも、サーバーは止めず画面だけにログを出す。
- `closeFn`:空の関数が初期値。ファイルを開けたときだけ「ファイルを閉じる関数」に差し替える。 `main.go` はこれを `defer` で呼ぶ。
- `slog.SetDefault(logger)`:標準の `slog.Info(...)` でも同じロガーが使われるように。

```go
func fatal(msg string, err error) {
	logger.Error(msg, "error", err.Error())
	os.Exit(1)
}
```

- 致命的なエラーをログに書いて、プログラムを異常終了(終了コード1)させる。

## 13-3 補助関数

```go
func clip(s string) string {
	runes := []rune(s)
	if len(runes) <= maxLoggedText {
		return s
	}
	return string(runes[:maxLoggedText]) + "…"
}
```

- 長すぎる文字列をログ用に**300文字で切る**。
- `[]rune(s)`:文字列を**1文字ずつ**(日本語も1文字と数える)のスライスに変換。 `len(s)` は**バイト数**なので、日本語を途中で切ると文字化けする。 `rune` に変換してから切れば安全。

```go
func hostOf(remote string) string {
	if host, _, err := net.SplitHostPort(remote); err == nil {
		return host
	}
	return remote
}
```

- `"1.2.3.4:5555"` → `"1.2.3.4"`。 `net.SplitHostPort` はアドレスをホストとポートに分ける。ポート番号は使わないので `_`。

```go
func logOutbound(client *serverClient, data []byte) {
	player, command := client.context()
	for _, line := range strings.Split(strings.TrimRight(string(data), "\r\n"), "\n") {
		switch {
		case line == "OK" || strings.HasPrefix(line, "OK "):
			logger.Info("response", "remote", client.remote, "player", player, "command", command, "line", clip(line))
		case strings.HasPrefix(line, "ERR "):
			code, _, _ := strings.Cut(strings.TrimPrefix(line, "ERR "), " ")
			logger.Warn("error_response", "remote", client.remote, "player", player, "command", command, "code", code, "line", clip(line))
		case strings.HasPrefix(line, "EVT "):
			logger.Debug("event", "remote", client.remote, "player", player, "line", clip(line))
		}
	}
}
```

- 6-5の `writeLoop` から、送信に成功するたびに呼ばれる。
- 送った内容を**種類別**にログに記録:
  - `OK` → 情報レベル(`response`)
  - `ERR` → **警告レベル**(`error_response`)、エラーコードも抜き出して記録。
  - `EVT` → デバッグレベル(`event`)。件数が多いので、通常は出ない。
- `strings.Cut(s, " ")`:最初の空白で2つに切る。 ここでは先頭の単語(エラーコード)だけを取り出している。
- どのプレイヤーのどのコマンドへの応答かは、 6-10の `client.context()` で取る。

## 13-4 接続の連打の検知 `abuseMonitor`

```go
type abuseMonitor struct {
	mu       sync.Mutex
	connects map[string][]time.Time
	lastWarn map[string]time.Time
}
```

- IPごとに「最近の接続時刻の一覧」と「最後に警告した時刻」を持つ。
- **`mu` は `Server.mu` とは別の鍵**。接続のたびに呼ばれるので、ゲーム処理の鍵と分けて、お互いを待たせない。

```go
func (m *abuseMonitor) noteConnection(ip string, now time.Time) {
	m.mu.Lock()
	recent := m.connects[ip][:0]
	for _, t := range m.connects[ip] {
		if now.Sub(t) < connectWindow {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	m.connects[ip] = recent
	count := len(recent)
	warn := count > maxConnectsPerWindow && now.Sub(m.lastWarn[ip]) >= abuseWarnInterval
	if warn {
		m.lastWarn[ip] = now
	}
	m.mu.Unlock()
	if warn {
		logger.Warn("abuse_rapid_connections", "ip", ip, "connections", count, "window_ms", connectWindow.Milliseconds())
	}
}
```

- `m.connects[ip][:0]`:同じ配列を使い回して長さ0にする(メモリの再確保を減らす定番の書き方)。
- 古い記録(10秒より前)を捨てて、新しい接続時刻を足す。
- 「10秒内の接続数が8を超えた」かつ「前回の警告から5秒以上」なら警告する。
- **ログ出力は鍵を外してから**行う(ログの書き込みに時間がかかっても、他の接続を待たせない)。
- 警告するだけで、**接続は拒否しない**。

## 13-5 コマンドの連打の検知 `floodTracker`

```go
type floodTracker struct {
	times    []time.Time
	lastWarn time.Time
}

func (f *floodTracker) record(now time.Time) (count int, warn bool) {
	recent := f.times[:0]
	for _, t := range f.times {
		if now.Sub(t) < commandWindow {
			recent = append(recent, t)
		}
	}
	f.times = append(recent, now)
	count = len(f.times)
	if count > maxCommandsPerWindow && now.Sub(f.lastWarn) >= abuseWarnInterval {
		f.lastWarn = now
		return count, true
	}
	return count, false
}
```

- 同じ考え方を、**接続1本ごと**に行う(`handleClient` の中の変数 `flood`)。1接続だけが使うので鍵は不要。
- 直近1秒の間のコマンド数を数え、 20を超えたら警告(`warn = true`)。 戻り値は(件数, 警告するか)。
- `(count int, warn bool)` は名前付き戻り値。

---

# 第14章 CLIクライアント `cmd/cli/main.go`(153行)

**役割**:ターミナルで動くクライアント。方針は「**生プロトコルの中継**」で、打った行をそのままサーバーへ送り、サーバーから来た行をそのまま画面に出します。ゲームのルールは何も知りません。

## 14-1 import

```go
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)
```

- `log`:簡単なログ出力。 `log.Fatalf` は**メッセージを出して終了**する。
- `io`:`io.Reader`(読み取れるもの)の型。

## 14-2 `main`

```go
func main() {
	address := "127.0.0.1:4242"
	if len(os.Args) > 2 {
		log.Fatalf("usage: %s [host:port]", os.Args[0])
	}
	if len(os.Args) == 2 {
		address = os.Args[1]
	}
```

- `os.Args`:**コマンドライン引数**のスライス。 `os.Args[0]` はプログラム自身の名前、 `os.Args[1]` が最初の引数。
- 引数なし → `127.0.0.1:4242`(自分のPCの4242番)に接続。
- 引数1つ → それを接続先に使う(`./cli 192.168.1.5:4242`)。
- 引数が2つ以上 → 使い方を表示して終了。 `%s` に `os.Args[0]` が入る。

```go
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to", address)
	runClient(conn.(*net.TCPConn), os.Stdin)
}
```

- `net.Dial`:サーバーに**接続しに行く**(サーバー側の `Accept` と対になる)。
- `%v`:値を標準の形式で表示する書式(エラーなど何でも表示できる)。
- `defer conn.Close()`:終了時に接続を閉じる予約。
- `conn.(*net.TCPConn)`:型アサーション。 `Dial` が返すのは `net.Conn`(interface)だが、実体は TCPの接続 `*net.TCPConn`。後で使う `CloseWrite()` はこちらの型にしかない。
- `os.Stdin`:**標準入力**(キーボード)。 `runClient` の第2引数に渡す。

## 14-3 `clientConn` interface

```go
type clientConn interface {
	net.Conn
	CloseWrite() error
}
```

- 「`net.Conn` の機能に加えて `CloseWrite()` ができるもの」という約束。
- `runClient` はこの interface で受け取るので、 **本物のTCP接続でも、テスト用の偽物でも**渡せる。第2引数 `stdin io.Reader` も同じ理由で、テストでは文字列を流し込める。

## 14-4 `runClient` の準備

```go
func runClient(conn clientConn, stdin io.Reader) {
	serverDone := make(chan error, 1)
	quitAcknowledged := make(chan struct{}, 1)
	quitRejected := make(chan struct{}, 1)
	var pendingMu sync.Mutex
	var pending []bool
```

- 3つのチャネル(いずれも容量1):
  - `serverDone`:**サーバーとの通信が終わった**ことを知らせる(エラーがあれば中に入れる)。
  - `quitAcknowledged`:`QUIT` に `OK bye` が返ってきた。
  - `quitRejected`:`QUIT` にエラーが返ってきた(終了できなかった)。
- `pending []bool`:**「送ったコマンドのうち、まだ応答が来ていないもの」を、送った順に並べたキュー**。各要素は「そのコマンドは `QUIT` か?」。
  - サーバーの応答は送った順に返るので、「いま来た応答は、キューの先頭のコマンドへのもの」と対応づけられる。
- `pendingMu`:`pending` を複数のgoroutineから触るので、鍵で守る。

## 14-5 受信用goroutine

```go
	go func() {
		scanner := bufio.NewScanner(conn)
		firstLine := true

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
```

- `go func() { ... }()`:**無名関数をその場でgoroutineとして起動**する書き方。
- サーバーから来た行を**1行ずつ読み、そのまま画面に出す**(`fmt.Println`)。

```go
			if firstLine {
				firstLine = false
				if line == "OK hello proto=1" {
					continue
				}
			}
			if line != "OK" && !strings.HasPrefix(line, "OK ") && !strings.HasPrefix(line, "ERR ") {
				continue
			}
```

- 最初の1行が `OK hello proto=1`(あいさつ)なら、**コマンドへの応答ではない**ので数えずに次へ。
- `OK` / `ERR` で始まらない行(= `EVT` 通知など)も応答ではないので、数えずに次へ。

```go
			pendingMu.Lock()
			isQuit := false
			if len(pending) > 0 {
				isQuit = pending[0]
				pending = pending[1:]
			}
			pendingMu.Unlock()
			if !isQuit {
				continue
			}
```

- 応答が来たので、キューの**先頭を取り出す**(`pending[1:]` で先頭を捨てる)。
- 先頭のコマンドが `QUIT` でなければ、これ以上やることは無い。

```go
			if line == "OK bye" {
				select {
				case quitAcknowledged <- struct{}{}:
				default:
				}
			} else if strings.HasPrefix(line, "ERR ") {
				select {
				case quitRejected <- struct{}{}:
				default:
				}
			}
		}

		serverDone <- scanner.Err()
	}()
```

- `QUIT` への応答が `OK bye` なら `quitAcknowledged` に合図、 `ERR` なら `quitRejected` に合図。
- `select { case ch <- v: ... default: }`:**詰まっていたら諦めて進む**送信。容量1のチャネルなので、受け取り手がまだ読んでいなくても、このgoroutineは止まらない。
- 読み取りが終わったら(サーバーが切れた)、 `serverDone` にエラー(正常終了なら `nil`)を入れて終わる。

## 14-6 入力用goroutine

```go
	input := make(chan string)

	go func() {
		defer close(input)

		scanner := bufio.NewScanner(stdin)

		for scanner.Scan() {
			input <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Println("input error:", err)
		}
	}()
```

- キーボードから1行読むたびに、 `input` チャネルへ送る。
- 入力が終わったら(Ctrl-D、またはパイプの終端)、 **`defer close(input)`** でチャネルを閉じる。閉じたことが、メインループへの「入力はもう無い」という合図になる。
- `make(chan string)`:**容量なし**のチャネル。メインループが受け取るまで、送信側は待つ。

## 14-7 メインループ

```go
	for {
		select {
		case line, ok := <-input:
			if !ok {
				if err := conn.CloseWrite(); err != nil {
					log.Println("close write failed:", err)
					return
				}
				input = nil
				continue
			}
```

- 3つの出来事のうち、起きたものを処理する。まず「入力が来た」。
- `line, ok := <-input`:チャネルから受け取る2値形式。**閉じられた**ときは `ok` が `false`。
- 入力が終わったら:
  - `conn.CloseWrite()`:**「もうこちらから送るものは無い」とサーバーに伝える**(読む側は開いたまま)。
  - **`input = nil`**:`nil` のチャネルからの受信は**永遠に待ち続ける**(= この `case` が二度と選ばれなくなる)。これで、閉じたチャネルから何度も「閉じてる」を受け取り続けるのを防ぐ。
  - その後も、サーバーの応答は受信用goroutineが読み続け、画面に出す。

```go
			if strings.TrimSpace(line) == "" {
				continue
			}

			isQuit := strings.EqualFold(strings.TrimSpace(line), "QUIT")
			pendingMu.Lock()
			pending = append(pending, isQuit)
			pendingMu.Unlock()
			if _, err := fmt.Fprintln(conn, line); err != nil {
				log.Println("send failed:", err)
				return
			}
```

- 空行は送らない。
- `QUIT` かどうかを調べて、 **送る前に `pending` に積む**(応答が先に来て対応がずれることを防ぐ)。
- 入力された行を、**そのまま**サーバーへ送る。

```go
			if isQuit {
				select {
				case <-quitAcknowledged:
					return
				case <-quitRejected:
					continue
				case err := <-serverDone:
					if err != nil {
						log.Println("receive error:", err)
					}
				}
				return
			}
```

- `QUIT` を送ったら、結果が分かるまで待つ。
  - `OK bye` が来た → 終了(`return`)。
  - エラーが来た → 終了できなかったので、 `continue` で通常の入力に戻る。
  - その前にサーバーが切れた → 終了。

```go
		case <-quitAcknowledged:
			return

		case err := <-serverDone:
			if err != nil {
				log.Println("receive error:", err)
			}
			fmt.Println("Disconnected from server")
			return
		}
	}
}
```

- 入力待ちの途中で `OK bye` が来た(別経路の `QUIT`)→ 終了。
- サーバーとの通信が終わった → 「Disconnected from server」と表示して終了。

---

# 第15章 GUIクライアント `cmd/gui/`

画面付きのクライアントです。ライブラリ **fyne**(GoのGUI)を使います。通信の仕組みはCLIと同じ考え方で、返ってきた内容を画面の部品に反映します。

## 15-0 GUIの全体像

```
マウスでボタン・対象を選択
   │ ui.send("MOVE north")
   ▼
protocolClient.Send ──▶ [送信goroutine] ──▶ サーバー
                                              │
サーバー ──▶ [受信goroutine] ──▶ incoming チャネル
                                              │
                         ui.receive ──▶ fyne.Do(画面更新) ──▶ handleMessage
```

- **fyneの決まり**:画面部品を触る処理は、専用のスレッドで行う。別のgoroutineから触るときは **`fyne.Do(func() {...})`** で包む。コードに何度も出てくる。
- ファイル構成は次のとおり。読む順番も同じです。
  1. `protocol.go`:通信(サーバーとの送受信)
  2. `model.go`:受け取ったデータの型
  3. `retro.go`:見た目(配色・フォント)
  4. `ui_locale.go`:言語切替
  5. `art_assets.go` / `item_photos.go`:画像
  6. `main.go`:接続・応答・イベントの処理
  7. `ui_layout.go`:画面組み立てと幅に応じた配置
  8. `ui_actions.go`:マウスでグループ・招待を選ぶ
  9. `ui_journal.go`:まわり・持ち物・クエストの表示
  10. `ui_bars.go`:HP・仲間のバー(15-11)
  11. `ui_story.go` / `ui_effects.go`:色付きの冒険ログと画面のフラッシュ(15-12、15-13)
  12. `ui_combat.go`:戦闘パネルと戦闘結果の表示(15-14)
  13. `ui_map.go` / `ui_endings.go`:ミニマップとエンディング図鑑(15-15、15-16)
  14. `ui_item_effects.go`:アイテム効果の表示(15-17)
  15. `cmd/server/gui_state.go`:GUI向けSTATE拡張の応答(15-18)
  16. テスト(15-19)

## 15-1 `protocol.go`(150行)— 通信層

### 15-1-1 メッセージの種類

```go
type messageKind uint8

const (
	messageGreeting messageKind = iota
	messageResponse
	messageEvent
	messageNotice
	messageDisconnect
)
```

- `messageKind uint8`:小さな整数(0〜255)を、**メッセージの種類を表す専用の型**にしている。
- **`iota`**:`const` の中で使うと、**行ごとに0, 1, 2…と自動で増える**数。 `messageGreeting`=0、 `messageResponse`=1、 `messageEvent`=2、 `messageNotice`=3、 `messageDisconnect`=4。(2行目以降は `messageKind = iota` を省略しても、前の行と同じ式が繰り返される。)
- 意味:あいさつ / コマンドへの応答 / 通知(EVT) / その他の行 / 切断。

```go
type serverMessage struct {
	kind    messageKind
	command string
	request string
	line    string
	err     error
}
```

- 受信した1件のメッセージ。 `command` は「この応答は何のコマンドへのものか」(例:`"MOVE"`)、 `request` は送った行そのもの(`"MOVE north"`)、 `line` は受信した行、 `err` は切断のときの理由。

### 15-1-2 `protocolClient`

```go
type protocolClient struct {
	conn     net.Conn
	outgoing chan string
	incoming chan serverMessage
	done     chan struct{}
	closeOne sync.Once
	mu       sync.Mutex
	pending  []string
}
```

- `outgoing`:送信待ちのコマンド(容量64)。 `incoming`:受信したメッセージ(容量128)。
- `done`:終了の合図。 `closeOne`:`Close` を1回だけ行うための `sync.Once`。
- `pending`:**送信済みで、応答待ちのコマンド行**のキュー。受信した応答と、どのコマンドへのものかを対応づける(CLIの `pending` と同じ考え方)。

```go
func newProtocolClient(conn net.Conn) *protocolClient {
	client := &protocolClient{
		conn:     conn,
		outgoing: make(chan string, 64),
		incoming: make(chan serverMessage, 128),
		done:     make(chan struct{}),
	}
	go client.writeLoop()
	go client.readLoop()
	return client
}
```

- 作ったらすぐ、**送信用と受信用の2つのgoroutine**を起動する。

### 15-1-3 `Send`

```go
func (client *protocolClient) Send(line string) error {
	if strings.TrimSpace(line) == "" || !utf8.ValidString(line) || strings.ContainsAny(line, "\r\n") {
		return errors.New("invalid command line")
	}
	select {
	case <-client.done:
		return net.ErrClosed
	default:
	}
	select {
	case client.outgoing <- line:
		return nil
	case <-client.done:
		return net.ErrClosed
	default:
		return errors.New("command queue is full")
	}
}
```

- 送る前に検査:空、不正なUTF-8、 **改行を含む**行は拒否(改行入りだと、プロトコルが壊れて2つのコマンドになってしまう)。
- すでに閉じていればエラー。
- キューに積む。**満杯なら待たずにエラー**(画面が固まらない)。

### 15-1-4 `Close` と `canPoll`

```go
func (client *protocolClient) Close() {
	client.closeOne.Do(func() {
		close(client.done)
		client.conn.Close()
	})
}

func (client *protocolClient) canPoll() bool {
	client.mu.Lock()
	pending := len(client.pending)
	client.mu.Unlock()
	return pending+len(client.outgoing) < 4
}
```

- `Close`:サーバー側の `serverClient.Close` と同じ書き方(1回だけ実行)。
- `canPoll`:定期更新(ポーリング)を送ってよいか。 **応答待ちとキューの合計が4つ未満**なら送ってよい。サーバーが遅いときに、定期更新で詰まらせないための判断。

### 15-1-5 `writeLoop`

```go
func (client *protocolClient) writeLoop() {
	for {
		select {
		case line := <-client.outgoing:
			client.mu.Lock()
			client.pending = append(client.pending, line)
			client.mu.Unlock()
			if err := client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				client.conn.Close()
				return
			}
			if _, err := fmt.Fprintln(client.conn, line); err != nil {
				client.conn.Close()
				return
			}
		case <-client.done:
			return
		}
	}
}
```

- キューから1行取り出し、 **送る直前に `pending` に記録**してから書く(応答が先に来てもずれないように)。
- 10秒の書き込み期限を付ける。失敗したら接続を閉じて終了。

### 15-1-6 `readLoop`

```go
func (client *protocolClient) readLoop() {
	defer close(client.incoming)
	defer client.Close()
	scanner := bufio.NewScanner(client.conn)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	firstLine := true
```

- 抜けるときに、 `incoming` を閉じて、接続も閉じる(2つの `defer` は**後に書いたものから先に**実行される)。
- `scanner.Buffer(初期バッファ, 最大)`:1行の最大サイズを**1MB**に広げる(標準は64KB)。

```go
	for scanner.Scan() {
		line := scanner.Text()
		message := serverMessage{line: line}
		switch {
		case firstLine && line == "OK hello proto=1":
			message.kind = messageGreeting
		case strings.HasPrefix(line, "EVT "):
			message.kind = messageEvent
		case line == "OK" || strings.HasPrefix(line, "OK ") || strings.HasPrefix(line, "ERR "):
			message.kind = messageResponse
			client.mu.Lock()
			if len(client.pending) > 0 {
				message.request = client.pending[0]
				message.command = strings.ToUpper(strings.Fields(message.request)[0])
				client.pending = client.pending[1:]
			}
			client.mu.Unlock()
		default:
			message.kind = messageNotice
		}
```

- 受信した行を分類する:
  1. 最初の行が `OK hello proto=1` → あいさつ。
  2. `EVT ` で始まる → **通知**(応答待ちのキューは消費しない)。
  3. `OK` / `ERR` → **応答**。 `pending` の先頭を取り出して「どのコマンドへの応答か」を付ける。 `strings.Fields(...)[0]` が最初の単語(コマンド名)。
  4. それ以外 → その他。

```go
		firstLine = false
		select {
		case client.incoming <- message:
		case <-client.done:
			return
		}
	}
	err := scanner.Err()
	if err == nil {
		err = io.EOF
	}
	select {
	case client.incoming <- serverMessage{kind: messageDisconnect, err: err}:
	case <-client.done:
	}
}
```

- 分類した結果を `incoming` に送る。
- ループが終わったら(サーバーが切れた)、理由を付けた**切断メッセージ**を送る。正常な切断はエラーが `nil` なので、 `io.EOF`(「終わり」を意味する標準のエラー)を入れる。

## 15-2 `model.go`(158行)— 受け取るデータの型

```go
type roomView struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

type lookView struct {
	Room    roomView `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
	// Defeated is the NPCs in the room this player has beaten (drawn lying down).
	Defeated []string `json:"defeated"`
}

type statusView struct {
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Status string `json:"status"`
}

type stateView struct {
	Crew            int      `json:"crew"`
	CrewInitialized bool     `json:"crew_initialized"`
	Players         []string `json:"players"`
	Group           string   `json:"group"`
	Invitations     []string `json:"invitations"`
}

type questView struct {
	QuestID  string `json:"quest_id"`
	Status   string `json:"status"`
	Progress string `json:"progress"`
}
```

- サーバーが返す JSON を受け取るための型。 `LOOK`、 `STATUS`、 `STATE`、 `QUESTS` の応答に対応する。
- `lookView` の `Defeated` は、サーバーの `LOOK` に**追加された欄**(7-12、8-7)です。**このプレイヤーが倒した、この部屋のNPCのID**が入ります。倒した敵がいなければサーバーは欄ごと省くので、受け取る側は `nil` になります。`composeScene`(15-5)がこれを見て、倒した敵を倒れた絵で描きます。サーバー側で定義した形(第4章の `roomView` など)と同じ。
- `stateView`は追加コマンドSTATEの応答です。`CrewInitialized`で「仲間をまだ付与していない」と「残り0人」を区別し、`Players`と`Invitations`はマウスで選ぶグループ操作に使います。15-11でサーバー側の処理を説明します。
- `json:` タグが応答のキーとフィールドを対応させる。たとえば `max_hp` は `MaxHP`、`quest_id` は `QuestID` に入る。

```go
type localizedName map[string]string

func (name localizedName) get(locale string) string {
	if value := name[locale]; value != "" {
		return value
	}
	return name["en"]
}

type catalogEntry struct {
	Name        localizedName     `json:"name"`
	Description localizedName     `json:"description"`
	Role        string            `json:"role"`
	GiverNPCID  string            `json:"giver_npc_id"`
	HP          int               `json:"hp"`
	Effects     []itemEffect      `json:"effects"`
	Exits       map[string]string `json:"exits"`
	Ending      *struct {
		ID         string        `json:"id"`
		Name       localizedName `json:"name"`
		RewardItem string        `json:"reward_item"`
		Blessing   *struct {
			God         localizedName `json:"god"`
			Name        localizedName `json:"name"`
			Description localizedName `json:"description"`
		} `json:"blessing"`
	} `json:"ending"`
	Hazard *struct {
		Type string `json:"type"`
	} `json:"hazard"`
}

type worldCatalog struct {
	Rooms  map[string]catalogEntry `json:"rooms"`
	Items  map[string]catalogEntry `json:"items"`
	NPCs   map[string]catalogEntry `json:"npcs"`
	Quests map[string]catalogEntry `json:"quests"`
}
```

- `world.json` から**表示名・説明文・危険の種類**を読むための型。`Name` と `Description` は言語ごとの文章を持つ。`Hazard` は無名構造体へのポインタで、危険が設定されていない部屋では `nil` になる。
- `Role`と`GiverNPCID`は、NPCの行に戦う・依頼のどのボタンを表示するかを決める情報です。
- `Effects []itemEffect` は、**アイテムの効果**(5-2、9-4)です。持ち物やまわりのアイテム行で、効果を色付きで表示するために読み込みます(15-17)。`itemEffect` はGUI側で定義した、種類と大きさだけの小さな型です。
- **JSONにあるキーのうち、構造体にあるものだけ**が読まれ、他は無視される。GUIは部屋・アイテム・NPC・クエストの辞書を使い、サーバーの戦闘処理などはここへ読み込まない。
- GUIが表示名を取るのは、サーバーが返すのが**IDだけ**(`LOOK` の `items` は ID の配列)だから。
- 部屋の情報は、出口の行き先を名前で表示する処理と、ゲームオーバー画面にも使う。

### `loadCatalog`

```go
func loadCatalog() (*worldCatalog, error) {
	var candidates []string
	for _, base := range []string{workingDirectory(), executableDirectory()} {
		for range 4 {
			candidates = append(candidates, filepath.Join(base, "data", "world.json"))
			base = filepath.Dir(base)
		}
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var catalog worldCatalog
		if err := json.Unmarshal(data, &catalog); err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		return &catalog, nil
	}
	return nil, fmt.Errorf("data/world.json was not found")
}
```

- `data/world.json` の**場所を探す**。「今のフォルダ」と「実行ファイルのあるフォルダ」それぞれについて、**出発点を含む4階層**を調べる。親へ3回さかのぼる範囲が候補になる。
- `filepath.Dir(base)`:1つ上のフォルダ。
- 候補を順に試して、最初に読めたファイルを使う。どこから起動しても動くようにするための工夫。
- `workingDirectory()` / `executableDirectory()` は、それぞれ `os.Getwd()` と `os.Executable()` の薄いラッパー。

### `label` と `decodeOK`

```go
func (catalog *worldCatalog) label(kind, id, locale string) string {
	if catalog == nil {
		return id
	}
	var entry catalogEntry
	switch kind {
	case "room":
		entry = catalog.Rooms[id]
	case "item":
		entry = catalog.Items[id]
	case "npc":
		entry = catalog.NPCs[id]
	case "quest":
		entry = catalog.Quests[id]
	}
	if name := entry.Name.get(locale); name != "" {
		return name
	}
	return id
}
```

- 画面に出す名前を選ぶ。名前が見つかれば `"黄金の羊毛皮"` のように表示する。部屋の名前も `kind == "room"` で取得できる。
- `catalog == nil`(`world.json` が読めなかった)でも**ポインタが `nil` のままメソッドを呼べる**ので、 `nil` の場合は単にIDを返す。
- 指定した言語の名前が無ければ英語を試し、名前自体が見つからなければIDを表示する。コマンドの送信や画像の選択には引き続きIDを使う。

```go
func decodeOK(line string, target any) error {
	if !strings.HasPrefix(line, "OK ") {
		return fmt.Errorf("unexpected response %q", line)
	}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "OK ")), target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
```

- `"OK {...json...}"` の形の応答から、 `OK ` を外して JSON を `target` にデコードする。
- `target any`:**どんな型へのポインタでも受け取れる**(`&lookView` でも `&statusView` でも)。

### `gameOverRoom`— 即死部屋の判定

```go
func (catalog *worldCatalog) gameOverRoom(roomID string) (catalogEntry, bool) {
	if catalog == nil {
		return catalogEntry{}, false
	}
	entry, ok := catalog.Rooms[roomID]
	if !ok || entry.Hazard == nil || entry.Hazard.Type != "lethal" {
		return catalogEntry{}, false
	}
	return entry, true
}
```

- ローカルのカタログから移動先の部屋を探す。`entry, ok` の `ok` は、そのIDが辞書に存在するかを表す。
- `Hazard.Type == "lethal"` の部屋なら、表示に使う名前・説明文と `true` を返す。カタログが無い、部屋が無い、危険が無い、種類が違う場合は `false`。
- 15-7-7のMOVE応答で呼ばれ、ゲームオーバー画面を開くか決める。判定に使う情報は、GUIが読み込んだ `data/world.json` にある。

## 15-3 `retro.go`(93行)— 見た目

```go
//go:embed assets/fonts/DroidSansFallbackFull.ttf
var droidFont []byte

var droidResource = fyne.NewStaticResource("DroidSansFallbackFull.ttf", droidFont)
```

- **`//go:embed ファイル`**:コンパイルのときに、**そのファイルの中身をプログラムの中に埋め込む**指示(`import _ "embed"` が必要)。直後の変数に中身が入る。
- 日本語のフォントを埋め込んであるので、**どのPCでも日本語が表示できる**(フォントが無い環境でも)。

```go
var (
	ink   = color.NRGBA{R: 5, G: 12, B: 30, A: 255}
	navy  = color.NRGBA{R: 12, G: 27, B: 56, A: 255}
	ivory = color.NRGBA{R: 244, G: 232, B: 191, A: 255}
	gold  = color.NRGBA{R: 204, G: 168, B: 90, A: 255}
	// dimGold is for frames and dividers: visible, but quieter than gold text and highlights.
	dimGold = color.NRGBA{R: 112, G: 96, B: 62, A: 255}
)
```

- 配色。ほぼ黒の紺(`ink`)、紺(`navy`)、象牙色(`ivory`)、金(`gold`)、**暗い金(`dimGold`)**。 `R, G, B, A` は赤・緑・青・不透明度(0〜255)。
- `dimGold` は**枠線と区切り線用の、控えめな金色**です。文字や選択中の強調に使う明るい金(`gold`)と分けることで、**枠が主張しすぎず、強調したい所が目立つ**ようにしています。

```go
type retroTheme struct{ base fyne.Theme }

func (t retroTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := storyColor(name); ok {
		return c
	}
	switch name {
	case theme.ColorNameBackground, theme.ColorNameMenuBackground:
		return ink
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return navy
	case theme.ColorNameForeground, theme.ColorNameForegroundOnPrimary:
		return ivory
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameSelection:
		return gold
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return dimGold
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 147, G: 153, B: 166, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 28, G: 38, B: 60, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 150, G: 156, B: 172, A: 255}
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t retroTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Symbol {
		return t.base.Font(style)
	}
	return droidResource
}

func (t retroTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t retroTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 15
	case theme.SizeNameHeadingText:
		return 19
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameButtonRadius, theme.SizeNameInputRadius,
		theme.SizeNameDialogRadius, theme.SizeNamePopupRadius:
		return 0
	}
	return t.base.Size(name)
}
```

- `type retroTheme struct{ base fyne.Theme }`:自作のテーマ。`base` は**元のテーマ**(Fyne標準)で、自分で決めなかった部分を任せます(**委譲**)。
- `Color`:色の名前(`name`)に対して、**使う色を返す**メソッドです。
  - 最初に `storyColor(name)`(15-12)を試し、**冒険ログ用に作った色の名前なら**その色を返します。
  - `switch name`:標準の色の名前ごとに、色を決めます。1つの `case` に名前を複数並べると「どれでも」です。
    - 背景・メニューの背景:`ink`(ほぼ黒の紺)。
    - ボタン・入力欄の背景:`navy`(紺)。
    - 文字・主役色の上の文字:`ivory`(象牙色)。
    - 主役の色(選択中のタブなど)・フォーカス・選択:`gold`(明るい金)。
    - **入力欄の枠・区切り線**:`dimGold`(暗い金)。
    - 入力欄の見本文字(`PlaceHolder`):灰色。**うすい文字色**で、`dimLabel`(15-8)もこの色を使います。
    - 押せないボタン・押せない文字:暗い灰色。
  - どれにも当てはまらなければ、`t.base.Color(name, theme.VariantDark)`、つまり**標準テーマの「暗いモード」の色**にします。ゲームは常に暗い配色なので、`variant`(明/暗)は無視して暗いほうを渡しています。
- `Font`:**使うフォント**を返します。記号(`Symbol`)用は標準のものにして、それ以外は**埋め込んだ日本語フォント**(`droidResource`)にします。これで、画面のすべての文字が日本語を含めて表示できます。
- `Icon`:アイコンは標準のまま(`t.base.Icon(name)`)。
- `Size`:大きさの指定。**文字は15**、見出しは19、**部品の間の余白は8**です。ボタン・入力欄・ダイアログ・ポップアップの**角丸は0**(レトロな四角いデザイン)。それ以外は標準に任せます。
- この4つのメソッド(`Color`・`Font`・`Icon`・`Size`)を持つ型は、Fyneの **`fyne.Theme`** として使えます(起動処理の `main.go` で `application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})` と設定します。15-7)。

- `fyne.Theme` という interface(`Color`、`Font`、`Icon`、`Size` の4メソッド)を**自分で実装**して、見た目を変える。
- 背景は `ink`、ボタンは `navy`、文字は `ivory`、強調(選択・フォーカス)は `gold`、**入力欄の枠と区切り線は `dimGold`**…と色を割り当て、決めていないものは元のテーマ(`t.base`)に任せる。
- 余白の大きさ(`theme.SizeNamePadding`)は**8**です(以前は6)。部品どうしの間隔を少し広げて、画面が詰まって見えないようにしました。
- `Font` は日本語フォントを返し、 `Size` は文字の大きさや余白を返す(角丸は0 = 四角く)。

```go
func framed(title string, content fyne.CanvasObject) fyne.CanvasObject {
	frame := canvas.NewRectangle(ink)
	frame.StrokeColor = dimGold
	frame.StrokeWidth = 1
	caption := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	inside := container.NewBorder(caption, nil, nil, nil, content)
	return container.NewStack(frame, container.NewPadded(inside))
}
```

- **枠付きのパネル**を作る部品。四角形(枠)の上に、見出し+中身を重ねる。枠は**暗い金の細線**(幅1)で、中身より目立たないようにしてあります。
- `container.NewBorder(上, 下, 左, 右, 中央)`:周りと中央に部品を配置する入れ物。 `nil` はその位置に何も置かない。
- `container.NewStack(a, b)`:**重ねて**表示する入れ物。

```go
func sceneImage() *canvas.Image {
	image := canvas.NewImageFromImage(loadArt("rooms", "unknown"))
	image.FillMode = canvas.ImageFillContain
	image.ScaleMode = canvas.ImageScaleSmooth
	return image
}
```

- 部屋の絵を表示する画像部品。最初は「不明の部屋」の絵。 `FillMode = Contain`(枠に収まるように拡縮)。

## 15-4 `ui_locale.go`(70行)— 言語切替と表示文

```go
func (ui *gui) tr(english, japanese string) string {
	if ui.locale == "ja" {
		return japanese
	}
	return english
}
```

- 画面の文字を選ぶ。 **`ui.tr("Connect", "接続")`** のように、英語と日本語を並べて書き、現在の言語の方を返す。

```go
func (ui *gui) switchLocale(locale string) {
	if locale != "ja" {
		locale = "en"
	}
	if locale == ui.locale {
		return
	}
	address := ui.hostEntry.Text
	name := ui.nameEntry.Text
	chatMessage := ui.chatEntry.Text
	chatScope := ui.chatScope.Selected
	messageTab := ui.messages.SelectedIndex()
	journalTab := ui.journal.SelectedIndex()
	room, inventory, quests, state := ui.room, ui.inventory, ui.quests, ui.state
	ui.clearChoices()
	ui.locale = locale
	ui.build()
	ui.hostEntry.SetText(address)
	ui.nameEntry.SetText(name)
	ui.chatEntry.SetText(chatMessage)
	ui.chatScope.SetSelected(chatScope)
	ui.messages.SelectIndex(messageTab)
	ui.journal.SelectIndex(journalTab)
	ui.showRoom(room)
	ui.showInventory(inventory)
	ui.showQuests(quests)
	ui.showState(state)
	ui.showFight(ui.fight)
	if ui.client != nil {
		ui.connectButton.Disable()
		ui.settingsButton.Disable()
		ui.languageSelect.Disable()
		ui.statusLabel.SetText(ui.tr("Connecting...", "接続中..."))
	}
	ui.window.Canvas().Unfocus()
}
```

- 言語を変えると、**画面全体を作り直す**(`ui.build()`)。作り直すと入力欄の内容が消えるので、 **先に退避しておき、作り直した後に戻す**。
- 接続先・名前・未送信のチャット・選択中のタブに加えて、部屋・持ち物・クエスト・仲間の表示データを保持します。ウインドウのサイズ変更では作り直さず、レイアウトだけが配置を変えます。
- 同じ言語ならすぐ終了。 `ja` 以外は `en` 扱い。

### `exitLabel`— 方角と行き先の名前

```go
var directionNames = map[string]string{"north": "北", "south": "南", "east": "東", "west": "西"}

func (ui *gui) exitLabel(direction, roomID string) string {
	name := direction
	if ui.locale == "ja" {
		if japanese, ok := directionNames[direction]; ok {
			name = japanese
		}
	}
	return name + ": " + ui.catalog.label("room", roomID, ui.locale)
}
```

- 日本語のときは、辞書にある方角を「北」「東」などへ変える。辞書に無い方角は受け取った文字列を使う。
- `catalog.label("room", ...)` で行き先の表示名を取り、「東: トロイアの浜」のように組み立てる。名前が無ければ部屋IDを表示する。
- まわりの「移動先」一覧で使う。ボタンの表示名は翻訳し、送信するMOVEには方角をそのまま使う。

### `statusWord`— HPとクエストの状態

```go
var statusWords = map[string]string{"healthy": "健康", "combat": "戦闘中", "active": "進行中", "completed": "達成"}

func (ui *gui) statusWord(status string) string {
	if ui.locale == "ja" {
		if japanese, ok := statusWords[status]; ok {
			return japanese
		}
	}
	return status
}
```

- 日本語表示のときに、HPの状態とクエストの状態を辞書の文章へ変える。
- 英語表示、または辞書に無い値なら元の文字列を返す。サーバーから受け取るJSONの値はそのまま使い、表示する段階で翻訳する。

## 15-5 `art_assets.go`(125行)— 画像

```go
const (
	artWidth     = 960
	artHeight    = 576
	spriteWidth  = 192
	spriteHeight = 312
)

//go:embed assets/rooms/*.png assets/npcs/*.png assets/npcs_defeated
var artAssets embed.FS
```

- 背景画像のサイズ(960×576)とNPC画像のサイズ(192×312)。
- `embed.FS`:埋め込んだファイルを**ファイルシステムのように**読める型。 `assets/rooms/*.png` のように `*` で複数をまとめて埋め込める。
- 3つめの `assets/npcs_defeated` は、`*.png` でなく**フォルダの名前**です。フォルダを指定すると、中のファイルを全部埋め込みます。こうしておくと、**倒された敵の絵が1枚も無くても**コンパイルが通ります(`*.png` と書くと、1枚も無いときに「一致するファイルが無い」エラーになる)。フォルダには説明用の `README.txt` を置いてあります。

```go
func loadArt(kind, id string) image.Image {
	if id == "" || strings.ContainsAny(id, `/\`) {
		id = "unknown"
	}
	data, err := artAssets.ReadFile("assets/" + kind + "/" + id + ".png")
	if err != nil {
		data, err = artAssets.ReadFile("assets/" + kind + "/unknown.png")
	}
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return img
}
```

- 画像を読み込む。 `kind` は `"rooms"` か `"npcs"`。
- IDに **`/` や `\` が含まれていたら `unknown` にする**:`../../etc/passwd` のようなID(パストラバーサル)で、意図しないファイルを読ませないための対策。
- 見つからなければ `unknown.png`。それも無ければ `nil`。
- バッククォートで囲んだ文字列 `` `/\` `` は**生の文字列**(`\` を特別扱いしない)。

```go
func composeScene(roomID string, npcs, defeated []string) *image.RGBA {
	background := image.NewRGBA(image.Rect(0, 0, artWidth, artHeight))
	if room := loadArt("rooms", roomID); room != nil {
		if room.Bounds().Dx() == artWidth && room.Bounds().Dy() == artHeight {
			draw.Draw(background, background.Bounds(), room, room.Bounds().Min, draw.Src)
		} else {
			xdraw.CatmullRom.Scale(background, background.Bounds(), room, room.Bounds(), draw.Src, nil)
		}
	}
```

- **部屋の背景にNPCの絵を重ねて、1枚の絵を作る**。`defeated` は**倒した敵のID一覧**(`LOOK` の `defeated`、15-2)で、含まれるNPCは立ち絵でなく**倒れた絵**で描きます(下の `defeatedSprite`)。
- `image.NewRGBA(...)`:空の画像(960×576)を作る。
- 部屋の画像がちょうど同じサイズなら、そのまま貼る(`draw.Draw`)。サイズが違えば **`CatmullRom.Scale`** で拡大縮小して貼る(`xdraw` は `golang.org/x/image/draw`)。

```go
	count := len(npcs)
	if count == 0 {
		return background
	}
	width := spriteWidth
	if count > 4 {
		width = (artWidth - 60) / count
	}
	if width < 42 {
		width = 42
	}
	height := spriteHeight * width / spriteWidth
```

- NPCがいなければ背景だけ返す。
- 5体以上なら、 **幅を縮めて横に収まるように**する(最小42ピクセル)。高さは**縦横比を保つ**ように計算。

```go
	for i, id := range npcs {
		x := 645
		if count > 1 {
			x = 36 + i*(artWidth-width-72)/(count-1)
		}
		if count == 2 {
			x = 495 + i*219
		}
		y := artHeight - height - 21
		shadow := &image.Uniform{C: color.NRGBA{R: 6, G: 15, B: 32, A: 125}}
		draw.Draw(background, image.Rect(x+21, artHeight-45, x+width-21, artHeight-30), shadow, image.Point{}, draw.Over)
		dest := image.Rect(x, y, x+width, y+height)
		var sprite image.Image
		if slices.Contains(defeated, id) {
			sprite = defeatedSprite(id)
			if sprite != nil {
				// A body lying down is wider than a person standing, and rests on the floor.
				lyingWidth := width * 3 / 2
				lyingHeight := lyingWidth * sprite.Bounds().Dy() / sprite.Bounds().Dx()
				dest = image.Rect(x-(lyingWidth-width)/2, artHeight-lyingHeight-27, x+(lyingWidth+width)/2, artHeight-27)
			}
		} else {
			sprite = loadArt("npcs", id)
		}
		if sprite == nil {
			continue
		}
		if sprite.Bounds().Dx() == width && sprite.Bounds().Dy() == height {
			draw.Draw(background, dest, sprite, sprite.Bounds().Min, draw.Over)
		} else {
			xdraw.CatmullRom.Scale(background, dest, sprite, sprite.Bounds(), draw.Over, nil)
		}
	}
	return background
}
```

- NPCを1体ずつ配置する。
  - 1体:右寄りの固定位置(x=645)。2体:決まった2か所。3体以上:左右に**等間隔**。
  - 足元に半透明の影(`shadow`)を描いてから、NPCの絵を重ねる(`draw.Over` = 透明部分を透かして重ねる)。
- 絵が見つからなければ、そのNPCは飛ばす。
- **倒した敵**(`slices.Contains(defeated, id)`)の場合:
  - 絵は `defeatedSprite(id)`(下)で用意します。
  - **倒れた体は、立っているときより横に長い**ので、描く幅を**1.5倍**にして(`lyingWidth := width * 3 / 2`)、高さは**絵の縦横比のまま**計算します。立ち位置の中心は変えず、左右に均等に広げます。
  - 縦の位置は、**床に横たわる**ように、立ち絵より下(`artHeight - lyingHeight - 27`)に置きます。
  - `slices.Contains(スライス, 値)` は、スライスに値が**含まれるか**を返す標準ライブラリの関数です。

### `defeatedSprite`— 倒れた絵を用意する

```go
func defeatedSprite(id string) image.Image {
	if data, err := artAssets.ReadFile("assets/npcs_defeated/" + id + ".png"); err == nil {
		if img, err := png.Decode(bytes.NewReader(data)); err == nil {
			return img
		}
	}
	standing := loadArt("npcs", id)
	if standing == nil {
		return nil
	}
```

- **専用の絵があれば、それを使います**(`assets/npcs_defeated/<NPCのID>.png`)。無い(`err != nil`)か、読めないときは、次の自動生成に進みます。
- 専用の絵が無ければ、**立ち絵を読み込んで加工します**。立ち絵も無ければ `nil`(描かない)。
- `if data, err := ...; err == nil {`:`if 準備文; 条件 {` の形で、**読めたときだけ**中に入ります。内側にも同じ名前の `img, err` を作っていますが、それは**内側の `if` の中だけの別の変数**です(シャドーイング)。

```go
	bounds := standing.Bounds()
	// Turn 90 degrees clockwise: the old top (the head) ends up on the right.
	lying := image.NewNRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c := color.NRGBAModel.Convert(standing.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)
			grey := uint8((int(c.R)*3 + int(c.G)*6 + int(c.B)) / 10 * 55 / 100)
			lying.SetNRGBA(bounds.Dy()-1-y, x, color.NRGBA{R: grey + 12, G: grey, B: grey, A: c.A})
		}
	}
	return lying
}
```

- **立ち絵を暗くして、横に倒した絵を作ります**(専用の絵が無いときの代わり)。
- `lying`:**幅と高さを入れ替えた**空の画像(192×312 の立ち絵なら 312×192)。`image.NewNRGBA` は、透明度付きの画像です。
- 二重の `for`:立ち絵の**全部のピクセル**を1つずつ見ます。`y` が行、`x` が列です。
- `color.NRGBAModel.Convert(...).(color.NRGBA)`:ピクセルの色を `NRGBA` 型に**変換**します。最後の `.(color.NRGBA)` は**型アサーション**で、「この値は `NRGBA` 型のはず」と取り出す書き方です。
- `grey := uint8((R*3 + G*6 + B) / 10 * 55 / 100)`:
  - `R*3 + G*6 + B` を10で割って、**人の目に近い明るさ**(緑を重く、青を軽く)の灰色にします。
  - そこに `55 / 100` を掛けて、**元の55%の明るさ**にします(暗くする)。
  - `uint8(...)` は 0〜255 の小さな整数への型変換です。
- `lying.SetNRGBA(bounds.Dy()-1-y, x, ...)`:**座標を入れ替えて書き込みます**。立ち絵の `(x, y)` の点を、横倒しの絵の `(高さ-1-y, x)` に置くと、**時計回りに90度回した**ことになります(コメントのとおり、頭が右に来ます)。
- 色は `R: grey + 12, G: grey, B: grey`:**赤だけ少し強く**して、血の気が引いた灰色の中に、わずかに赤みを残します。透明度 `A` は元のまま(輪郭の外は透明のまま)。
- 専用の絵(PNG、11枚。約38〜75KB)を置くと自動でそちらに置き換わります。絵の向きは**頭が右**という約束です(この関数のコメントの向き)。

> **ねらい**:絵が無くても「倒した」ことが分かる最低限の見た目を**コードで作り**、絵が用意できたら**ファイルを置くだけ**で差し替わるようにしています。`art_test.go` の `TestDefeatedEnemyIsDrawnDifferently` が、倒した敵と生きている敵の絵が**違うこと**と、倒れた絵が**横長**であることを確認します。

## 15-6 `item_photos.go`(29行)

```go
//go:embed assets/items/*.png
var itemPhotoAssets embed.FS

func itemPhotoCard(id string) fyne.CanvasObject {
	data, err := itemPhotoAssets.ReadFile("assets/items/" + id + ".png")
	if err != nil {
		return nil
	}
	photo := canvas.NewImageFromResource(fyne.NewStaticResource(id+".png", data))
	photo.FillMode = canvas.ImageFillContain
	photo.ScaleMode = canvas.ImageScaleSmooth
	frame := canvas.NewRectangle(color.NRGBA{R: 12, G: 27, B: 56, A: 255})
	frame.StrokeColor = gold
	frame.StrokeWidth = 2
	return container.NewGridWrap(fyne.NewSize(104, 104), container.NewStack(
		frame, container.NewPadded(photo),
	))
}
```

- アイテムのIDから、 **金の枠付きの画像カード**(104×104)を作る。画像が無いアイテムは `nil`(カードを出さない)。

## 15-7 `main.go`(569行)— 接続と応答の処理

画面の組み立ては15-8へ、グループ選択は15-9へ、一覧の表示は15-10へ分けています。ゲーム中の文字入力はチャットだけです。接続先と名前は接続設定の画面で入力します。

### 15-7-1 型の定義

#### `menuChoice`

```go
type menuChoice struct {
	label   string
	command string
	action  string
}
```

- `label`は表示文、`command`は送信する文字列、`action`は次の選択画面を開くための値です。

#### `gui`

```go
type gui struct {
	window           fyne.Window
	catalog          *worldCatalog
	client           *protocolClient
	pollStop         chan struct{}
	dialing          bool
	connected        bool
	locale           string
	room             lookView
	inventory        []string
	choices          []menuChoice
	state            stateView
	stateUnavailable bool
	quests           []questView
	hostEntry        *widget.Entry
	nameEntry        *widget.Entry
	languageSelect   *widget.Select
	connectButton    *widget.Button
	quitButton       *widget.Button
	settingsButton   *widget.Button
	statusLabel      *widget.Label
	roomTitle        *widget.Label
	roomDesc         *widget.Label
	scene            *canvas.Image
	roomCount        *widget.Label
	totalCount       *widget.Label
	hpLabel          *widget.Label
	hpBar            *statBar
	combatPanel      *fyne.Container
	combatName       *widget.Label
	combatHP         *widget.Label
	combatBar        *statBar
	fight            *fightState
	visited          map[string]bool
	endingBox        *fyne.Container
	flash            *canvas.Rectangle
	flashAnim        *fyne.Animation
	lastArc          string
	mapBox           *fyne.Container
	sceneMapBox      *fyne.Container
	sceneMapPanel    fyne.CanvasObject
	crewBar          *statBar
	crewLabel        *widget.Label
	groupLabel       *widget.Label
	exitBox          *fyne.Container
	playerBox        *fyne.Container
	itemBox          *fyne.Container
	itemPhotoBox     *fyne.Container
	itemPhotoScroll  *container.Scroll
	photoStrip       *fyne.Container
	npcBox           *fyne.Container
	inventoryBox     *fyne.Container
	questBox         *fyne.Container
	choiceTitle      *widget.Label
	choiceBox        *fyne.Container
	choicePopup      *widget.PopUp
	commandButtons   *fyne.Container
	journal          *container.AppTabs
	playArea         *fyne.Container
	scenePanel       fyne.CanvasObject
	detailPanel      fyne.CanvasObject
	chatScope        *widget.Select
	chatEntry        *widget.Entry
	chatLabel        *widget.Label
	storyText        *widget.RichText
	logLabel         *widget.Label
	chatScroll       *container.Scroll
	storyScroll      *container.Scroll
	logScroll        *container.Scroll
	messages         *container.AppTabs
	chatLines        []string
	storyLines       []storyEntry
	logLines         []string
}
```

- 接続・表示データ・各部品をまとめます。`state`は仲間の人数・オンラインの名前・グループ・招待を保持し、`stateUnavailable`はSTATE拡張を使えないサーバーかを記録します。

### 15-7-2 `main`— 起動

#### `main`

```go
func main() {
	application := app.NewWithID("io.github.anju0618.the-answer-protocol")
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	window := application.NewWindow("The Answer Protocol")
	window.Resize(fyne.NewSize(1280, 900))
	catalog, catalogErr := loadCatalog()
	ui := &gui{window: window, catalog: catalog, locale: "en"}
	ui.build()
	if catalogErr != nil {
		ui.addLog(ui.tr("Could not load display names: ", "表示名データを読み込めません: ") + catalogErr.Error())
	}
	window.SetOnClosed(func() {
		if ui.client != nil {
			ui.client.Close()
		}
	})
	window.Show()
	time.AfterFunc(300*time.Millisecond, func() { fyne.Do(window.RequestFocus) })
	application.Run()
}
```

- アプリとウインドウ、配色を作り、カタログを読み込みます。`build`で部品を作った後に`Show`と`Run`で表示・イベント処理を始めます。ウインドウを閉じると接続も閉じます。

- `app.NewWithID`で固定IDを指定し、発見した即死部屋の履歴をPreferencesで保存・復元します。`make run-client-gui`は、Fyneのパッケージ初期化前に`LANGUAGE`の空の候補を取り除いて起動します。直接`go run ./cmd/gui`で起動する場合は、このMakefileの補正は入りません。

### 15-7-3 チャット入力

#### `sendChat`

```go
func (ui *gui) sendChat() {
	message := strings.TrimSpace(ui.chatEntry.Text)
	if message == "" {
		return
	}
	if ui.send("CHAT " + ui.chatScope.Selected + " " + message) {
		ui.chatEntry.SetText("")
	}
}
```

- 空のメッセージは送信しません。選んだGLOBAL/ROOM/GROUPとメッセージを組み立て、キューへの送信に成功したときだけ入力欄を空にします。

### 15-7-4 コマンド送信

#### `send`

```go
func (ui *gui) send(line string) bool {
	if !ui.connected || ui.client == nil {
		ui.addLog(ui.tr("Connect before sending commands", "接続後にコマンドを送信してください"))
		return false
	}
	if err := ui.client.Send(line); err != nil {
		ui.addLog(ui.tr("Send failed: ", "送信失敗: ") + err.Error())
		return false
	}
	ui.addLog("> " + line)
	if strings.EqualFold(strings.TrimSpace(line), "QUIT") {
		ui.stopPolling()
		ui.quitButton.Disable()
	}
	return true
}
```

- 接続済みかを調べ、`protocolClient.Send`へ渡します。QUITを送れたら定期更新を止め、切断ボタンを無効にします。結果のboolは入力欄や選択画面を閉じてよいかの判定に使います。

#### `refresh`

```go
func (ui *gui) refresh(commands ...string) {
	for _, command := range commands {
		if command == "STATE" && ui.stateUnavailable {
			continue
		}
		ui.send(command)
	}
}
```

- 指定されたコマンドを順番に送信します。STATE拡張を使えないサーバーにはSTATEを送りません。

### 15-7-5 接続 `connect`

#### `connect`

```go
func (ui *gui) connect() {
	if ui.dialing || ui.client != nil {
		return
	}
	address := strings.TrimSpace(ui.hostEntry.Text)
	name := strings.TrimSpace(ui.nameEntry.Text)
	if address == "" || len(strings.Fields(name)) != 1 || strings.ContainsAny(name, "\r\n") {
		ui.showMessage(ui.tr("Connect", "接続"), ui.tr("Enter the server host:port and a name without spaces.", "サーバーのhost:portと空白のない名前を入力してください。"))
		return
	}
	ui.connectButton.Disable()
	ui.settingsButton.Disable()
	ui.languageSelect.Disable()
	ui.dialing = true
	ui.statusLabel.SetText(ui.tr("Connecting...", "接続中..."))
	ui.locale = "en"
	if ui.languageSelect.Selected == japaneseLanguageOption {
		ui.locale = "ja"
	}
	go func() {
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		fyne.Do(func() {
			ui.dialing = false
			if err != nil {
				ui.statusLabel.SetText(ui.tr("Connection failed", "接続失敗"))
				ui.connectButton.Enable()
				ui.settingsButton.Enable()
				ui.languageSelect.Enable()
				ui.addLog(ui.tr("Connection failed: ", "接続失敗: ") + err.Error())
				return
			}
			ui.client = newProtocolClient(conn)
			ui.window.Canvas().Unfocus()
			ui.addLog(ui.tr("Connected to server: ", "接続: ") + address)
			client := ui.client
			go ui.receive(client)
			if ui.locale == "ja" && !ui.sendHandshake("LANG ja") {
				return
			}
			ui.sendHandshake("CONNECT " + name)
		})
	}()
}
```

- 二重接続を防ぎ、空白なしの名前・接続先を検査します。接続・設定・言語の変更を無効にし、別goroutineで5秒のタイムアウト付きTCP接続を行います。画面更新は`fyne.Do`内で行います。日本語ならLANG ja、その後CONNECTを送ります。

#### `sendHandshake`

```go
func (ui *gui) sendHandshake(line string) bool {
	if err := ui.client.Send(line); err != nil {
		ui.addLog(ui.tr("Send failed: ", "送信失敗: ") + err.Error())
		ui.disconnect()
		return false
	}
	ui.addLog("> " + line)
	return true
}
```

- まだCONNECT応答が来ていない段階で送るための関数です。送信キューへ追加できなければ切断し、成功した場合はログに記録します。

### 15-7-6 受信と定期更新

#### `receive`

```go
func (ui *gui) receive(client *protocolClient) {
	for message := range client.incoming {
		message := message
		fyne.Do(func() {
			if ui.client == client {
				ui.handleMessage(message)
			}
		})
	}
}
```

- incomingチャネルを読み続けます。`fyne.Do`内で現在の接続と同じかを確認してから応答を処理するので、古い接続の通知が再接続後の画面を更新することを防ぎます。

#### `poll`

```go
func (ui *gui) poll(client *protocolClient, stop <-chan struct{}) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fyne.Do(func() {
				select {
				case <-stop:
					return
				default:
				}
				ui.pollOnce(client)
			})
		case <-stop:
			return
		case <-client.done:
			return
		}
	}
}
```

- 4秒ごとに、GUIのスレッド上で`pollOnce`を呼びます。停止チャネルか接続の終了が来れば戻り、tickerを止めます。

#### `pollOnce`

```go
func (ui *gui) pollOnce(client *protocolClient) {
	if ui.client == client && ui.connected && client.canPoll() {
		ui.refresh("LOOK", "WHO", "STATUS", "STATE")
	}
}
```

- 現在の接続が接続済みで、待っている応答が多すぎなければLOOK・WHO・STATUS・STATEを更新します。待機中のHP回復と仲間の人数も画面に反映できます。

#### `stopPolling`

```go
func (ui *gui) stopPolling() {
	if ui.pollStop != nil {
		close(ui.pollStop)
		ui.pollStop = nil
	}
}
```

- 停止チャネルを閉じてnilに戻します。nilなら何もしないため、二重に閉じません。

### 15-7-7 通知と応答の処理

#### `handleMessage`

```go
func (ui *gui) handleMessage(message serverMessage) {
	switch message.kind {
	case messageDisconnect:
		ui.addLog(ui.tr("Disconnected: ", "切断: ") + message.err.Error())
		ui.disconnect()
	case messageEvent:
		ui.addLog(message.line)
		ui.handleEvent(message.line)
	case messageGreeting, messageNotice:
		ui.addLog(message.line)
	case messageResponse:
		ui.addLog(message.line)
		ui.handleResponse(message.command, message.request, message.line)
	}
}
```

- 受信メッセージの種類で振り分けます。切断は状態をリセット、EVTはhandleEvent、コマンドの返事はhandleResponseへ渡します。

#### `handleEvent`

```go
func (ui *gui) handleEvent(line string) {
	parts := strings.SplitN(line, " ", 5)
	if len(parts) >= 5 && parts[0] == "EVT" && parts[2] == "CHAT" {
		ui.addChat("[" + parts[1] + "] " + parts[3] + ": " + parts[4])
		return
	}
	if strings.HasPrefix(line, "EVT PLAYER ") {

		kind, text, _ := strings.Cut(strings.TrimPrefix(line, "EVT PLAYER "), " ")
		ui.addStoryKind(storyKindOf(kind), text)
		switch kind {
		case "DEATH":
			ui.flashScene(flashDeath, 900*time.Millisecond)
			ui.showFight(nil)
			ui.refresh("LOOK", "INVENTORY", "STATUS", "STATE")
		case "QUEST":
			ui.refresh("STATUS", "QUESTS")
		case "ENDING":
			ui.refresh("INVENTORY")
		case "TEAM":
			ui.refresh("STATUS", "QUESTS", "STATE")
		}
		return
	}
	if strings.HasPrefix(line, "EVT STATS players=") {
		ui.setTotal(strings.TrimPrefix(line, "EVT STATS players="))
	}
	if strings.HasPrefix(line, "EVT ROOM PRESENCE ") {
		ui.send("LOOK")
	}
	if strings.HasPrefix(line, "EVT ROOM COMBAT ") {
		ui.addStoryKind(storyCombat, strings.TrimPrefix(line, "EVT ROOM COMBAT "))
		ui.send("LOOK")
		ui.send("STATUS")
		ui.refresh("STATE")
	}
	if strings.HasPrefix(line, "EVT GROUP INVITE ") {
		leader := strings.TrimPrefix(line, "EVT GROUP INVITE ")
		if !slices.Contains(ui.state.Invitations, leader) {
			ui.state.Invitations = append(ui.state.Invitations, leader)
		}
		ui.addStory(leader + ui.tr(" invited you. Open Group to accept.", " から招待されました。「グループ」から参加できます。"))
	}
	if strings.HasPrefix(line, "EVT GROUP JOIN ") || strings.HasPrefix(line, "EVT GROUP LEAVE ") {
		ui.refresh("STATE")
	}
}
```

- チャットはチャット欄へ、物語の通知は冒険欄へ追加します。死亡や戦闘などで関連する表示を再取得します。招待されたリーダー名は保存し、Groupの参加画面から選べるようにします。

#### `handleResponse`

```go
func (ui *gui) handleResponse(command, request, line string) {
	if strings.HasPrefix(line, "ERR ") {
		if command == "STATE" && strings.HasPrefix(line, "ERR 400 ") {
			ui.stateUnavailable = true
			ui.addLog(ui.tr("Crew and online names are unavailable on this server.", "このサーバーでは仲間の人数と全体の名前一覧を取得できません。"))
			return
		}
		ui.addStoryKind(storyError, line)
		if command == "QUIT" {
			ui.quitButton.Enable()
			if ui.pollStop == nil && ui.client != nil {
				ui.pollStop = make(chan struct{})
				go ui.poll(ui.client, ui.pollStop)
			}
		}
		if command == "LANG" {
			ui.switchLocale("en")
			ui.addLog("Japanese extension unavailable; continuing in English")
		}
		if command == "CONNECT" {
			ui.disconnect()
		}
		return
	}
	switch command {
	case "CONNECT":
		ui.connected = true
		ui.statusLabel.SetText(ui.tr("Playing as: ", "接続中: ") + ui.nameEntry.Text)
		ui.quitButton.Enable()
		ui.pollStop = make(chan struct{})
		go ui.poll(ui.client, ui.pollStop)
		for _, cmd := range []string{"LOOK", "INVENTORY", "STATUS", "QUESTS", "WHO", "STATE"} {
			ui.send(cmd)
		}
	case "LOOK":
		var view lookView
		if err := decodeOK(line, &view); err != nil || view.Room.ID == "" {
			ui.addLog(ui.tr("Could not parse LOOK: ", "LOOKを解析できません: ") + parseError(err))
			return
		}
		ui.showRoom(view)
	case "INVENTORY":
		var items []string
		if err := decodeOK(line, &items); err != nil {
			ui.addLog(ui.tr("Could not parse INVENTORY: ", "INVENTORYを解析できません: ") + err.Error())
			return
		}
		ui.showInventory(items)
	case "STATUS":
		var status statusView
		if err := decodeOK(line, &status); err != nil {
			ui.addLog(ui.tr("Could not parse STATUS: ", "STATUSを解析できません: ") + err.Error())
			return
		}
		ui.hpLabel.SetText(fmt.Sprintf("HP: %d/%d", status.HP, status.MaxHP))
		ui.hpBar.Set(status.HP, status.MaxHP)
		if status.Status != "healthy" {
			ui.hpLabel.SetText(ui.hpLabel.Text + " (" + ui.statusWord(status.Status) + ")")
		}
	case "STATE":
		var state stateView
		if err := decodeOK(line, &state); err != nil {
			ui.addLog(err.Error())
			return
		}
		ui.showState(state)
	case "WHO":
		if strings.HasPrefix(line, "OK players=") {
			ui.setTotal(strings.TrimPrefix(line, "OK players="))
		}
	case "MOVE":
		destination := strings.TrimPrefix(line, "OK room=")
		ui.addStory(ui.tr("Moved to: ", "移動: ") + destination)
		ui.flashScene(flashTravel, 450*time.Millisecond)
		ui.showFight(nil)
		ui.markVisited(destination)
		ui.refresh("LOOK", "STATUS", "QUESTS", "STATE")
		if room, ok := ui.catalog.gameOverRoom(destination); ok {
			ui.recordFatalRoom(destination)
			ui.showGameOver(destination, room)
		}
	case "TAKE", "DROP":
		id := strings.TrimPrefix(strings.TrimPrefix(line, "OK taken="), "OK dropped=")
		if command == "TAKE" {
			ui.addStory(ui.tr("Taken: ", "手に入れた: ") + ui.catalog.label("item", id, ui.locale))
		} else {
			ui.addStory(ui.tr("Dropped: ", "置いた: ") + ui.catalog.label("item", id, ui.locale))
		}
		ui.refresh("LOOK", "INVENTORY", "STATUS", "QUESTS", "STATE")
	case "DEFEND":
		if !ui.handleDefend(line) {
			ui.addStory(strings.TrimPrefix(line, "OK "))
		}
		ui.refresh("STATUS")
	case "ATTACK", "FLEE":
		handled := ui.handleAttack(request, line)
		if command == "FLEE" {
			handled = ui.handleFlee(line)
		}
		if !handled {
			ui.addStory(strings.TrimPrefix(line, "OK "))
		}
		ui.refresh("LOOK", "STATUS", "QUESTS", "STATE")
	case "TALK":
		words := strings.TrimPrefix(line, "OK ")
		ui.addStory(words)
		ui.showMessage("TALK", words)
		if line == "OK dead" {
			ui.refresh("LOOK", "STATUS", "STATE")
		}
	case "QUEST":
		words := strings.TrimPrefix(line, "OK ")
		var quest struct {
			QuestID     string `json:"quest_id"`
			Description string `json:"description"`
			Reward      int    `json:"reward"`
		}
		if json.Unmarshal([]byte(words), &quest) == nil && quest.Description != "" {
			words = ui.catalog.label("quest", quest.QuestID, ui.locale) + "\n" + quest.Description + fmt.Sprintf(ui.tr("\nMax HP +%d", "\n最大HP +%d"), quest.Reward)
		}
		ui.addStory(words)
		ui.showMessage("QUEST", words)
		ui.refresh("QUESTS")
	case "QUESTS":
		var quests []questView
		if err := decodeOK(line, &quests); err != nil {
			ui.addLog(ui.tr("Could not parse QUESTS: ", "QUESTSを解析できません: ") + err.Error())
			return
		}
		ui.showQuests(quests)
	case "GROUP":
		ui.addStory(line)
		if strings.HasPrefix(line, "OK group=") {
			ui.state.Group = strings.TrimPrefix(line, "OK group=")
			ui.state.Invitations = nil
		} else if strings.EqualFold(request, "GROUP LEAVE") {
			ui.state.Group = ""
		}
		ui.showState(ui.state)
		ui.refresh("STATE")
	case "QUIT":
		ui.addStory(ui.tr("Until our next journey.", "また旅を続けよう。"))
		ui.disconnect()
	}
}
```

- ERRなら失敗として扱います。STATEへのERR 400は拡張非対応として記録し、以後はRFCコマンドで続けます。CONNECT成功後に各表示を取得し、LOOKは部屋、INVENTORYは持ち物、STATUSはHP、STATEは仲間とグループ、QUESTSは依頼一覧へ反映します。MOVEなどの操作成功後は関連情報を再取得します。QUITの失敗では接続を継続し、定期更新を再開します。

### 15-7-8 表示の補助と切断

#### `disconnect`

```go
func (ui *gui) disconnect() {
	ui.stopPolling()
	if ui.client != nil {
		ui.client.Close()
		ui.client = nil
	}
	ui.connected = false
	ui.stateUnavailable = false
	ui.state = stateView{}
	ui.room = lookView{}
	ui.inventory = nil
	ui.quests = nil
	ui.clearChoices()
	ui.connectButton.Enable()
	ui.settingsButton.Enable()
	ui.quitButton.Disable()
	ui.languageSelect.Enable()
	ui.statusLabel.SetText(ui.tr("Not connected", "未接続"))
	ui.roomCount.SetText(ui.tr("Here: -", "部屋: - 人"))
	ui.totalCount.SetText(ui.tr("Online: -", "全体: - 人"))
	ui.hpLabel.SetText("HP: -")
	ui.hpBar.Set(0, 1)
	ui.showFight(nil)
	ui.visited, ui.lastArc = nil, ""
	ui.showState(stateView{})
	ui.scene.Resource = nil
	ui.scene.Image = loadArt("rooms", "unknown")
	ui.scene.Refresh()
	ui.showItemPhotos(nil)
	ui.showRoom(lookView{})
	ui.showInventory(nil)
	ui.showQuests(nil)
}
```

- 定期更新と接続を閉じ、部屋・持ち物・クエスト・仲間・グループ・選択画面を初期状態へ戻します。名前と接続先、チャット・冒険・ログの履歴は次の接続でも使えます。

#### `parseError`

```go
func parseError(err error) string {
	if err == nil {
		return "room.id is empty"
	}
	return err.Error()
}
```

- JSON解析に失敗した場合はそのエラーを返します。解析できても部屋IDが空の場合は、その状態を説明する文章を返します。

#### `setTotal`

```go
func (ui *gui) setTotal(value string) {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil && number >= 0 {
		ui.totalCount.SetText(fmt.Sprintf(ui.tr("Online: %d", "全体: %d 人"), number))
	}
}
```

- WHOやEVT STATSの数を読み、0以上なら人数表示を更新します。

#### `showMessage`

```go
func (ui *gui) showMessage(title, message string) {
	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord
	var popup *widget.PopUp
	closeButton := widget.NewButton(ui.tr("Close", "とじる"), func() {
		popup.Hide()
		ui.window.Canvas().Unfocus()
	})
	content := framed(title, container.NewBorder(nil, closeButton, nil, nil, container.NewVScroll(label)))
	popup = widget.NewModalPopUp(container.NewGridWrap(fyne.NewSize(520, 220), content), ui.window.Canvas())
	popup.Show()
	ui.window.Canvas().Focus(closeButton)
}
```

- TALKやQUESTなどの文章を、スクロールと閉じるボタンがあるモーダル画面へ表示します。

#### `showGameOver`

```go
func (ui *gui) showGameOver(roomID string, room catalogEntry) {
	scene := canvas.NewImageFromImage(loadArt("rooms", roomID))
	scene.FillMode = canvas.ImageFillContain
	scene.ScaleMode = canvas.ImageScaleSmooth
	scene.SetMinSize(fyne.NewSize(480, 288))
	title := widget.NewLabelWithStyle(room.Name.get(ui.locale), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapWord
	story := widget.NewLabel(room.Description.get(ui.locale))
	story.Wrapping = fyne.TextWrapWord
	var popup *widget.PopUp
	closeButton := widget.NewButton(ui.tr("Return to the Hall of the Fates", "運命の間へ戻る"), func() {
		popup.Hide()
		ui.window.Canvas().Unfocus()
	})
	content := framed(ui.tr("GAME OVER", "ゲームオーバー"),
		container.NewBorder(nil, closeButton, nil, nil, container.NewVScroll(textVBox(title, scene, story))))
	popup = widget.NewModalPopUp(container.NewGridWrap(fyne.NewSize(540, 520), content), ui.window.Canvas())
	popup.Show()
	ui.window.Canvas().Focus(closeButton)
}
```

- 即死部屋の背景・折り返し可能な名前・説明文をスクロール内へ表示し、戻るボタンを下部へ固定します。長い日本語の説明もボタンの上にはみ出しません。戻るボタンは画面を閉じる操作です。サーバー側の死亡処理と復活は既に行われています。

#### `addChat`

```go
func (ui *gui) addChat(line string) {
	ui.chatLines = appendLine(ui.chatLines, line)
	ui.chatLabel.SetText(strings.Join(ui.chatLines, "\n"))
	ui.chatScroll.ScrollToBottom()
}
```

- 時刻付きの履歴を更新し、チャット欄の末尾へスクロールします。

#### `addStory`

```go
func (ui *gui) addStory(line string) { ui.addStoryKind(storyPlain, line) }
```

- `addStory`は`ui_story.go`にあります。種類付きの履歴をRichTextへ反映し、冒険欄の末尾へスクロールします。死亡・クエスト・戦闘・エンディング・仲間・ヒント等で色を変えます。

#### `addLog`

```go
func (ui *gui) addLog(line string) {
	ui.logLines = appendLine(ui.logLines, line)
	ui.logLabel.SetText(strings.Join(ui.logLines, "\n"))
	ui.logScroll.ScrollToBottom()
}
```

- 時刻付きの履歴を更新し、通信ログ欄の末尾へスクロールします。

#### `appendLine`

```go
func appendLine(lines []string, line string) []string {
	lines = append(lines, time.Now().Format("15:04:05")+"  "+line)
	if len(lines) > 300 {
		return lines[len(lines)-300:]
	}
	return lines
}
```

- 現在時刻と文章を追加し、300行を超えたら古い行を取り除きます。

## 15-8 `ui_layout.go`(414行)— 画面とウインドウサイズ

### `build`

```go
func (ui *gui) build() {
	ui.hostEntry = widget.NewEntry()
	ui.hostEntry.SetText("127.0.0.1:4242")
	ui.hostEntry.SetPlaceHolder("host:port")
	ui.nameEntry = widget.NewEntry()
	ui.nameEntry.SetPlaceHolder(ui.tr("player name", "プレイヤー名"))
	ui.languageSelect = widget.NewSelect([]string{"English", japaneseLanguageOption}, nil)
	ui.languageSelect.SetSelected(ui.tr("English", japaneseLanguageOption))
	ui.connectButton = widget.NewButton(ui.tr("Connect", "接続"), func() {
		if strings.TrimSpace(ui.nameEntry.Text) == "" {
			ui.showConnectionSettings()
		} else {
			ui.connect()
		}
	})
	ui.settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), ui.showConnectionSettings)
	ui.quitButton = widget.NewButton(ui.tr("Disconnect", "切断"), func() { ui.send("QUIT") })
	ui.quitButton.Importance = widget.LowImportance
	ui.connectButton.Importance = widget.HighImportance
	ui.quitButton.Disable()
	ui.statusLabel = dimLabel(ui.tr("Not connected", "未接続"))
	ui.roomCount = dimLabel(ui.tr("Here: -", "部屋: - 人"))
	ui.totalCount = dimLabel(ui.tr("Online: -", "全体: - 人"))
	ui.hpLabel = boldLabel("HP: -")
	ui.crewLabel = boldLabel(ui.tr("Crew: -", "仲間: - 人"))
	ui.hpBar, ui.crewBar = newStatBar(), newStatBar()
	ui.groupLabel = dimLabel(ui.tr("Group: -", "グループ: -"))

	ui.roomTitle = widget.NewLabelWithStyle(ui.tr("Your journey", "冒険の旅"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ui.roomTitle.Truncation = fyne.TextTruncateEllipsis
	ui.roomDesc = widget.NewLabel(ui.tr("Connect to begin your journey.", "接続して冒険を始めましょう。"))
	ui.roomDesc.Wrapping = fyne.TextWrapWord
	ui.exitBox = textVBox()
	ui.playerBox = textVBox()
	ui.itemBox = textVBox()
	ui.npcBox = textVBox()
	ui.inventoryBox = textVBox()
	ui.questBox = textVBox()
	ui.endingBox = newEndingBox()
	ui.choiceTitle = widget.NewLabel("")
	ui.choiceBox = textVBox()
	ui.itemPhotoBox = container.NewHBox()
	ui.itemPhotoScroll = container.NewHScroll(ui.itemPhotoBox)
	ui.itemPhotoScroll.Hide()
	ui.scene = sceneImage()
	photos := container.NewBorder(nil, nil,
		widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { ui.scrollItemPhotos(-1) }),
		widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { ui.scrollItemPhotos(1) }),
		ui.itemPhotoScroll,
	)
	ui.photoStrip = photos
	ui.photoStrip.Hide()
	ui.flash = newFlashLayer()
	ui.sceneMapBox = container.New(&mapLayout{})
	ui.sceneMapPanel = framed(ui.tr("Map", "地図"), ui.mapView(ui.sceneMapBox))
	ui.sceneMapPanel.Hide()
	visualLayout := &sceneVisualLayout{photos: ui.itemPhotoBox}
	sceneVisual := container.NewScroll(container.New(visualLayout, ui.scene, ui.flash, ui.photoStrip, ui.sceneMapPanel))
	sceneVisual.Direction = container.ScrollNone
	sceneFrame := canvas.NewRectangle(ink)
	sceneFrame.StrokeColor = gold
	sceneFrame.StrokeWidth = 1
	visual := container.NewStack(sceneFrame, sceneVisual)
	ui.scenePanel = container.NewBorder(ui.roomTitle, nil, nil, nil,
		container.New(&sceneLayout{}, visual, container.NewVScroll(ui.roomDesc)))

	surroundings := textVBox(
		journalSection(ui.tr("Paths", "移動先"), ui.exitBox),
		journalSection(ui.tr("People & creatures", "人物・生きもの"), ui.npcBox),
		journalSection(ui.tr("Items here", "落ちている道具"), ui.itemBox),
		journalSection(ui.tr("Players here", "この部屋のプレイヤー"), ui.playerBox),
	)
	mapTab := container.NewTabItem(ui.tr("Map", "地図"), ui.buildMapTab())
	ui.journal = container.NewAppTabs(
		container.NewTabItem(ui.tr("Around", "まわり"), container.NewVScroll(surroundings)),
		container.NewTabItem(ui.tr("Inventory", "持ち物"), container.NewVScroll(ui.inventoryBox)),
		container.NewTabItem(ui.tr("Quests", "クエスト"), container.NewVScroll(textVBox(ui.questBox, ui.endingBox))),
		mapTab,
	)
	journal := ui.journal
	visualLayout.onMapVisibility = func(visible bool) {
		if visible && len(journal.Items) == 4 {
			if journal.Selected() == mapTab {
				journal.SelectIndex(0)
			}
			journal.Remove(mapTab)
		} else if !visible && len(journal.Items) == 3 {
			journal.Append(mapTab)
		}
	}
	ui.journal.OnSelected = func(item *container.TabItem) {
		if !ui.connected {
			return
		}
		switch item {
		case ui.journal.Items[1]:
			ui.send("INVENTORY")
		case ui.journal.Items[2]:
			ui.send("QUESTS")
		}
	}
	ui.commandButtons = container.NewGridWithColumns(3,
		ui.commandButton(ui.tr("Refresh", "更新"), func() { ui.refresh("LOOK", "INVENTORY", "STATUS", "QUESTS", "WHO", "STATE") }),
		ui.commandButton(ui.tr("Group", "グループ"), func() { ui.chooseAction("GROUP") }),
		ui.commandButton(ui.tr("Chat", "チャット"), ui.focusChat),
	)
	ui.buildCombatPanel()
	ui.detailPanel = container.NewBorder(ui.combatPanel, ui.commandButtons, nil, nil, ui.journal)
	ui.playArea = container.New(&adventureLayout{}, ui.scenePanel, ui.detailPanel)

	ui.chatScope = widget.NewSelect([]string{"GLOBAL", "ROOM", "GROUP"}, nil)
	ui.chatScope.SetSelected("ROOM")
	ui.chatEntry = widget.NewEntry()
	ui.chatEntry.SetPlaceHolder(ui.tr("Write a message…", "メッセージを入力…"))
	ui.chatEntry.OnSubmitted = func(string) { ui.sendChat() }
	ui.chatLabel = widget.NewLabel("")
	ui.chatLabel.Wrapping = fyne.TextWrapWord
	ui.chatScroll = container.NewVScroll(ui.chatLabel)
	chatInput := container.NewBorder(nil, nil, ui.chatScope, widget.NewButton(ui.tr("Send", "送信"), ui.sendChat), ui.chatEntry)
	chatPane := container.NewBorder(nil, chatInput, nil, nil, ui.chatScroll)
	ui.logLabel = widget.NewLabel("")
	ui.logLabel.Wrapping = fyne.TextWrapWord
	ui.logScroll = container.NewVScroll(ui.logLabel)
	ui.storyText = newStoryText()
	ui.storyScroll = container.NewVScroll(ui.storyText)
	ui.messages = container.NewAppTabs(
		container.NewTabItem(ui.tr("Adventure", "ぼうけん"), ui.storyScroll),
		container.NewTabItem(ui.tr("Chat", "チャット"), chatPane),
		container.NewTabItem(ui.tr("Log", "ログ"), ui.logScroll),
	)
	brand := canvas.NewText("THE ANSWER PROTOCOL", gold)
	brand.TextSize = 17
	brand.TextStyle.Bold = true
	header := container.New(&headerLayout{}, brand,
		container.NewHBox(ui.languageSelect, ui.settingsButton, ui.connectButton, ui.quitButton))
	stats := container.New(&statsLayout{}, container.NewStack(ui.hpBar.box, ui.hpLabel), container.NewStack(ui.crewBar.box, ui.crewLabel), ui.groupLabel, ui.roomCount, ui.totalCount, ui.statusLabel)
	ui.window.SetContent(container.NewPadded(container.New(&screenLayout{}, header, stats, ui.playArea, ui.messages)))
	ui.languageSelect.OnChanged = func(selection string) {
		if ui.client != nil || ui.connected || ui.dialing {
			return
		}
		if selection == japaneseLanguageOption {
			ui.switchLocale("ja")
		} else {
			ui.switchLocale("en")
		}
	}
	ui.showRoom(lookView{})
	ui.showInventory(nil)
	ui.showQuests(nil)
	ui.showEndings()
	for _, saved := range []struct {
		lines []string
		label *widget.Label
	}{
		{ui.chatLines, ui.chatLabel}, {ui.logLines, ui.logLabel},
	} {
		if len(saved.lines) > 0 {
			saved.label.SetText(strings.Join(saved.lines, "\n"))
		}
	}
	if len(ui.storyLines) > 0 {
		ui.storyText.Segments = storySegments(ui.storyLines)
		ui.storyText.Refresh()
	} else {
		ui.addStory(ui.tr("The gods of Greece await you.", "ギリシアの神々があなたを待っている。"))
	}
}
```

- 接続・状態・部屋の絵・まわり／持ち物／クエスト／地図・冒険／チャット／ログを作ります。HP・仲間のバー、戦闘パネル、エンディング一覧もここで組み立てます。画像はスクロール操作を無効にしたScrollの内側でクリップし、アイテムの絵は左右のボタンで送れます。「逃げる」は敵の「戦う」と同じ行にあります。

### `compactLabel`

```go
func compactLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Truncation = fyne.TextTruncateEllipsis
	return label
}
```

- 長い状態表示は省略記号で収めます。文字の高さはテーマから計算します。

### `dimLabel` と `boldLabel`— 画面の強弱

```go
// dimLabel is for secondary info (counts, status): same size, quieter colour.
func dimLabel(text string) *widget.Label {
	label := compactLabel(text)
	label.Importance = widget.LowImportance
	return label
}

// boldLabel is for the numbers the player watches most (HP, crew).
func boldLabel(text string) *widget.Label {
	label := compactLabel(text)
	label.TextStyle.Bold = true
	return label
}
```

- 画面上部の状態行(HP・仲間・グループ・人数・接続状態)は、以前はすべて同じ見た目でした。**重要なものほど目立つ**ように、2種類の補助関数で見た目を分けています。
- `dimLabel`:**控えめな文字色**のラベル。`Importance`(重要度)を `LowImportance` にすると、Fyneはテーマの「うすい文字色」(`ColorNamePlaceHolder`、15-3)で描きます。グループ・部屋の人数・全体の人数・接続状態に使います。
- `boldLabel`:**太字**のラベル。`TextStyle` の `Bold` を `true` にします。**いちばん見たい数字**であるHPと仲間の人数に使います。
- どちらも中身は `compactLabel`(上)で作ってから、1つだけ設定を足しています。返す型は同じ `*widget.Label` なので、あとで `SetText` で文字を変える処理(`main.go`)はそのまま使えます。
- 同じ考え方で、ヘッダーのボタンも**接続は強調**(`HighImportance`、色付き)、**切断は目立たなく**(`LowImportance`、枠だけ)にしています(`build` の中)。ゲーム中に押す機会の多いボタンが、見つけやすくなります。

### `showConnectionSettings`

```go
func (ui *gui) showConnectionSettings() {
	if ui.client != nil || ui.dialing {
		return
	}
	var popup *widget.PopUp
	closePopup := func() { popup.Hide(); ui.window.Canvas().Unfocus() }
	connect := widget.NewButton(ui.tr("Connect", "接続"), func() { closePopup(); ui.connect() })
	ui.nameEntry.OnSubmitted = func(string) { closePopup(); ui.connect() }
	ui.hostEntry.OnSubmitted = func(string) { ui.window.Canvas().Focus(ui.nameEntry) }
	content := container.NewVBox(
		widget.NewLabelWithStyle(ui.tr("Connection settings", "接続設定"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(ui.tr("Server", "接続先")), ui.hostEntry,
		widget.NewLabel(ui.tr("Player name", "名前")), ui.nameEntry,
		container.NewGridWithColumns(2, widget.NewButton(ui.tr("Cancel", "戻る"), closePopup), connect),
	)
	popup = widget.NewModalPopUp(container.NewPadded(content), ui.window.Canvas())
	popup.Resize(fyne.NewSize(min(440, ui.window.Canvas().Size().Width-30), content.MinSize().Height+12))
	popup.Show()
	ui.window.Canvas().Focus(ui.nameEntry)
}
```

- 接続先とプレイヤー名を入力する画面です。接続中や接続処理中には開きません。

### `screenLayout`

```go
type screenLayout struct{}
```

- 見出し・状態・冒険領域・メッセージ欄を縦に配置するレイアウトです。

### `screenLayout.MinSize`

```go
func (*screenLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(3) * theme.Padding()
	for _, object := range objects {
		height += object.MinSize().Height
	}
	return fyne.NewSize(560, max(680, height))
}
```

- 560×680を基準に、各部品に必要な高さを合計します。文字サイズや戦闘パネルの表示で必要な最小高さが増えます。

### `screenLayout.Layout`

```go
func (*screenLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	headerHeight := heightForWidth(objects[0], size.Width)
	statsHeight := heightForWidth(objects[1], size.Width)
	messagesHeight := min(float32(230), size.Height*0.26)
	if size.Width < 960 {
		messagesHeight = min(float32(170), size.Height*0.22)
	}
	messagesHeight = max(objects[3].MinSize().Height, min(messagesHeight, size.Height-headerHeight-statsHeight-objects[2].MinSize().Height-3*gap))
	playHeight := max(float32(0), size.Height-headerHeight-statsHeight-messagesHeight-3*gap)
	y := float32(0)
	for index, height := range []float32{headerHeight, statsHeight, playHeight, messagesHeight} {
		objects[index].Move(fyne.NewPos(0, y))
		objects[index].Resize(fyne.NewSize(size.Width, height))
		y += height + gap
	}
}
```

- 見出しと状態の実際の高さを測り、残りを絵と一覧へ割り当てます。メッセージ欄の希望高さは広い画面で最大230、狭い画面で最大170とし、冒険領域の最小高さを確保します。

### `heightForWidth`

```go
func heightForWidth(object fyne.CanvasObject, width float32) float32 {
	object.Resize(fyne.NewSize(width, object.MinSize().Height))
	return object.MinSize().Height
}
```

- 先に表示幅を設定してから最小高さを測ります。日本語などの折り返しで増えた高さを次の行の位置へ反映できます。

### `headerLayout`

```go
type headerLayout struct{ stacked bool }
```

- 題名と接続操作を並べるレイアウトです。幅が足りるかをstackedへ保存します。

### `headerLayout.MinSize`

```go
func (l *headerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	brand, controls := objects[0].MinSize(), objects[1].MinSize()
	if l.stacked {
		return fyne.NewSize(max(brand.Width, controls.Width), brand.Height+controls.Height+theme.Padding())
	}
	return fyne.NewSize(brand.Width+controls.Width+theme.Padding(), max(brand.Height, controls.Height))
}
```

- 横並びなら幅を合計し、上下なら高さを合計します。固定の文字高さを使いません。

### `headerLayout.Layout`

```go
func (l *headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	brand, controls := objects[0].MinSize(), objects[1].MinSize()
	l.stacked = brand.Width+controls.Width+theme.Padding() > size.Width
	objects[0].Resize(brand)
	objects[1].Resize(controls)
	if l.stacked {
		objects[0].Move(fyne.NewPos(0, 0))
		objects[1].Move(fyne.NewPos(max(0, size.Width-controls.Width), brand.Height+theme.Padding()))
		return
	}
	height := max(brand.Height, controls.Height)
	objects[0].Move(fyne.NewPos(0, (height-brand.Height)/2))
	objects[1].Move(fyne.NewPos(size.Width-controls.Width, (height-controls.Height)/2))
}
```

- 題名と接続操作が同じ行に収まらなければ、接続操作を下へ置きます。言語選択や大きな文字でも題名と重なりません。

### `statsLayout`

```go
type statsLayout struct{ columns int }
```

- HP・仲間・グループ・部屋人数・全体人数・接続状態を並べます。

### `statsLayout.MinSize`

```go
func (l *statsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return layout.NewGridLayoutWithColumns(max(3, l.columns)).MinSize(objects)
}
```

- 現在の列数で必要な高さを計算します。HPと仲間の背景バーも含めて測ります。

### `statsLayout.Layout`

```go
func (l *statsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.columns = 6
	if size.Width < 1000 {
		l.columns = 3
	}
	layout.NewGridLayoutWithColumns(l.columns).Layout(objects, size)
}
```

- 幅1000以上は6列、それ未満は3列2行にします。状態表示の部品は作り直しません。

### `textVBox`

```go
func textVBox(objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(&textVBoxLayout{}, objects...)
}
```

- 幅を設定してから折り返しの高さを測る縦並びのコンテナを作ります。一覧のカードや文章を含むダイアログで使います。

### `textVBoxLayout`

```go
type textVBoxLayout struct{}
```

- 文章の幅と高さを順に計算する縦並びのレイアウトです。

### `textVBoxLayout.MinSize`

```go
func (*textVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return layout.NewVBoxLayout().MinSize(objects)
}
```

- 表示中の部品の最小高さと間隔を合計します。

### `textVBoxLayout.Layout`

```go
func (*textVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, object := range objects {
		if !object.Visible() {
			continue
		}
		height := heightForWidth(object, size.Width)
		object.Move(fyne.NewPos(0, y))
		object.Resize(fyne.NewSize(size.Width, height))
		y += height + theme.Padding()
	}
}
```

- 各部品へ幅を渡し、折り返し後の高さを測ってから次の部品を置きます。長い名前・クエスト説明・祝福文が次の行やボタンと重なることを防ぎます。

### `adventureLayout`

```go
type adventureLayout struct{ wide bool }
```

- 絵と詳細欄を横並びまたは縦並びにするレイアウトです。

### `adventureLayout.MinSize`

```go
func (l *adventureLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := max(objects[0].MinSize().Height, objects[1].MinSize().Height)
	if !l.wide {
		height = objects[0].MinSize().Height + objects[1].MinSize().Height + theme.Padding()
	}
	return fyne.NewSize(0, max(400, height))
}
```

- 横並びでは高い方、縦並びでは両方の高さと間隔を使います。戦闘パネルの表示も最小サイズへ反映します。

### `adventureLayout.Layout`

```go
func (l *adventureLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	l.wide = size.Width >= 960
	if l.wide {
		sideWidth := max(float32(330), objects[1].MinSize().Width)
		objects[0].Move(fyne.NewPos(0, 0))
		objects[0].Resize(fyne.NewSize(size.Width-sideWidth-gap, size.Height))
		objects[1].Move(fyne.NewPos(size.Width-sideWidth, 0))
		objects[1].Resize(fyne.NewSize(sideWidth, size.Height))
		return
	}
	sceneHeight := min(size.Width*0.6+80, max(100, size.Height-190))
	sceneHeight = min(max(objects[0].MinSize().Height, sceneHeight), max(0, size.Height-objects[1].MinSize().Height-gap))
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, sceneHeight))
	objects[1].Move(fyne.NewPos(0, sceneHeight+gap))
	objects[1].Resize(fyne.NewSize(size.Width, max(0, size.Height-sceneHeight-gap)))
}
```

- 幅960以上なら左右へ配置し、詳細欄の幅は330または実際の最小幅の大きい方です。狭い画面では上下にし、一覧・戦闘パネルの最小高さを確保して絵の高さを調整します。

### `sceneLayout`

```go
type sceneLayout struct{}
```

- 部屋の絵とスクロールできる説明文を上下へ置きます。

### `sceneLayout.MinSize`

```go
func (*sceneLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	art, description := objects[0].MinSize(), objects[1].MinSize()
	return fyne.NewSize(max(art.Width, description.Width), art.Height+max(description.Height, textLineHeight())+theme.Padding())
}
```

- 絵の最小高さに説明文1行と間隔を加えます。

### `sceneLayout.Layout`

```go
func (*sceneLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	descriptionHeight := min(float32(62), max(textLineHeight(), size.Height*0.2))
	descriptionHeight = min(descriptionHeight, max(0, size.Height-objects[0].MinSize().Height-theme.Padding()))
	artHeight := max(float32(0), size.Height-descriptionHeight-theme.Padding())
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, artHeight))
	objects[1].Move(fyne.NewPos(0, artHeight+theme.Padding()))
	objects[1].Resize(fyne.NewSize(size.Width, descriptionHeight))
}
```

- 説明文は少なくとも1行が読める高さを確保し、残りを絵へ使います。複数行の説明はスクロールできます。

### `textLineHeight`

```go
func textLineHeight() float32 {
	textTheme := theme.Current()
	style := fyne.TextStyle{}
	size, _ := fyne.CurrentApp().Driver().RenderedTextSize("Ag国", textTheme.Size(theme.SizeNameText), style, textTheme.Font(style))
	return size.Height + 2*textTheme.Size(theme.SizeNameInnerPadding)
}
```

- 現在のテーマのフォントと文字サイズを使って、日本語を含む1行の高さを測ります。上下の内側余白も加えます。

### `sceneVisualLayout`

```go
type sceneVisualLayout struct {
	photos          *fyne.Container
	onMapVisibility func(bool)
}
```

- 背景・フラッシュ・アイテムの絵を、部屋の絵の枠内へ配置します。

### `sceneVisualLayout.MinSize`

```go
func (*sceneVisualLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(100, 100) }
```

- 絵の表示領域の最小サイズを100×100にします。

### `sceneVisualLayout.Layout`

```go
func (l *sceneVisualLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	inside := fyne.NewSize(max(0, size.Width-2*gap), max(0, size.Height-2*gap))
	for _, layer := range objects[:2] {
		layer.Move(fyne.NewPos(gap, gap))
		layer.Resize(inside)
	}
	mapPanel := objects[3]
	renderedWidth := min(inside.Width, inside.Height*float32(artWidth)/float32(artHeight))
	leftSpace := (inside.Width - renderedWidth) / 2
	mapSize := mapPanel.MinSize()
	if leftSpace >= mapSize.Width+2*gap && inside.Height >= mapSize.Height {
		mapPanel.Show()
		mapPanel.Move(fyne.NewPos(gap+(leftSpace-mapSize.Width)/2, gap+(inside.Height-mapSize.Height)/2))
		mapPanel.Resize(mapSize)
	} else {
		mapPanel.Hide()
	}
	if l.onMapVisibility != nil {
		l.onMapVisibility(mapPanel.Visible())
	}
	if objects[2].Visible() {
		thumbnail := fyne.NewSquareSize(min(104, max(48, inside.Height*0.3)))
		for _, photo := range l.photos.Objects {
			card := photo.(*fyne.Container)
			if card.MinSize() != thumbnail {
				card.Layout = layout.NewGridWrapLayout(thumbnail)
				card.Resize(thumbnail)
			}
		}
		strip := fyne.NewSize(min(290, inside.Width), objects[2].MinSize().Height)
		strip.Height = min(strip.Height, inside.Height)
		objects[2].Move(fyne.NewPos(size.Width-gap-strip.Width, size.Height-gap-strip.Height))
		objects[2].Resize(strip)
	}
}
```

- 背景とフラッシュを枠から余白分だけ離します。アイテムは右下に置き、絵の高さに応じて48～104の正方形へ縮小・拡大します。アイテム欄の幅と高さも枠の内側へ収めます。

- 横に広い画面では、縦横比を保って表示した絵の左に余白が生まれます。その余白に地図と余白分の幅・高さが収まる場合だけ、地図を常時表示します。絵の表示範囲を狭めたり、絵の上に地図を重ねたりはしません。

- 左の地図を表示するときは右の地図タブを取り除き、表示できない幅に戻すと同じタブを追加します。地図タブを選択中なら「まわり」へ戻し、持ち物・クエストを選択中ならそのまま保ちます。

## 15-9 `ui_actions.go`(94行)— マウス操作とグループ

### `commandButton`

```go
func (ui *gui) commandButton(text string, action func()) *widget.Button {
	return widget.NewButton(text, func() {
		action()
		if ui.window.Canvas().Focused() != ui.chatEntry {
			ui.window.Canvas().Unfocus()
		}
	})
}
```

- クリックで渡された処理を実行します。チャットへフォーカスを移す操作以外では、入力欄のフォーカスを外します。

### `focusChat`

```go
func (ui *gui) focusChat() {
	ui.messages.SelectIndex(1)
	ui.window.Canvas().Focus(ui.chatEntry)
}
```

- チャットのタブを開いて入力欄へフォーカスを移します。送信は画面のボタンでもできます。

### `chooseAction`

```go
func (ui *gui) chooseAction(action string) {
	if !ui.connected {
		ui.addLog(ui.tr("Connect before sending commands", "接続後に操作してください"))
		return
	}
	ui.clearChoices()
	var choices []menuChoice
	title := ui.tr("Choose a target", "対象を選ぶ")
	switch action {
	case "GROUP":
		title = ui.tr("Player group", "プレイヤーのグループ")
		choices = []menuChoice{
			{label: ui.tr("Create a group", "グループを作る"), command: "GROUP CREATE"},
			{label: ui.tr("Invite a player", "プレイヤーを招待"), action: "INVITE"},
			{label: ui.tr("Accept an invitation", "招待を受ける"), action: "JOIN"},
			{label: ui.tr("Leave the group", "グループから抜ける"), command: "GROUP LEAVE"},
		}
	case "INVITE":
		title = ui.tr("Invite a player", "招待するプレイヤー")
		players := ui.state.Players
		if ui.stateUnavailable || players == nil {
			players = ui.room.Players
		}
		for _, name := range players {
			if name != strings.TrimSpace(ui.nameEntry.Text) {
				choices = append(choices, menuChoice{label: name, command: "GROUP INVITE " + name})
			}
		}
	case "JOIN":
		title = ui.tr("Accept an invitation", "参加するグループのリーダー")
		for _, leader := range ui.state.Invitations {
			choices = append(choices, menuChoice{label: leader, command: "GROUP JOIN " + leader})
		}
	}
	ui.choices = choices
	ui.choiceTitle.SetText(title)
	var rows []fyne.CanvasObject
	for index, choice := range choices {
		rows = append(rows, ui.commandButton(choice.label, func() { ui.runChoice(index) }))
	}
	ui.setJournalRows(ui.choiceBox, rows, ui.tr("No available targets.", "選べる対象がありません。"))
	cancel := widget.NewButton(ui.tr("Back", "戻る"), ui.clearChoices)
	content := container.NewBorder(ui.choiceTitle, cancel, nil, nil, container.NewVScroll(ui.choiceBox))
	ui.choicePopup = widget.NewModalPopUp(container.NewPadded(content), ui.window.Canvas())
	ui.choicePopup.Resize(fyne.NewSize(min(480, ui.window.Canvas().Size().Width-40), min(440, ui.window.Canvas().Size().Height-60)))
	ui.choicePopup.Show()
}
```

- グループ作成・招待・参加・退出の選択画面を開きます。招待する相手はSTATEのオンライン一覧、参加先は受け取った招待から選びます。拡張非対応ではLOOKの同室プレイヤーとEVTの招待を利用します。選択肢は9件で切らず、全件をスクロールで選べます。

### `runChoice`

```go
func (ui *gui) runChoice(index int) {
	if index < 0 || index >= len(ui.choices) {
		return
	}
	choice := ui.choices[index]
	if choice.action != "" {
		ui.chooseAction(choice.action)
		return
	}
	if ui.send(choice.command) {
		ui.clearChoices()
	}
}
```

- 範囲内の選択肢を確認し、次の選択画面を開くかコマンドを送信します。送信キューへ追加できた場合だけ選択画面を閉じます。

### `clearChoices`

```go
func (ui *gui) clearChoices() {
	if ui.choicePopup != nil {
		ui.choicePopup.Hide()
		ui.choicePopup = nil
		ui.window.Canvas().Unfocus()
	}
	ui.choices = nil
}
```

- 開いている選択画面を閉じ、古い選択肢を取り除きます。部屋変更と切断の際にも使います。

## 15-10 `ui_journal.go`(274行)— まわり・持ち物・クエスト

敵と未知のNPCでは「戦う／逃げる」を同じ行に表示し、会話・依頼とは別の行にします。文章を含むカードは`textVBox`で高さを測ります。アイテムの左右ボタンは現在のカード1枚と余白分だけスクロールします。

### `journalSection`

```go
func journalSection(title string, content fyne.CanvasObject) fyne.CanvasObject {
	heading := canvas.NewText(title, gold)
	heading.TextSize = 14
	heading.TextStyle.Bold = true
	return textVBox(container.NewPadded(heading), content)
}
```

- 一覧のまとまりに金色の小見出しを付けます。

### `journalCard`

```go
func journalCard(content fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(navy)
	background.CornerRadius = 4
	return container.NewStack(background, container.NewPadded(content))
}
```

- 紺色の背景と内側の余白を作り、項目同士を見分けやすくします。

### `journalName`

```go
func journalName(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord
	return label
}
```

- 名前は太字にし、幅に収まらなければ折り返します。

### `journalRow`

```go
func journalRow(name string, action *widget.Button) fyne.CanvasObject {
	if action == nil {
		return journalCard(journalName(name))
	}
	return journalCard(container.NewBorder(nil, nil, nil, action, journalName(name)))
}
```

- 名前と操作ボタンを1項目にまとめます。自分の名前など、操作ボタンが不要な行も作れます。

### `showRoom`

```go
func (ui *gui) showRoom(view lookView) {
	previous := ui.room
	if previous.Room.ID != view.Room.ID {
		ui.clearChoices()
	}
	ui.room = view
	if view.Room.ID != "" {
		if ui.visited == nil {
			ui.visited = map[string]bool{}
		}
		ui.visited[view.Room.ID] = true
	}
	defer ui.refreshMap()
	if ui.scene.Image == nil || previous.Room.ID != view.Room.ID || !slices.Equal(previous.NPCs, view.NPCs) || !slices.Equal(previous.Defeated, view.Defeated) {
		ui.scene.Resource = nil
		ui.scene.Image = composeScene(view.Room.ID, view.NPCs, view.Defeated)
		ui.scene.Refresh()
	}
	if previous.Room.ID != view.Room.ID || !slices.Equal(previous.Items, view.Items) {
		ui.showItemPhotos(view.Items)
	}
	if view.Room.ID == "" {
		ui.roomTitle.SetText(ui.tr("Your journey", "冒険の旅"))
		ui.roomDesc.SetText(ui.tr("Connect to begin your journey.", "接続して冒険を始めましょう。"))
	} else {
		ui.roomTitle.SetText(view.Room.Name)
		ui.roomDesc.SetText(view.Room.Description)
		ui.roomCount.SetText(fmt.Sprintf(ui.tr("Here: %d", "部屋: %d 人"), len(view.Players)))
	}
	if view.Room.ID != "" && reflect.DeepEqual(previous, view) {
		return
	}
	directions := make([]string, 0, len(view.Room.Exits))
	for direction := range view.Room.Exits {
		directions = append(directions, direction)
	}
	sort.Strings(directions)
	var exits, players, items, npcs []fyne.CanvasObject
	for _, direction := range directions {
		exits = append(exits, journalRow(ui.exitLabel(direction, view.Room.Exits[direction]),
			ui.commandButton(ui.tr("Go", "進む"), func() { ui.send("MOVE " + direction) })))
	}
	for _, id := range view.NPCs {
		buttons := []fyne.CanvasObject{ui.commandButton(ui.tr("Talk", "話す"), func() { ui.send("TALK " + id) })}
		known := ui.catalog != nil && ui.catalog.NPCs[id].Role != ""
		if !known || ui.catalog.hasQuest(id) {
			buttons = append(buttons, ui.commandButton(ui.tr("Quest", "依頼"), func() { ui.send("QUEST " + id) }))
		}
		actions := container.NewGridWithColumns(len(buttons), buttons...)
		content := textVBox(journalName(ui.catalog.label("npc", id, ui.locale)), actions)
		switch {
		case slices.Contains(view.Defeated, id):
			// Nothing more to do to someone already beaten.
		case !known || ui.catalog.NPCs[id].Role == "enemy":
			content.Add(container.NewGridWithColumns(2,
				ui.commandButton(ui.tr("Attack", "戦う"), func() { ui.send("ATTACK " + id) }),
				ui.commandButton(ui.tr("Flee", "逃げる"), func() { ui.send("FLEE") }),
			))
		default:
			// A person can be attacked too, but the result is rarely good.
			content.Add(ui.commandButton(ui.tr("Attack", "攻撃する"), func() { ui.send("ATTACK " + id) }))
		}
		npcs = append(npcs, journalCard(content))
	}
	for _, id := range view.Items {
		items = append(items, ui.itemRow(id, "TAKE", ui.tr("Take", "取る")))
	}
	for _, name := range view.Players {
		var invite *widget.Button
		if name != strings.TrimSpace(ui.nameEntry.Text) {
			invite = ui.commandButton(ui.tr("Invite", "招待"), func() { ui.send("GROUP INVITE " + name) })
		}
		players = append(players, journalRow(name, invite))
	}
	ui.setJournalRows(ui.exitBox, exits, ui.tr("No paths from here.", "移動先がありません。"))
	ui.setJournalRows(ui.npcBox, npcs, ui.tr("No one to talk to here.", "話しかける相手はいません。"))
	ui.setJournalRows(ui.itemBox, items, ui.tr("No items here.", "道具は落ちていません。"))
	ui.setJournalRows(ui.playerBox, players, ui.tr("No players here.", "プレイヤーはいません。"))
	ui.journal.Items[0].Content.Refresh()
}
```

NPCの行の攻撃ボタンは、`switch` で3通りに分かれます(上から順に、最初に当てはまったもの)。
- **倒した敵**(`slices.Contains(view.Defeated, id)`):**何も出しません**。もう戦えないので、攻撃・逃げるのボタンは要りません(話す・依頼のボタンは別の行に出ています)。
- **敵**(未知のNPCも含む):「戦う」と「逃げる」の2つを横に並べます。
- **それ以外(敵ではない人)**:「**攻撃する**」ボタンを1つだけ出します。サーバーは敵以外への `ATTACK` も受け付ける(8-5のケース3・4)ので、ボタンから押せます。ただし結果は、一般人なら殺して自分が死に、強い人なら一撃で殺されるので、**「戦う」ではなく「攻撃する」**と違う言葉にして、普通の戦闘と区別しています。「逃げる」は出しません(一般人は戦闘状態にならず、逃げる相手がいないため)。

一覧の差し替えが終わったら、「まわり」全体のスクロール領域もRefreshします。子の一覧だけを更新すると、既に表示している親に古い高さが残り、接続直後の項目が重なることがありました。親まで更新することで、タブ切替やサイズ変更をしなくても初回のLOOKから正しい高さになります。

- 部屋IDが変われば古い選択画面を閉じます。部屋・NPC・倒されたNPCの一覧が変わったときに背景を作り直し、道具が変わったときにアイテムの絵を更新します。倒された敵の専用画像は`assets/npcs_defeated/<NPC ID>.png`から読み込みます。同じ部屋で敵を倒した直後にも表示が切り替わります。移動先・NPC・道具・プレイヤーを操作ボタン付きで表示します。カタログで分かるNPCには対応する依頼・戦闘だけを表示し、未知のNPCでは各操作を残します。同じLOOKなら一覧やスクロール位置を保ちます。

### `hasQuest`

```go
func (catalog *worldCatalog) hasQuest(npcID string) bool {
	if catalog == nil {
		return false
	}
	for _, quest := range catalog.Quests {
		if quest.GiverNPCID == npcID {
			return true
		}
	}
	return false
}
```

- カタログのgiver_npc_idを調べ、依頼を持つ人物かを判定します。

### `itemRow`

```go
func (ui *gui) itemRow(id, command, text string) fyne.CanvasObject {
	var picture fyne.CanvasObject
	if data, err := itemPhotoAssets.ReadFile("assets/items/" + id + ".png"); err == nil {
		image := canvas.NewImageFromResource(fyne.NewStaticResource(id+".png", data))
		image.FillMode = canvas.ImageFillContain
		picture = container.NewGridWrap(fyne.NewSize(44, 44), image)
	}
	button := ui.commandButton(text, func() { ui.send(command + " " + id) })
	info := textVBox(journalName(ui.catalog.label("item", id, ui.locale)))
	for _, line := range ui.itemEffectLines(id) {
		info.Add(line)
	}
	buttons := fyne.CanvasObject(button)
	if description := ui.itemDescription(id); description != "" {
		// The explanation is hidden until the player taps the info button, so the list stays short.
		explanation := widget.NewLabel(description)
		explanation.Wrapping = fyne.TextWrapWord
		explanation.Hide()
		info.Add(explanation)
		infoButton := widget.NewButtonWithIcon("", theme.InfoIcon(), nil)
		infoButton.OnTapped = func() {
			if explanation.Visible() {
				explanation.Hide()
			} else {
				explanation.Show()
			}
			ui.journal.Refresh()
		}
		buttons = container.NewHBox(infoButton, button)
	}
	return journalCard(container.NewBorder(nil, nil, picture, buttons, info))
}
```

- 道具の小さな絵・表示名・**効果**・取る／置くボタン・**解説の開閉ボタン**をまとめます。表示は名前、送信はIDを使います。「まわり」の道具(`TAKE`)と持ち物(`DROP`)の**両方がこの関数**で作られます。
- `info`:アイテム名を先頭にした**縦並びの入れ物**です。ここに、効果の行と解説を足していきます。
- `for _, line := range ui.itemEffectLines(id) { info.Add(line) }`:**効果を1行ずつ足します**。良い効果は緑、悪い効果は赤の小さな文字です(15-17)。効果の無いアイテムは、何も足されません。
- `buttons := fyne.CanvasObject(button)`:右側に置く部品を、まず「取る/置くボタンだけ」にします。`fyne.CanvasObject(button)` は**型の変換**で、`*widget.Button` を、どの部品も入れられる `fyne.CanvasObject` 型の変数として持ちます(あとで、横並びの入れ物に差し替えられるようにするため)。
- **解説がある**アイテム(`itemDescription` が空でない)の場合:
  - `explanation`:解説の文章のラベル。`TextWrapWord` で折り返し、**最初は隠します**(`Hide()`)。解説は長いので、普段は隠して一覧を短く保ちます。
  - `info.Add(explanation)`:隠したまま、アイテムの欄に入れておきます。
  - `infoButton`:**ⓘ(情報)のアイコンのボタン**(`theme.InfoIcon()`)。文字は付けず、**アイコンだけ**なので狭い画面でも場所を取りません。
  - `infoButton.OnTapped = func() { ... }`:押されたときの処理。解説が見えていれば隠し、隠れていれば見せます(`Visible()` で今の状態を調べる)。そのあと `ui.journal.Refresh()` で、**タブ全体を描き直して**、高さが変わった分を反映します(文章が出ると、カードが縦に伸びるため)。
  - ボタンを作ったあとで `OnTapped` を設定しているのは、押されたときの処理の中で `explanation` を使うためです(作る前には存在しない)。
  - `buttons = container.NewHBox(infoButton, button)`:右側を**ⓘ と 取る/置く の横並び**に差し替えます。
- 最後に、左に絵、右にボタン、中央に `info`(名前・効果・解説)を置いたカードを返します。

### `itemDescription`

```go
// itemDescription is the item's explanation from data/world.json in the current language (empty if unknown).
func (ui *gui) itemDescription(id string) string {
	if ui.catalog == nil {
		return ""
	}
	return ui.catalog.Items[id].Description.get(ui.locale)
}
```

- アイテムの**解説文**を、世界データ(`data/world.json`)の `description` から、**今の言語**で返します。
- 世界データが読めていない(`catalog == nil`)、またはそのアイテムが載っていないときは、`""` を返します。`itemRow` は、`""` ならⓘボタンを出しません。
- `.get(ui.locale)` は、15-2の `localizedName.get` で、**その言語の文章が無ければ英語**を返します。

### `showInventory`

```go
func (ui *gui) showInventory(ids []string) {
	ui.inventory = append([]string(nil), ids...)
	var rows []fyne.CanvasObject
	if len(ids) > 0 {
		rows = append(rows, widget.NewLabel(fmt.Sprintf(ui.tr("%d items carried", "持ち物 %d 個"), len(ids))))
	}
	for _, id := range ids {
		rows = append(rows, ui.itemRow(id, "DROP", ui.tr("Drop", "置く")))
	}
	ui.setJournalRows(ui.inventoryBox, rows, ui.tr("Your bag is empty. Pick up items in Around.", "持ち物はありません。「まわり」から道具を拾えます。"))
	ui.showEndings()
}
```

- 持ち物のコピーを保持し、個数・小さな絵・名前・置くボタンを表示します。空なら取得方法を案内します。

### `showQuests`

```go
func (ui *gui) showQuests(quests []questView) {
	ui.quests = append([]questView(nil), quests...)
	var rows []fyne.CanvasObject
	for _, quest := range quests {
		name := journalName(ui.catalog.label("quest", quest.QuestID, ui.locale))
		status := widget.NewLabel(ui.statusWord(quest.Status) + "  ·  " + quest.Progress)
		content := textVBox(name, status)
		if ui.catalog != nil {
			if description := ui.catalog.Quests[quest.QuestID].Description.get(ui.locale); description != "" {
				label := widget.NewLabel(description)
				label.Wrapping = fyne.TextWrapWord
				content.Add(label)
			}
		}
		current, goal, _ := strings.Cut(quest.Progress, "/")
		progress, progressErr := strconv.Atoi(current)
		target, targetErr := strconv.Atoi(goal)
		if progressErr == nil && targetErr == nil && target > 0 {
			bar := widget.NewProgressBar()
			bar.SetValue(float64(progress) / float64(target))
			content.Add(bar)
		}
		rows = append(rows, journalCard(content))
	}
	ui.setJournalRows(ui.questBox, rows, ui.tr("No quests yet. Ask people in Around for a quest.", "受けているクエストはありません。「まわり」の人物から依頼を受けられます。"))
}
```

- 依頼名・状態・進捗・説明文を表示します。進捗がa/bの数値なら進捗バーも表示します。クエストの報酬やサーバー側の進行処理は変更しません。

### `setJournalRows`

```go
func (ui *gui) setJournalRows(box *fyne.Container, rows []fyne.CanvasObject, empty string) {
	if len(rows) == 0 {
		label := widget.NewLabel(empty)
		label.Wrapping = fyne.TextWrapWord
		rows = []fyne.CanvasObject{container.NewPadded(label)}
	}
	box.Objects = rows
	box.Refresh()
}
```

- 項目が無い場合は、その欄に合う空状態の文章を表示します。

### `showItemPhotos`

```go
func (ui *gui) showItemPhotos(ids []string) {
	photos := make([]fyne.CanvasObject, 0, len(ids))
	for _, id := range ids {
		if card := itemPhotoCard(id); card != nil {
			photos = append(photos, card)
		}
	}
	ui.itemPhotoBox.Objects = photos
	ui.itemPhotoBox.Refresh()
	if len(photos) == 0 {
		ui.itemPhotoScroll.Hide()
		ui.photoStrip.Hide()
	} else {
		ui.itemPhotoScroll.Show()
		ui.photoStrip.Show()
	}
	ui.itemPhotoScroll.ScrollToOffset(fyne.Position{})
}
```

- 現在の部屋にあるアイテムの絵を並べます。絵が無ければ左右ボタンを含む欄を隠します。

### `scrollItemPhotos`

```go
func (ui *gui) scrollItemPhotos(direction float32) {
	if len(ui.itemPhotoBox.Objects) == 0 {
		return
	}
	step := ui.itemPhotoBox.Objects[0].MinSize().Width + theme.Padding()
	offset := ui.itemPhotoScroll.Offset
	offset.X = max(0, offset.X+direction*step)
	ui.itemPhotoScroll.ScrollToOffset(offset)
}
```

- directionが1なら右、-1なら左へ、現在のカード1枚の幅と余白分だけ送ります。0未満にはしません。狭い画面でカードが縮小しても1枚ずつ操作できます。

### `showState`

```go
func (ui *gui) showState(state stateView) {
	ui.state = state
	if state.CrewInitialized {
		ui.crewLabel.SetText(fmt.Sprintf(ui.tr("Crew: %d", "仲間: %d 人"), state.Crew))
		ui.crewBar.Set(state.Crew, max(startingCrew, state.Crew))
	} else {
		ui.crewLabel.SetText(ui.tr("Crew: -", "仲間: - 人"))
		ui.crewBar.Set(0, startingCrew)
	}
	group := "-"
	if state.Group != "" {
		group = state.Group
	}
	ui.groupLabel.SetText(ui.tr("Group: ", "グループ: ") + group)
}
```

- STATE応答を保存し、仲間の人数とグループを更新します。crew_initializedがfalseなら未取得を表す「-」、trueで0なら「0人」です。

## 15-11 `ui_bars.go`(65行)— HPと仲間のバー

**役割**:HPと仲間の人数の**ラベルの後ろに、残量を色付きの帯で描く**部品です。ラベルは「HP: 80/100」と正確な数字を出し、帯は「だいたいどのくらい残っているか」を一目で見せます。

```go
type statBar struct {
	box    *fyne.Container
	fill   *canvas.Rectangle
	layout *barLayout
}

type barLayout struct{ ratio float32 }
```

- `statBar`:1本のバーを表す構造体です。3つのフィールドを持ちます。
  - `box`:バーの**入れ物**(`*fyne.Container`)。背景と塗りを重ねて入れます。画面に置くのはこれです。
  - `fill`:**塗り(残量)の四角形**。`*canvas.Rectangle` は、色を付けて描ける四角形の部品です。
  - `layout`:この入れ物の**並べ方**。下の `barLayout` で、塗りの幅を決めます。
- `barLayout struct{ ratio float32 }`
  - 中身は `ratio`(0〜1の小数)だけの**とても小さな構造体**です。「塗りが全体の何割か」を覚えます。
  - `float32` は32ビットの小数です。Fyneの座標や大きさは `float32` で表すので、それに合わせています。

```go
func (*barLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(0, 0) }

// Layout stretches the background over the whole area and the fill over `ratio` of its width.
func (l *barLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
	objects[1].Move(fyne.NewPos(0, 0))
	objects[1].Resize(fyne.NewSize(size.Width*l.ratio, size.Height))
}
```

- この2つのメソッドを持つ型は、Fyneの**レイアウト**(`fyne.Layout`)として使えます。Goには「`implements` と書く」仕組みが無く、**必要なメソッドを持っていれば自動的にその型として扱われる**(これを**インターフェース**と呼びます)。
- `MinSize`:この入れ物の**最小の大きさ**。`(0, 0)` と答えるので、バーは**大きさを要求しません**(ラベルの大きさに合わせて広がります)。
  - 引数 `[]fyne.CanvasObject` は使わないので、名前を付けずに型だけ書いています。`(*barLayout)` のようにレシーバにも名前が無いのは、使わないからです。
- `Layout`:入れ物の中の部品を**どこにどの大きさで置くか**を決めます。`objects` は中の部品のスライスで、順番は「0番が背景、1番が塗り」です。
  - `objects[0].Move(0,0)` と `Resize(size)`:背景を、**全体いっぱい**に敷きます。
  - `objects[1].Resize(fyne.NewSize(size.Width*l.ratio, size.Height))`:塗りは**幅だけ `ratio` 倍**にします。`ratio` が0.5なら半分の幅で、左から半分だけ塗られます。高さは全体と同じです。

```go
func newStatBar() *statBar {
	back := canvas.NewRectangle(color.NRGBA{R: 28, G: 38, B: 60, A: 255})
	fill := canvas.NewRectangle(barColor(1))
	layout := &barLayout{}
	return &statBar{box: container.New(layout, back, fill), fill: fill, layout: layout}
}
```

- バーを1本作ります。
- `back`:**背景**の四角形(暗い紺色)。`color.NRGBA{R, G, B, A}` は、赤・緑・青・透明度(0〜255)で色を作ります。`A: 255` は不透明です。
- `fill`:**塗り**の四角形。最初は満タン(`barColor(1)`)の色にします。
- `container.New(layout, back, fill)`:自作のレイアウトで、背景と塗りを入れた入れ物を作ります。**引数の順番が重なり順**で、あとに書いた `fill` が上に描かれます。

```go
func (b *statBar) Set(current, max int) {
	b.layout.ratio = barRatio(current, max)
	b.fill.FillColor = barColor(b.layout.ratio)
	b.box.Refresh()
}
```

- `Set`:**バーの量を更新**する。`current`(今の値)と `max`(最大値)を渡します。
- 割合を計算して `layout.ratio` に入れ、割合に応じた色を `fill.FillColor` に入れ、`Refresh()` で**描き直しを頼みます**。Fyneは「値を変えただけでは画面が変わらず、`Refresh` で初めて描き直す」仕組みです。
- 引数の名前 `max` は、組み込み関数 `max` と同じ名前ですが、**この関数の中だけ**で変数として使えます(組み込み関数を隠してしまうので、中では `max()` が呼べなくなります)。

```go
func barRatio(current, max int) float32 {
	if max <= 0 || current <= 0 {
		return 0
	}
	if current >= max {
		return 1
	}
	return float32(current) / float32(max)
}
```

- `current / max` を、**0〜1の範囲に収めて**返します。
- 最大が0以下、または今が0以下なら **0(空)**。今が最大以上なら **1(満タン)**。0で割るエラーも、はみ出しもここで防ぎます。
- `float32(current) / float32(max)`:整数どうしの割り算は小数にならない(切り捨て)ので、**先に小数に変換してから**割ります。

```go
func barColor(ratio float32) color.NRGBA {
	switch {
	case ratio > 0.5:
		return color.NRGBA{R: 46, G: 125, B: 80, A: 255}
	case ratio > 0.25:
		return color.NRGBA{R: 176, G: 128, B: 32, A: 255}
	}
	return color.NRGBA{R: 170, G: 48, B: 48, A: 255}
}
```

- **残量に応じた色**を返します。半分より多ければ**緑**、4分の1より多ければ**琥珀色**、それ以下は**赤**です。信号のように、減るほど危険に見えます。
- `switch { case 条件: ... }`:条件式だけを並べる `switch` です。上から順に見て、最初に当てはまったものが実行されます。

> **使われる場所**:HPと仲間のバー(`ui_layout.go`、15-8)と、戦闘パネルの敵のHPバー(`ui_combat.go`、15-14)で、同じ部品を使い回しています。`ui_bars_test.go` が、割合の計算(0〜1に収まる:最大0、マイナス、最大超えなど)と、**残量で色が変わる**ことを確認します。

## 15-12 `ui_story.go`(128行)— 色付きの冒険ログ

**役割**:「ぼうけん」タブの文章を、**種類ごとに色分けして**表示します。死亡は赤、クエストは金、戦闘は橙、のように、一目で何が起きたか分かるようにします。

```go
type storyKind int

const (
	storyPlain storyKind = iota
	storyDeath
	storyQuest
	storyCombat
	storyEnding
	storyTeam
	storyError
	storyHint
)
```

- `storyKind`:**文章の種類**を表す型です。実体は整数(`int`)ですが、「ただの数字」と区別するために別の名前の型にします。
- `const ( ... )` の中の `iota`:Goの**連番を作る仕組み**です。最初の `storyPlain = iota` が0、次の `storyDeath` が1、…と**自動で1ずつ増えます**(2行目以降は `= iota` を省略できる)。
- 種類は、**普通・死亡・クエスト・戦闘・エンディング・仲間・エラー・ヒント**の8つ。

```go
type storyEntry struct {
	kind storyKind
	time string
	text string
}
```

- 1行分のログ。**種類・時刻・本文**を持ちます。`time` は「15:04:05」の形の文字列です。

```go
const (
	colorStoryDeath  fyne.ThemeColorName = "storyDeath"
	colorStoryQuest  fyne.ThemeColorName = "storyQuest"
	colorStoryCombat fyne.ThemeColorName = "storyCombat"
	colorStoryEnding fyne.ThemeColorName = "storyEnding"
	colorStoryTeam   fyne.ThemeColorName = "storyTeam"
	colorStoryHint   fyne.ThemeColorName = "storyHint"
)
```

- **テーマの色の名前**を自分で作っています。`fyne.ThemeColorName` は「色の名前」を表す文字列の型で、`"foreground"`(文字色)や `"background"` が標準で用意されています。
- `RichText`(装飾付きの文章)は、**色を名前で指定**します。そこで、「`storyDeath` という名前なら赤」とテーマに教えます(`retro.go` の `retroTheme.Color`、15-3)。
- 色は直接値を書かず名前にしておくので、**テーマを変えるだけで色を差し替えられます**。

```go
func storyColorName(kind storyKind) fyne.ThemeColorName {
	switch kind {
	case storyDeath, storyError:
		return colorStoryDeath
	case storyQuest:
		return colorStoryQuest
	case storyCombat:
		return colorStoryCombat
	case storyEnding:
		return colorStoryEnding
	case storyTeam:
		return colorStoryTeam
	case storyHint:
		return colorStoryHint
	}
	return "foreground"
}
```

- **種類から色の名前**を返します。`case storyDeath, storyError:` のように、複数の値を1つの `case` にまとめると「どちらでも」の意味です。**死亡とエラーは同じ赤**です。
- どれにも当てはまらない(普通)なら、標準の文字色 `"foreground"` を返します。

```go
func storyColor(name fyne.ThemeColorName) (color.Color, bool) {
	switch name {
	case colorStoryDeath:
		return color.NRGBA{R: 232, G: 96, B: 96, A: 255}, true
	case colorStoryQuest:
		return gold, true
	...
	}
	return nil, false
}
```

- **色の名前から、実際の色**を返します。戻り値は「色」と「この名前を知っていたか(`bool`)」の2つです。
- 値は、死亡=赤、クエスト=金(`retro.go` の `gold`)、戦闘=橙、エンディング=紫、仲間=緑、ヒント=水色。
- `retroTheme.Color`(15-3)が最初にこれを呼び、**知っている名前ならその色**、知らなければ標準の処理に任せます。
- 緑(仲間)と赤(死亡)は、15-17のアイテム効果の「良い/悪い」の色にも使い回されます。

```go
func storyKindOf(word string) storyKind {
	switch word {
	case "DEATH":
		return storyDeath
	case "QUEST":
		return storyQuest
	case "ENDING":
		return storyEnding
	case "TEAM":
		return storyTeam
	case "HINT":
		return storyHint
	}
	return storyPlain
}
```

- サーバーが送る `EVT PLAYER <種類> <本文>` の**種類の単語**(`DEATH` `QUEST` など)を、`storyKind` に変換します(11-1の通知と対応)。知らない単語は普通扱いです。

```go
func newStoryText() *widget.RichText {
	text := widget.NewRichText()
	text.Wrapping = fyne.TextWrapWord
	return text
}

func storySegments(entries []storyEntry) []widget.RichTextSegment {
	segments := make([]widget.RichTextSegment, 0, len(entries))
	for _, entry := range entries {
		segments = append(segments, &widget.TextSegment{
			Text: entry.time + "  " + entry.text,
			Style: widget.RichTextStyle{
				ColorName: storyColorName(entry.kind),
				Inline:    false,
			},
		})
	}
	return segments
}
```

- `newStoryText`:装飾付きの文章の部品(`RichText`)を作ります。`TextWrapWord` は、**単語の切れ目で折り返す**設定です。
- `storySegments`:ログの全行を、**色付きの段落のリスト**に変換します。
  - `RichText` は「**セグメント**(装飾の単位)」を並べて表示します。1行のログ=1セグメントです。
  - `make([]T, 0, len(entries))`:長さ0で、**容量だけ先に確保した**スライスを作ります(あとで `append` するとき、メモリの取り直しが起きません)。
  - 本文は「時刻 + 空白2つ + 本文」。`ColorName` に種類の色を、`Inline: false` で**1行ごとに改行**します。
  - `&widget.TextSegment{...}`:構造体を作って、そのポインタ(`&`)を入れます。スライスの要素はインターフェース型なので、ポインタで入れます。

```go
func (ui *gui) addStory(line string) { ui.addStoryKind(storyPlain, line) }

func (ui *gui) addStoryKind(kind storyKind, line string) {
	ui.storyLines = append(ui.storyLines, storyEntry{kind: kind, time: time.Now().Format("15:04:05"), text: strings.TrimSpace(line)})
	if len(ui.storyLines) > 300 {
		ui.storyLines = ui.storyLines[len(ui.storyLines)-300:]
	}
	ui.storyText.Segments = storySegments(ui.storyLines)
	ui.storyText.Refresh()
	ui.storyScroll.ScrollToBottom()
}
```

- `addStory`:**普通の文章**を足す、`addStoryKind` の短縮形です。
- `addStoryKind`:ログに**1行足して、画面に反映**します。
  - `time.Now().Format("15:04:05")`:今の時刻を文字列にします。Goの書式は**特殊で、「2006年1月2日 15時4分5秒」という特定の日時を見本にして書きます**(`15:04:05` は「時:分:秒」を24時間表記にする意味)。
  - `strings.TrimSpace(line)`:前後の空白や改行を取り除きます。
  - `if len(...) > 300`:**300行を超えたら古いほうを捨てます**(`スライス[n:]` は「n番目から最後まで」)。ログが無限に増えて重くなるのを防ぎます。
  - 最後に、**全行を作り直して画面に反映**し(`Refresh`)、**一番下までスクロール**します(`ScrollToBottom`)。新しい文章がいつも見えます。
- 画面を作る側(`ui_layout.go`)が、再接続や言語切替のときにも `ui.storyLines` から作り直すので、**ログは接続を切っても残ります**(15-8)。

## 15-13 `ui_effects.go`(37行)— 画面のフラッシュ

**役割**:ダメージを受けたとき、移動したとき、死んだときに、**部屋の絵を一瞬だけ色付きで光らせ、すっと消す**演出です。

```go
var (
	flashClear  = color.NRGBA{}
	flashHurt   = color.NRGBA{R: 200, G: 30, B: 30, A: 110}
	flashDeath  = color.NRGBA{R: 120, G: 0, B: 0, A: 200}
	flashTravel = color.NRGBA{A: 230}
)
```

- 4つの色を**変数**で用意します(`var ( ... )` でまとめて宣言)。
  - `flashClear`:`color.NRGBA{}`。**フィールドを何も書かないと全部0**(ゼロ値)になるので、これは**完全な透明**です。
  - `flashHurt`:**赤・半透明**(`A: 110`)。ダメージを受けたとき。
  - `flashDeath`:**暗い赤・かなり濃い**(`A: 200`)。死んだとき。
  - `flashTravel`:**黒・ほぼ不透明**(`A: 230`、RGBは0)。部屋を移動するとき、暗転してから明るくなる効果。

```go
func newFlashLayer() *canvas.Rectangle {
	return canvas.NewRectangle(flashClear)
}
```

- **透明な四角形**を1枚作ります。部屋の絵の**上に重ねて**置き(`ui_layout.go`の `sceneVisualLayout`、15-8)、普段は透明なので何も見えません。演出のときだけ、この四角形の色を変えます。

```go
func (ui *gui) flashScene(from color.NRGBA, duration time.Duration) {
	if ui.flash == nil {
		return
	}
	if ui.flashAnim != nil {
		ui.flashAnim.Stop()
	}
	ui.flashAnim = canvas.NewColorRGBAAnimation(from, flashClear, duration, func(c color.Color) {
		ui.flash.FillColor = c
		ui.flash.Refresh()
	})
	ui.flashAnim.Curve = fyne.AnimationEaseOut
	ui.flashAnim.Start()
}
```

- **`from` の色から透明まで、`duration` の時間をかけて薄くしていく**アニメーションを始めます。
- `if ui.flash == nil { return }`:四角形がまだ作られていなければ何もしません(画面を作る前に呼ばれても落ちないように)。
- `if ui.flashAnim != nil { ui.flashAnim.Stop() }`:**前のアニメーションが動いていたら止めます**。連続でダメージを受けたときに、2つのアニメーションが同じ四角形を取り合わないようにするためです。
- `canvas.NewColorRGBAAnimation(開始の色, 終わりの色, 時間, 色が変わるたびに呼ぶ関数)`:色を少しずつ変えるアニメーションを作ります。
  - 最後の引数は**関数**です。色が変わるたびに呼ばれ、四角形の色を新しい色 `c` に変えて `Refresh` します。**関数を引数として渡せる**のが、Goの便利なところです。
- `Curve = fyne.AnimationEaseOut`:**最初は速く、終わりに向けてゆっくり**変化させます(自然な消え方)。
- `Start()`:アニメーションを始めます。Fyneが**別の流れで**時間を進めてくれるので、この関数はすぐ戻ります。

> **使われる場所**:敵の反撃を受けたとき(`flashHurt`、15-14の戦闘)、部屋を移動したとき(`flashTravel`)、死んだとき(`flashDeath`)に呼ばれます。`ui_effects_test.go` が、①画面を作る前(`flash` がまだ `nil`)に `flashScene` を呼んでも**落ちない**こと、②呼ぶとアニメーションが**始まる**こと、③戦闘と移動の応答を処理しても落ちないことを確認します。

## 15-14 `ui_combat.go`(146行)— 戦闘パネルと戦闘の結果表示

**役割**:戦っている間だけ現れる「戦闘パネル」(敵の名前・HPバー・戦う/構える/逃げるのボタン)を作り、`ATTACK` `FLEE` `DEFEND` の**応答を文章にして冒険ログへ出す**ファイルです。

```go
type attackView struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

type fightState struct {
	npcID string
	hp    int
	maxHP int
}
```

- `attackView`:`ATTACK` の応答JSON(8-5の `combatResult`)を**受け取る型**。サーバーと同じ4つのフィールドです。
- `fightState`:**今の戦闘**を覚える構造体(敵のID・敵の今のHP・敵の最大HP)。戦っていないときは `ui.fight` が `nil` です(コメントのとおり)。

### `buildCombatPanel`— パネルを作る

```go
func (ui *gui) buildCombatPanel() {
	ui.combatName = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ui.combatName.Truncation = fyne.TextTruncateEllipsis
	ui.combatHP = compactLabel("")
	ui.combatBar = newStatBar()
	attack := ui.commandButton(ui.tr("Attack", "戦う"), func() {
		if ui.fight != nil {
			ui.send("ATTACK " + ui.fight.npcID)
		}
	})
	flee := ui.commandButton(ui.tr("Flee", "逃げる"), func() { ui.send("FLEE") })
	defend := ui.commandButton(ui.tr("Brace", "構える"), func() { ui.send("DEFEND") })
```

- 敵の名前のラベル(太字、長いと `…` で省略)、HPのラベル、HPのバー(15-11の `newStatBar`)を作ります。
- 3つのボタンを作ります。`ui.commandButton`(15-9)はボタンを作る補助で、**押したあと入力欄からフォーカスを外す**処理が付いています。
  - 「戦う」:**今戦っている敵**(`ui.fight.npcID`)に `ATTACK` を送ります。`if ui.fight != nil` で、戦闘中でなければ何もしません。
  - 「逃げる」「構える」:`FLEE` / `DEFEND` を送るだけです(対象はサーバーが覚えているので、IDは不要)。
- ボタンの文字は `ui.tr("英語", "日本語")`(15-4)で、**今の言語のほう**が選ばれます。
- `func() { ... }` は**無名関数**で、ボタンが押されたときに呼ばれる処理をその場で書いています。

```go
	frame := canvas.NewRectangle(ink)
	frame.StrokeColor, _ = storyColor(colorStoryCombat)
	frame.StrokeWidth = 2
	inside := textVBox(
		ui.combatName,
		container.NewStack(ui.combatBar.box, ui.combatHP),
		container.NewGridWithColumns(3, attack, defend, flee),
	)
	ui.combatPanel = container.NewStack(frame, container.NewPadded(inside))
	ui.combatPanel.Hide()
}
```

- `frame`:パネルの**枠**。背景は暗い色(`ink`)、枠線は**戦闘の色(橙)**を2ピクセルで引きます。
- `frame.StrokeColor, _ = storyColor(...)`:`storyColor` は「色」と「知っていたか」の2つを返す(15-12)ので、**2つ目は使わず `_` で捨てます**。
- `inside`:パネルの中身を**縦に並べます**(`textVBox`、15-8)。上から、敵の名前 → HPバーとHPの文字 → 3つのボタン(`NewGridWithColumns(3, ...)` は**3列の格子**)。
  - `container.NewStack(bar.box, label)`:HPバーの**上に**HPの文字を**重ねます**(15-11で説明した「ラベルの後ろにバー」)。
- `container.NewStack(frame, container.NewPadded(inside))`:枠の上に、余白を付けた中身を重ねます。
- `ui.combatPanel.Hide()`:**最初は隠しておきます**。戦闘が始まったら `showFight` で見せます。

### `showFight`— 表示の更新

```go
func (ui *gui) showFight(fight *fightState) {
	ui.fight = fight
	if fight == nil {
		ui.combatPanel.Hide()
		return
	}
	ui.combatName.SetText(ui.tr("In combat: ", "戦闘中: ") + ui.catalog.label("npc", fight.npcID, ui.locale))
	ui.combatHP.SetText(fmt.Sprintf(ui.tr("Enemy HP: %d/%d", "敵のHP: %d/%d"), max(fight.hp, 0), fight.maxHP))
	ui.combatBar.Set(fight.hp, fight.maxHP)
	ui.combatPanel.Show()
}
```

- `fight` が `nil`(戦闘終了)ならパネルを**隠して終わり**。
- 戦闘中なら、敵の名前(`catalog.label`、15-2)・HPの文字・HPのバーを更新して、パネルを見せます。
- `max(fight.hp, 0)`:HPが**マイナスの表示にならない**ようにします。

### `handleAttack`— `ATTACK` の応答を文章にする

```go
func (ui *gui) handleAttack(request, line string) bool {
	var result attackView
	if decodeOK(line, &result) != nil || result.Status == "" {
		return false
	}
	npcID := strings.TrimSpace(strings.TrimPrefix(request, "ATTACK"))
	enemy := ui.catalog.label("npc", npcID, ui.locale)
	switch result.Status {
```

- `request`:自分が送った行(`"ATTACK npc.harpy"`)。応答のJSONには**敵のIDが入っていない**ので、**送った行から敵のIDを取り出します**(`TrimPrefix` で先頭の `ATTACK` を除き、`TrimSpace` で空白を除く)。
- `decodeOK`(15-2)で応答のJSONを `attackView` に変換します。変換できない、または `status` が空なら **`false` を返し**、呼んだ側が**生の文章をそのまま出す**ようにします(予想外の応答でも画面が壊れません)。
- `enemy`:敵の**表示名**(その言語の名前)。以降の文章に入れます。
- `switch result.Status` で、**応答の `status` ごとに分岐**します。サーバーの7つの状態(8-5の表)に対応します。

```go
	case "combat":
		maxHP := ui.catalog.maxHP(npcID, result.TargetHP)
		if ui.fight != nil && ui.fight.npcID == npcID {
			maxHP = max(maxHP, ui.fight.maxHP)
		}
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("You hit %s for %d damage. (%d HP left)", "%s に %d ダメージ!(残りHP %d)"), enemy, result.Damage, result.TargetHP))
		ui.showFight(&fightState{npcID: npcID, hp: result.TargetHP, maxHP: maxHP})
		ui.flashScene(flashHurt, 350*time.Millisecond) // the enemy survived, so it hit back
```

- **戦闘が続いている**:敵の最大HPを求めます。`catalog.maxHP`(下)で世界データから引き、それが分からなければ**今のHP**で代用します。すでに同じ敵と戦っているなら、**大きいほう**を使います(HPバーの割合がぶれないように)。
- 「◯◯に△ダメージ!(残りHP □)」を、**戦闘の色**で冒険ログに足します。
- 戦闘パネルを更新(`showFight`)し、**赤く光らせます**(`flashHurt`、15-13)。「敵が生き残った=反撃を受けた」ので、ダメージの演出をします。

```go
	case "victory":
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("Victory! %s is defeated.", "勝利!%s を倒した。"), enemy))
		ui.showFight(nil)
	case "dead":
		ui.addStoryKind(storyDeath, fmt.Sprintf(ui.tr("%s struck you down.", "%s にやられた。"), enemy))
		ui.showFight(nil)
```

- **勝利**:文章を出して、パネルを閉じます(`showFight(nil)`)。
- **死亡**(`dead`):死亡の色(赤)で文章を出して、パネルを閉じます。

```go
	case "wounded":
		// An ordinary person cannot fight back, so there is no fight panel, only the hit.
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("You strike %s for %d damage. (%d HP left)", "%s に %d ダメージを与えた。(残りHP %d)"), enemy, result.Damage, result.TargetHP))
	case "murder":
		ui.addStoryKind(storyDeath, fmt.Sprintf(ui.tr("You killed %s. The Fates cut your thread.", "%s を手にかけた。運命の女神たちがあなたの糸を断ち切った。"), enemy))
		ui.showFight(nil)
	case "smitten":
		ui.addStoryKind(storyDeath, fmt.Sprintf(ui.tr("You raised your hand against %s, and were struck dead before the blow landed.", "%s に手を上げた。一撃が届く前に打ち殺された。"), enemy))
		ui.showFight(nil)
```

- **敵以外への攻撃**で増えた3つです(8-5のケース3・4)。
  - `wounded`(一般人を傷つけた):文章だけを出します。**戦闘パネルは出しません**(一般人は反撃せず、戦闘にならないため)。
  - `murder`(殺した)と `smitten`(一撃で殺された):**死亡の色**で文章を出して、パネルを閉じます。サーバー側では、このあと自動的に運命の間へ戻される(復活する)ので、画面の更新は別の処理(`LOOK` の再取得)が行います。
- サーバーが新しい `status` を増やしても、**ここに `case` を足さない限りは `default` に落ちて**、生の文章が表示されます。

```go
	case "overwhelmed":
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("%s is too strong to beat by force.", "%s は力では敵わない。"), enemy))
		ui.showFight(nil)
	default:
		return false
	}
	return true
}
```

- `overwhelmed`(勝てない敵):「力では敵わない」と出してパネルを閉じます。
- 知らない `status` は `false` を返します(上で説明した、生の文章の表示)。

### `handleFlee` と `handleDefend`

```go
func (ui *gui) handleFlee(line string) bool {
	var result struct {
		Result string `json:"result"`
	}
	if decodeOK(line, &result) != nil || result.Result == "" {
		return false
	}
	if result.Result == "success" {
		ui.addStoryKind(storyCombat, ui.tr("You got away.", "うまく逃げ切った。"))
		ui.showFight(nil)
	} else {
		ui.addStoryKind(storyCombat, ui.tr("You could not get away!", "逃げられなかった!"))
	}
	return true
}
```

- `FLEE` の応答 `{"hp":..,"result":"success|failure|failure_dead"}`(8-6)を読みます。
- `var result struct { ... }`:**その場で型を定義して**変数にします(この関数でしか使わない、2フィールドだけの型)。
- **成功**なら「うまく逃げ切った」を出してパネルを閉じます。**失敗**なら「逃げられなかった!」を出して、**パネルは開いたまま**です(戦闘は続く)。

```go
func (ui *gui) handleDefend(line string) bool {
	...
	if decodeOK(line, &result) != nil || result.Result != "braced" {
		return false
	}
	ui.addStoryKind(storyCombat, ui.tr("You brace yourself. The next counter-attack will hurt half as much.", "身構えた。次の反撃のダメージは半分になる。"))
	return true
}
```

- `DEFEND` の応答(`"braced"`)を読んで、「身構えた。次の反撃のダメージは半分になる。」と出します。

### `worldCatalog.maxHP`— 敵の最大HP

```go
func (catalog *worldCatalog) maxHP(npcID string, fallback int) int {
	if catalog != nil {
		if hp := catalog.NPCs[npcID].HP; hp > 0 {
			return hp
		}
	}
	return fallback
}
```

- 敵の**最大HP**を、世界データ(`data/world.json`、15-2の `catalog`)の `hp` から引きます。
- サーバーの応答には「敵の**今の**HP」しか入っていないので、バーの割合(今/最大)を作るために、**最大のほう**はこうして調べます。
- 世界データが読めない(`catalog == nil`)、またはその敵が載っていないときは、`fallback`(今のHP)を返します。
- メソッドのレシーバは `*worldCatalog` ですが、`nil` でも呼べます(中で `nil` を確認している)。
- **ゼロ値の活用**:`catalog.NPCs[npcID]` は、存在しないIDなら**ゼロ値の `catalogEntry`**(HPは0)を返すので、`hp > 0` の判定だけで「載っていない」も弾けます。

## 15-15 `ui_map.go`(237行)— ミニマップ

**役割**:**訪れた部屋を、つながりの形のまま小さな地図に描く**ファイルです。部屋の絵の左の余白と、「地図」タブの2か所に同じ地図が出ます。地図の位置は、世界データの出口の方角(北・南・東・西)から**自分で計算します**(世界データに座標は書いていません)。

```go
const hubRoomID = "loc.hall_of_fates"

type gridPos struct{ x, y int }

var directionSteps = map[string]gridPos{
	"north": {0, -1}, "south": {0, 1}, "east": {1, 0}, "west": {-1, 0},
}
```

- `hubRoomID`:ハブ(運命の間)のID。地図の**原点**にします。
- `gridPos`:格子上の位置(`x`, `y`)。`struct{ x, y int }` は「同じ型のフィールドをまとめて書く」書き方です。
- `directionSteps`:**方角から、格子で1歩進む量**。北は `y` が -1(画面では上)、南は +1、東は `x` が +1、西は -1。**この4方向しか無い**ので、世界データの出口も東西南北だけで作ってあります(隠し部屋への出口が `up` や `down` でなく**南**なのは、この制約のためです)。

```go
func arcOf(roomID string) string {
	if roomID == hubRoomID {
		return ""
	}
	name, _, _ := strings.Cut(roomID, "_")
	return name
}
```

- **部屋が、どの「編」に属するか**を返します。部屋のIDは `loc.ody_cicones` のように `loc.編名_部屋名` の形なので、最初の `_` の前(`loc.ody`)が編の名前です。ハブは編に属さないので `""`。
- `strings.Cut(文字列, "_")`:最初の `_` で**前・後ろ・見つかったか**の3つに分けて返します。ここでは前だけ使い、残りは `_` で捨てます。

```go
func computeMapLayout(catalog *worldCatalog, arc string) map[string]gridPos {
	positions := map[string]gridPos{}
	if catalog == nil {
		return positions
	}
	taken := map[gridPos]bool{{0, 0}: true}
	positions[hubRoomID] = gridPos{0, 0}
	queue := []string{hubRoomID}
	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for direction, to := range catalog.Rooms[from].Exits {
			step, ok := directionSteps[direction]
			if _, placed := positions[to]; !ok || placed {
				continue
			}
			if _, exists := catalog.Rooms[to]; !exists || (to != hubRoomID && arcOf(to) != arc) {
				continue
			}
			pos := gridPos{positions[from].x + step.x, positions[from].y + step.y}
			if taken[pos] {
				continue
			}
			taken[pos] = true
			positions[to] = pos
			queue = append(queue, to)
		}
	}
	return positions
}
```

- **1つの編の全部の部屋に、格子の位置を割り当てます**。ハブを (0,0) に置き、**出口をたどって**隣の部屋を決めていきます(**幅優先探索**:近い部屋から順に)。
- `positions`:部屋ID → 位置。`taken`:**使用済みのマス**。`map[gridPos]bool{{0, 0}: true}` は、キーが構造体の辞書(キーの `gridPos{0,0}` は型名を省略できる)です。
- `queue`:これから調べる部屋の**順番待ちリスト**。`from := queue[0]` で先頭を取り出し、`queue = queue[1:]` で先頭を捨てます(`スライス[1:]` は「1番目から最後まで」)。
- 出口を1つずつ見て、次の場合は**飛ばします**(`continue`)。
  - 東西南北でない出口、または**もう位置が決まっている**部屋(`!ok || placed`)。
  - 世界に無い部屋、または**ほかの編の部屋**(`arcOf(to) != arc`、ただしハブは例外)。この関数は「1つの編の地図」だけを作ります。
  - 計算した位置が**すでに別の部屋に取られている**(`taken[pos]`):**最初に取った部屋がそのマスを持つ**というルールで、重なりを避けます。
- 位置は「今の部屋の位置 + 方角の1歩」。決まったら `taken` と `positions` に記録し、`queue` に足して**その先の部屋も調べます**。
- **一方通行の出口**(クレタ→イオルコスなど)もたどりますが、`placed` の判定で、もう位置がある部屋は上書きされません。

```go
func (ui *gui) markVisited(roomID string) {
	if roomID == "" {
		return
	}
	if ui.visited == nil {
		ui.visited = map[string]bool{}
	}
	if !ui.visited[roomID] {
		ui.visited[roomID] = true
	}
	ui.refreshMap()
}
```

- **部屋を「訪れた」と記録**して、地図を描き直します。`ui.visited` は「部屋ID → `true`」の辞書で、**最初は `nil`**なので、書き込む前に `make` 相当(`map[string]bool{}`)で作ります(3-7で説明した「`nil` の辞書には書けない」)。
- 移動が成功したとき(`main.go`)と、最初の部屋を見たとき(`ui_journal.go`)に呼ばれます。

```go
type mapCell struct {
	obj fyne.CanvasObject
	pos gridPos
	to  *gridPos // when set, obj is a line from pos to *to
}

type mapLayout struct {
	cells      []mapCell
	minX, minY int
	cols, rows int
}
```

- `mapCell`:**地図の部品1つ**。部屋の四角形か、部屋と部屋をつなぐ線です。`to` が設定されていれば線(`pos` から `*to` まで)、`nil` なら四角形です。`to` が**ポインタ**なのは、「線かどうか」を `nil` で区別するためです(2-4の `Hazard` と同じ考え方)。
- `mapLayout`:地図全体の**並べ方**(部品の一覧と、格子の範囲)。15-11の `barLayout` と同じく、Fyneのレイアウトとして働きます。

```go
func (*mapLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(200, 120) }

func (l *mapLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	if l.cols == 0 || l.rows == 0 {
		return
	}
	cell := min(size.Width/float32(l.cols), size.Height/float32(l.rows))
	originX := (size.Width - cell*float32(l.cols)) / 2
	originY := (size.Height - cell*float32(l.rows)) / 2
	center := func(p gridPos) fyne.Position {
		return fyne.NewPos(originX+(float32(p.x-l.minX)+0.5)*cell, originY+(float32(p.y-l.minY)+0.5)*cell)
	}
	...
}
```

- `MinSize`:最小の大きさは 200×120。
- `Layout`:**格子を、パネルの大きさに合わせて拡大縮小して**部品を置きます。
  - `cell`:**1マスの大きさ**。横と縦のうち、**小さいほう**に合わせます(`min`)。はみ出さず、マスが正方形になります。
  - `originX`/`originY`:地図全体を**中央に寄せる**ための、左上の位置。
  - `center`:マス `p` の**中心の座標**を返す関数(関数を変数に入れています)。`(p.x - minX + 0.5) * cell` は、「左端から何マス目か + 半マス」。
- 続き(省略)では、部品を1つずつ置きます。線(`to != nil`)は始点と終点を `center` で決め、四角形は**マスの7割の大きさ**(`cell * 0.7`)でマスの中央に置きます。マスと同じ大きさにしないのは、**隣と少し隙間を空けて**部屋を見やすくするためです。

```go
var (
	mapUnknown = color.NRGBA{R: 60, G: 70, B: 96, A: 255}
	mapVisited = color.NRGBA{R: 112, G: 128, B: 170, A: 255}
	mapHazard  = color.NRGBA{R: 214, G: 140, B: 56, A: 255}
	mapLethal  = color.NRGBA{R: 190, G: 56, B: 56, A: 255}
	mapHub     = color.NRGBA{R: 70, G: 170, B: 110, A: 255}
)

func (ui *gui) roomMapColor(roomID string) color.Color {
	if roomID == hubRoomID {
		return mapHub
	}
	if entry := ui.catalog.Rooms[roomID]; entry.Hazard != nil {
		if entry.Hazard.Type == "lethal" {
			return mapLethal
		}
		return mapHazard
	}
	return mapVisited
}
```

- 5色:**未踏の隣の部屋**(暗い)、**訪れた部屋**(青灰)、**危険な部屋**(橙)、**即死の部屋**(赤)、**ハブ**(緑)。README の地図の配色と同じです。
- `roomMapColor`:部屋の色を決めます。ハブは緑、危険のある部屋は(`lethal` なら赤、それ以外は橙)、ほかは青灰です。`entry.Hazard != nil` の `Hazard` は、世界データの `catalogEntry`(15-2)にあるポインタです。

```go
func (ui *gui) refreshMap() {
	if ui.catalog == nil {
		return
	}
	arc := arcOf(ui.room.Room.ID)
	if arc == "" {
		arc = ui.lastArc
	}
	if arc == "" {
		arc = "loc.ody"
	}
	ui.lastArc = arc
	positions := computeMapLayout(ui.catalog, arc)
	for _, box := range []*fyne.Container{ui.mapBox, ui.sceneMapBox} {
		if box != nil {
			ui.redrawMap(box, positions)
		}
	}
}
```

- **地図を描き直す入口**です。
- **どの編の地図を出すか**を決めます:今いる部屋の編。**ハブにいる**(編が空)ときは、**直前にいた編**(`ui.lastArc`)を使います。それも無ければ最初の編(`loc.ody`)です。
- 求めた編を `lastArc` に覚え、`computeMapLayout` で**位置を計算**します。
- 地図の入れ物は**2つ**あります(`ui.mapBox` は「地図」タブ、`ui.sceneMapBox` は部屋の絵の左の余白)。`for ... range []*fyne.Container{...}` で、**2つとも同じ位置データで**描きます。使われていない(`nil`)ほうは飛ばします。

```go
func (ui *gui) redrawMap(box *fyne.Container, positions map[string]gridPos) {
	layout := &mapLayout{}
	first := true
	for _, pos := range positions {
		if first || pos.x < layout.minX {
			layout.minX = pos.x
		}
		if first || pos.y < layout.minY {
			layout.minY = pos.y
		}
		first = false
	}
	maxX, maxY := layout.minX, layout.minY
	for _, pos := range positions {
		maxX, maxY = max(maxX, pos.x), max(maxY, pos.y)
	}
	layout.cols, layout.rows = maxX-layout.minX+1, maxY-layout.minY+1
```

- **1つの入れ物に地図を描きます**。まず、全部の位置から**格子の範囲**(最小・最大の `x` `y`)を求め、**列数と行数**を決めます(`最大 - 最小 + 1`)。
- `first` は「最初の1件か」の印。最初の1件では比較せず、そのまま最小値にします(**空の状態から最小値を求める定番の書き方**)。
- `maxX, maxY := ...`:2つの変数を同時に作り、次の `for` で `max` を使って更新します。

```go
	known := map[string]bool{}
	for id := range ui.visited {
		known[id] = true
		for direction, to := range ui.catalog.Rooms[id].Exits {
			if _, ok := directionSteps[direction]; ok {
				known[to] = true
			}
		}
	}
```

- **地図に出す部屋の集合**(`known`)を作ります。**訪れた部屋と、その隣の部屋**です。行ったことのない部屋でも、訪れた部屋の出口の先は「暗い四角」として出し、まだ行ける場所だと分かるようにします。
- それ以外の部屋は**出しません**(ネタバレを避けるため、全体の形は見せない)。

```go
	var lines, cells []mapCell
	for id, pos := range positions {
		if !known[id] {
			continue
		}
		if ui.visited[id] {
			for direction, to := range ui.catalog.Rooms[id].Exits {
				target, ok := positions[to]
				if _, isDir := directionSteps[direction]; ok && isDir && known[to] {
					line := canvas.NewLine(mapUnknown)
					line.StrokeWidth = 2
					lines = append(lines, mapCell{obj: line, pos: pos, to: &target})
				}
			}
		}
		rect := canvas.NewRectangle(mapUnknown)
		if ui.visited[id] {
			rect.FillColor = ui.roomMapColor(id)
		}
		if id == ui.room.Room.ID {
			rect.StrokeColor, rect.StrokeWidth = gold, 3
		} else {
			rect.StrokeColor, rect.StrokeWidth = ivory, 1
		}
		cells = append(cells, mapCell{obj: rect, pos: pos})
	}
```

- 出す部屋を1つずつ、部品にします。
- **訪れた部屋**からは、出口の先の部屋(出す対象で、位置もある)に**線**を引きます(`canvas.NewLine`)。行ったことのない部屋どうしは線で結びません。
- 四角形は、**初めは暗い色**(未踏)、**訪れていれば部屋の色**(`roomMapColor`)に塗り替えます。
- **今いる部屋だけ**、枠を**金色・太く**(3)します。ほかは薄い象牙色の細い枠(1)です。

```go
	layout.cells = append(lines, cells...)
	box.Layout = layout
	box.Objects = nil
	for _, item := range layout.cells {
		box.Objects = append(box.Objects, item.obj)
	}
	box.Refresh()
}
```

- **線を先、四角形をあとに**並べます(あとが上に描かれるので、四角形が線の上に乗って、線が四角形の下に隠れます)。`append(lines, cells...)` の `...` は、スライスの中身を**1つずつ展開して**渡す書き方です。
- 入れ物のレイアウトを今作った `mapLayout` に差し替え、中身の部品を入れ替えて、`Refresh` で描き直します。**毎回、部品を全部作り直す**単純な方式ですが、部屋が数十個なので十分速いです。

```go
func (ui *gui) buildMapTab() fyne.CanvasObject {
	ui.mapBox = container.NewWithoutLayout()
	return container.NewVScroll(ui.mapView(ui.mapBox))
}

func (ui *gui) mapView(box *fyne.Container) fyne.CanvasObject {
	legend := container.NewGridWithColumns(2,
		legendEntry(mapHub, ui.tr("Hub", "ハブ")),
		legendEntry(mapVisited, ui.tr("Visited", "訪れた")),
		legendEntry(mapHazard, ui.tr("Hazard", "危険")),
		legendEntry(mapLethal, ui.tr("Fatal", "即死")),
	)
	return container.NewBorder(nil, legend, nil, nil, box)
}

func legendEntry(c color.Color, text string) fyne.CanvasObject {
	swatch := canvas.NewRectangle(c)
	swatch.SetMinSize(fyne.NewSize(theme.Padding()*2, theme.Padding()*2))
	return container.NewBorder(nil, nil, container.NewCenter(swatch), nil, compactLabel(text))
}
```

- `buildMapTab`:**「地図」タブの中身**を作ります。`container.NewWithoutLayout()` は「並べ方を決めない入れ物」で、並べ方は `redrawMap` が後から `mapLayout` を差し込みます。縦にスクロールできるようにして返します。
- `mapView`:入れ物の**下に凡例**(色の見本と名前)を付けます。`NewBorder(上, 下, 左, 右, 中央)` は、四辺に部品を置き、残りを中央が占める配置です(`nil` は「置かない」)。凡例は2列で、**ハブ・訪れた・危険・即死**の4つです。
- `legendEntry`:**色の見本の四角と、名前**を横に並べた1行。`swatch.SetMinSize` で四角の大きさを(余白の2倍)にしています。
- この2つは `ui_layout.go`(15-8)の画面組み立てから呼ばれます。画面が広いときは、部屋の絵の左の余白に同じ地図を置いて**「地図」タブを取り除く**ので、`mapView` が2か所で使われます。

## 15-16 `ui_endings.go`(112行)— エンディング図鑑

**役割**:「クエスト」タブの下に、**達成したエンディング(と祝福)**と、**見つけた即死の部屋の数**を出すファイルです。

```go
var endingOrder = []string{"ending.argo", "ending.troy", "ending.odyssey", "ending.final"}

type endingSlot struct {
	id, rewardItem string
	name           localizedName
	blessing       string // "God: Name. Description" in the current locale, or "" if the ending has none
}
```

- `endingOrder`:図鑑に**並べる順番**(アルゴ・トロイア・オデュッセイア、最後に最終エンディング)。
- `endingSlot`:図鑑の1枠。エンディングのID、**報酬の記念品のID**、名前、**祝福の説明文**(その言語で作った文章。祝福が無ければ `""`)。

```go
func (catalog *worldCatalog) endingSlots(locale string) []endingSlot {
	var slots []endingSlot
	if catalog == nil {
		return slots
	}
	for _, npc := range catalog.NPCs {
		if npc.Ending != nil {
			slot := endingSlot{id: npc.Ending.ID, rewardItem: npc.Ending.RewardItem, name: npc.Ending.Name}
			if b := npc.Ending.Blessing; b != nil {
				slot.blessing = b.God.get(locale) + ": " + b.Name.get(locale) + " — " + b.Description.get(locale)
			}
			slots = append(slots, slot)
		}
	}
```

- **世界データから、全エンディングの枠を作ります**。エンディングは**NPCに付いている**(`npc.Ending`)ので、全NPCを調べ、エンディングを持つ(`!= nil`)ものを集めます。
- 祝福があれば、「神: 祝福名 — 説明」の1つの文章にします。`b.God.get(locale)` は**その言語の名前**(無ければ英語)を返します(15-2の `localizedName.get`)。

```go
	rank := func(id string) int {
		if i := slices.Index(endingOrder, id); i >= 0 {
			return i
		}
		return len(endingOrder)
	}
	slices.SortFunc(slots, func(a, b endingSlot) int { return rank(a.id) - rank(b.id) })
	return slots
}
```

- **並べ替え**:`rank` は「`endingOrder` の何番目か」(知らないIDは一番後ろ)を返す関数です。`slices.SortFunc` に「2つを比べる関数」を渡し、`rank` の差(負なら a が先)で並べます。辞書の `range` は順番がバラバラなので、**決まった順番に整える**ための処理です。
- `slices` はGoの標準ライブラリで、スライスの検索・並べ替えなどを提供します。

```go
func (catalog *worldCatalog) lethalRoomCount() int {
	count := 0
	if catalog == nil {
		return count
	}
	for id := range catalog.Rooms {
		if _, ok := catalog.gameOverRoom(id); ok {
			count++
		}
	}
	return count
}
```

- **即死の部屋が世界にいくつあるか**を数えます。`gameOverRoom`(15-2)は、「ハザードが `lethal` の部屋か」を返す関数で、`ok` が `true` の部屋を数えます。

```go
func (ui *gui) fatalKey() string { return "fatal." + strings.TrimSpace(ui.nameEntry.Text) }

func (ui *gui) fatalRooms() []string {
	if app := fyne.CurrentApp(); app != nil {
		return app.Preferences().StringList(ui.fatalKey())
	}
	return nil
}

func (ui *gui) recordFatalRoom(roomID string) {
	app := fyne.CurrentApp()
	if app == nil || slices.Contains(ui.fatalRooms(), roomID) {
		return
	}
	app.Preferences().SetStringList(ui.fatalKey(), append(ui.fatalRooms(), roomID))
	ui.showEndings()
}
```

- **見つけた即死の部屋の記録**(「見つけた即死の選択: 2/7」)を、アプリの**設定(`Preferences`)に保存**します。Fyneの設定は、**アプリを閉じても残る**小さな保存場所です。
- `fatalKey`:保存する**キー名**。`"fatal." + プレイヤー名` なので、**プレイヤーごとに別々**に記録されます。
- `fatalRooms`:記録済みの部屋IDの一覧を読みます。
- `recordFatalRoom`:新しい部屋を足します。**すでに記録済みなら何もしません**(`slices.Contains`)。足したら、図鑑を描き直します。
- `fyne.CurrentApp() != nil` の確認は、テストなどでアプリが無い状態でも落ちないための備えです。
- **これはサーバーのデータではなくGUIだけの記録**です。ゲームのルール(`LOOK` など)には関わりません。

```go
func (ui *gui) showEndings() {
	if ui.endingBox == nil {
		return
	}
	var rows []fyne.CanvasObject
	reached := 0
	slots := ui.catalog.endingSlots(ui.locale)
	for _, slot := range slots {
		card := textVBox(journalName("？？？"))
		if slices.Contains(ui.inventory, slot.rewardItem) {
			card = textVBox(journalName("★ " + slot.name.get(ui.locale)))
			if slot.blessing != "" {
				blessing := widget.NewLabel(slot.blessing)
				blessing.Wrapping = fyne.TextWrapWord
				card.Add(blessing)
			}
			reached++
		}
		rows = append(rows, journalCard(card))
	}
```

- **図鑑を描き直します**。エンディングごとに「カード」を作ります。
- **まだ達成していないエンディング**は、名前を伏せて **「？？？」**。
- **達成した**エンディングは、**「★ 名前」と祝福の説明**を出します。
- **達成したかどうかの判定**が面白いところです:`slices.Contains(ui.inventory, slot.rewardItem)`、つまり**報酬の記念品を持ち物に持っているか**で判断します。エンディングを迎えると記念品が持ち物に入り(9-3)、**死んでも記念品は失われない**(8-2)ので、「持っている=達成した」と見なせます。サーバーに問い合わせる必要がありません。
- `journalName`・`journalCard`・`textVBox` は15-10・15-8の部品です。

```go
	if len(slots) == 0 {
		ui.endingBox.Objects = nil
		ui.endingBox.Refresh()
		return
	}
	header := journalName(fmt.Sprintf(ui.tr("Endings %d/%d", "エンディング %d/%d"), reached, len(slots)))
	fatal := widget.NewLabel(fmt.Sprintf(ui.tr("Fatal choices found: %d/%d", "見つけた即死の選択: %d/%d"), len(ui.fatalRooms()), ui.catalog.lethalRoomCount()))
	fatal.Wrapping = fyne.TextWrapWord
	ui.endingBox.Objects = append(append([]fyne.CanvasObject{header}, rows...), fatal)
	ui.endingBox.Refresh()
}

func newEndingBox() *fyne.Container { return textVBox() }
```

- エンディングが1つも無い(世界データが読めていない)なら、空にして終わりです。
- 上に「エンディング 2/4」、下に「見つけた即死の選択: 3/7」を付けます。`append(append([]T{header}, rows...), fatal)` は「見出し + 全カード + 最後の1行」を**1本のスライスにつなぐ**書き方です。
- `newEndingBox`:図鑑の入れ物を作る小さな関数です(`textVBox` は、幅に合わせて文章の高さを測る縦並びの入れ物、15-8)。

## 15-17 `ui_item_effects.go`(58行)— アイテムの効果を表示する

**役割**:持ち物と「まわり」のアイテム行に、**そのアイテムの効果を緑(良い)と赤(悪い)の文字で出す**ファイルです。効果の中身は `data/world.json`(5-2、9-4)にあり、GUIはそれを**読んで表示するだけ**です。

```go
type itemEffect struct {
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}
```

- `itemEffect`:世界データの効果1つ(**種類と大きさ**)。サーバーの `ItemEffect`(9-4)と同じ形で、GUIが自分で読み込みます(`catalogEntry.Effects`、15-2)。

```go
func (ui *gui) effectLine(e itemEffect) string {
	switch e.Effect {
	case "damage_bonus":
		return fmt.Sprintf(ui.tr("Damage dealt %+d", "与ダメージ %+d"), e.Value)
	case "counter_reduction":
		return fmt.Sprintf(ui.tr("Damage taken %+d%%", "被ダメージ %+d%%"), -e.Value)
	case "regen_bonus":
		return fmt.Sprintf(ui.tr("Regeneration %+d", "回復速度 %+d"), e.Value)
	case "max_hp":
		return fmt.Sprintf(ui.tr("Max HP %+d", "最大HP %+d"), e.Value)
	}
	return ""
}
```

- **効果1つを、1行の文章にします**。
- `%+d`:**符号付き**の整数(`+3` や `-5`)を埋め込む書式です。「与ダメージ +3」のように、プラスにも `+` が付きます。
- `counter_reduction`(反撃の軽減)だけ**符号を逆にします**(`-e.Value`)。サーバーでは「軽減%が大きいほど良い」ので `+15` は「15%軽減」ですが、**画面では「被ダメージ -15%」と書く**ほうが自然だからです。悪い効果(`-20`)は「被ダメージ +20%」になります。`%%` は、書式の中で文字の `%` を出す書き方です。
- 知らない種類は `""` を返し、**表示しません**(世界データに新しい効果を足しても、GUIが落ちません)。

```go
func (ui *gui) itemEffectLines(id string) []fyne.CanvasObject {
	if ui.catalog == nil {
		return nil
	}
	var lines []fyne.CanvasObject
	for _, e := range ui.catalog.Items[id].Effects {
		text := ui.effectLine(e)
		if text == "" {
			continue
		}
		colour := storyColorFor(colorStoryTeam)
		if e.Value < 0 {
			colour = storyColorFor(colorStoryDeath)
		}
		line := canvas.NewText(text, colour)
		line.TextSize = 12
		lines = append(lines, line)
	}
	return lines
}
```

- **アイテムの効果を、色付きの文字の部品のリストにして返します**。
- 効果を1つずつ文章にし、**値が0より小さければ赤、そうでなければ緑**にします。「悪い効果は赤」と、一目で分かります。
- `canvas.NewText(文字, 色)`:色付きの1行の文字の部品です(`widget.Label` と違い、**折り返しはしません**が、色を自由に付けられます)。`TextSize = 12` で、普通より少し小さくします。
- 戻り値は `[]fyne.CanvasObject`(部品のスライス)。呼ぶ側(`itemRow`、15-10)が、アイテム名の下に1行ずつ足します。

```go
func storyColorFor(name fyne.ThemeColorName) color.Color {
	c, _ := storyColor(name)
	return c
}
```

- `storyColor`(15-12)は「色」と「`bool`」の2つを返すので、**色だけがほしいとき用の包み関数**です。2つ目を `_` で捨てて、色だけを返します。
- 緑(仲間の色)と赤(死亡の色)を**そのまま使い回す**ので、効果の色とログの色が揃います。

> **ねらい**:アイテムの効果の**データ**(`world.json`)、**計算**(サーバーの `item_effects.go`、9-4)、**表示**(このファイル)の3つが、別々のファイルに分かれています。数値を調整したいときは `world.json` だけを直せば、計算にも画面にも反映されます。

## 15-18 `cmd/server/gui_state.go`— STATE拡張

GUI上部の仲間表示とマウスによるグループ操作に使う、接続済みプレイヤー向けの追加コマンドです。RFCのSTATUS・LOOK・WHOの応答形式にはフィールドを足しません。STATEを送ったクライアントへだけ返します。

```text
STATE
OK {"crew":8,"crew_initialized":true,"players":["alice","bob"],"group":"group.1","invitations":[]}
```

```go
func handleState(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	s.mu.Lock()
	player := s.players[*name]
	if player == nil || player.exiting {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	state := struct {
		Crew            int      `json:"crew"`
		CrewInitialized bool     `json:"crew_initialized"`
		Players         []string `json:"players"`
		Group           string   `json:"group"`
		Invitations     []string `json:"invitations"`
	}{
		Crew: player.Crew, CrewInitialized: player.CrewInitialized,
		Group: s.groupByPlayer[*name], Players: []string{}, Invitations: []string{},
	}
	for otherName, other := range s.players {
		if !other.exiting {
			state.Players = append(state.Players, otherName)
		}
	}
	for _, group := range s.groups {
		if _, invited := group.Invited[*name]; invited {
			state.Invitations = append(state.Invitations, group.Leader)
		}
	}
	sort.Strings(state.Players)
	sort.Strings(state.Invitations)
	data, err := json.Marshal(state)
	if err != nil || len(data)+len("OK ") > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}
```

- 引数はSTATEだけです。未接続・余分な引数にはERR 400、退出中や不正な状態にはERR 500を返します。
- `s.mu`を保持し、Crewと初期化済みか、退出中でないオンラインの名前、自分のグループ、自分宛ての招待をコピーします。招待一覧には他人宛ての招待を含めません。
- 名前と招待をソートし、空の一覧もJSONの`[]`として返します。応答の長さを検査し、送信キューへ追加してからロックを解除します。TCP書き込み完了を待つ間はロックを保持しません。
- 拡張非対応のサーバーでSTATEがERR 400なら、GUIは再送を止めます。仲間は「-」表示となり、招待先は同じ部屋の人、参加先は受信した招待から選びます。RFCのゲーム操作とチャットは続けられます。

### 関連テスト

- `ui_mouse_test.go`:一覧からIDを送る操作、持ち物／クエストを開く操作、名前入力なしの招待／参加、10件目以降の対象、部屋変更での選択画面の破棄、0人と未取得の区別、非対応サーバー、HPの定期取得、大小のウインドウを確認します。
- `gui_state_test.go`:自分宛ての招待だけが返ること、Crew・オンライン・グループの内容、STATUSとWHOの従来応答、TCP接続上の移動による12人から10人への変化を確認します。

---

## 15-19 GUIのテスト(戦闘・地図・演出・表示の検証)

GUIの各ファイルは、15-1〜15-17で1つずつ説明しました(ファイルの一覧は冒頭の索引を参照)。ここでは、**表示が崩れないこと**を確かめる、画面まわりのテストを説明します。テスト全体の一覧は第18章にあります。

`ui_overlap_test.go`は日本語・英語、文字サイズ15・22、画面の横並び・縦並びの切替で、文章・ボタンの重なりと画像の枠外表示を確認します。4種類の一覧、戦闘中のパネル、ゲームオーバー画面、アイテム画像の縮小と1枚ずつのスクロールも確認します。実際の部屋・NPC・クエストを使い、長い日本語名を追加して折り返しを検証します。

`ui_initial_room_test.go`では、未接続の画面を表示し、英語から日本語へ切り替えた後に、最初の部屋を読み込みます。タブ切替・サイズ変更をしない状態の「まわり」の配置を検証し、その後タブを切り替えても一覧の高さが変わらないことを確認します。文字サイズ15・22、1280×900・640×900で確認します。

`ui_map_test.go`では、日本語・英語、文字サイズ15・22で、広い画面の地図が左の余白に収まり、右の地図タブを取り除くことを確認します。幅を狭めると同じ地図タブを戻すこと、持ち物・クエストの選択とチャットの下書きを保つこと、言語切替後も広い画面では地図タブを出さないことも確認します。2か所の地図は描画部品を共有せずに訪問済みの部屋を更新します。

Fyneのテスト用ウインドウは実ウインドウの最小サイズ制約を自動では適用しないため、テスト側で現在の最小サイズ以上へ調整します。`TAP_GUI_PREVIEW_DIR`に既存の出力ディレクトリを指定すると、同じ検証画面をPNGへ保存できます。通常のテストではPNGをファイルへ書きません。これはFyneのテスト用Canvasでの確認であり、OSの実ウインドウの起動確認とは別です。

# 第16章 `data/world.json`(ゲームの内容)

**役割**:ゲームの**中身のすべて**(部屋・アイテム・NPC・クエスト・ヒント)を書いたデータファイルです。コードを1行も変えずに、ここを書き換えるだけでゲームの内容を変えられます。サーバーは起動時に読み込み(`loadWorld`、第5章)、検査(`validate`)してから使います。GUIも**表示名や解説を引くために**同じファイルを読みます(`loadCatalog`、15-2)。

## 16-1 全体の形

```json
{
  "start_room_id": "loc.hall_of_fates",
  "rooms":  { "loc.hall_of_fates": { ... }, ... },
  "items":  { "item.olive_stake":  { ... }, ... },
  "npcs":   { "npc.polyphemus":    { ... }, ... },
  "quests": { "quest.blind_the_cyclops": { ... }, ... },
  "hints":  { "npc.polyphemus": { "en": "...", "ja": "..." }, ... }
}
```

- 一番外側が `World` 構造体(5-5)に1対1で対応します。
- `start_room_id`:新しいプレイヤーが始まる部屋(運命の間)。
- `rooms` `items` `npcs` `quests`:**「ID → 中身」の辞書**です。IDは `loc.`(部屋)、`item.`、`npc.`、`quest.` で始まります。この接頭辞はルールではなく**読みやすさのための決まり**です。
- `hints`:死因のID(NPC・部屋・アイテム)→ モイライのヒント文(11-1-6)。
- 現在の数:**部屋47・アイテム21・NPC44・クエスト16・ヒント22**。
- 文章はすべて **`{"en": "...", "ja": "..."}`** の形です(`LocalizedText`、第4章)。英語と日本語の両方を書きます。

## 16-2 部屋 `rooms`

```json
"loc.ody_cicones": {
  "id": "loc.ody_cicones",
  "name": { "en": "Ismarus, Land of the Cicones", "ja": "イスマロス、キコネス人の地" },
  "description": { "en": "...", "ja": "..." },
  "exits": { "west": "loc.ody_troy_shore", "east": "loc.ody_lotus", "north": "loc.ody_ismarus_feast" },
  "hazard": { "type": "crew_cost", "crew_loss": 2 }
}
```

- `id`:キーと**同じ文字列**を中にも書きます(違うと起動時にエラー、5-7)。
- `exits`:「方角 → 行き先の部屋ID」。方角は `north` `south` `east` `west` の4つだけ使っています(GUIの地図が4方向しか扱えないため、15-15)。**行き先が存在しないとエラー**。
- `hazard`(任意):部屋に入ったときの危険。`type` は `lethal`(必ず死ぬ)、`item_gate`(`required_item_id` が無いと死ぬ)、`crew_gate`(仲間が足りないと死ぬ。`crew_loss`、`min_party_total`)、`crew_cost`(仲間が減る。`crew_loss`)(2-3、8-1)。
- `secret_exits`(任意):特定のエンディングのあとに開く出口。運命の間の南だけが持っています(2-5)。
- **一方通行**の出口は、行き先の側に戻る出口を書かないだけで作れます(例:イタケの岸辺の東は運命の間へ行くが、運命の間の東はトロイアからの出発へ行く)。

## 16-3 アイテム `items`

```json
"item.olive_stake": {
  "name": { "en": "Sharpened Olive Stake", "ja": "研がれたオリーブの杭" },
  "description": { "en": "...", "ja": "..." },
  "effects": [ { "effect": "damage_bonus", "value": 3 } ],
  "room_id": "loc.ody_cyclops",
  "obtainable": true,
  "renewable": true
}
```

- `room_id`:**最初に置いてある部屋**。
- `obtainable`:`TAKE` で拾えるか。
- `renewable`:`true` なら、**拾っても部屋から無くならず、何度でも手に入る**。クエストや神話の関門に必要なアイテムは、**誰かが持ち去って他の人が進めなくなる**のを防ぐため、すべて `renewable` です(「みんなの世界で、誰も他の人を詰ませない」設計、README の World Design)。
- `reward_only`:`true` なら、**エンディングの報酬としてだけ手に入る記念品**(部屋に置かない。`room_id` は空、`obtainable` は `false` が必須、5-7)。全部で4つ(アルゴ号の月桂冠・トロイアの燃えさし・イタケのオリーブの枝・運命の糸)。
- `effects`:**持っている間だけ効く効果**(9-4)。`effect` は `damage_bonus`(与ダメージ)・`counter_reduction`(被ダメージの軽減%)・`regen_bonus`(回復速度)・`max_hp`(最大HP)、`value` は大きさ(**マイナスは悪い効果**)。全アイテムに設定済みで、良い物(オデュッセウスの弓:与ダメージ+4)も悪い物(ヘリオスの牛:反撃+20%)もあります。
- アイテムの絵は `cmd/gui/assets/items/<アイテムID>.png`(15-6)。

## 16-4 NPC `npcs`

```json
"npc.polyphemus": {
  "name": { ... }, "description": { ... },
  "role": "enemy",
  "room_id": "loc.ody_cyclops",
  "hp": 70,
  "dialogue": [ { "en": "...", "ja": "..." } ],
  "myth_requirement_item": "item.olive_stake",
  "flee_accurate": true
}
```

- `role`:**役割**。`enemy`(戦う敵・12人)、`quest_giver`(依頼者・16人)、`dialogue`(話すだけ・16人)(課題の「3つ以上の役割」の要件)。
- `hp`:敵の最大HP。**敵以外のNPCも持っています**(殴ったときに倒せるまでの回数になる、8-5のケース4)。
- `dialogue`:台詞のリスト。`TALK` は先頭を返し、案内役は残りも通知で送る(7-19、11-1-5)。
- 追加の設定(**必要な人だけ**書く。数は現在のもの):

| キー | 意味 | 数 | 詳しくは |
|---|---|---|---|
| `myth_requirement_item` / `myth_requirement_quest` | 攻撃・会話の前に**必要なアイテム/達成済みクエスト**。無いまま挑むと即死 | 6 / 1 | 3-5、8-5 |
| `flee_accurate` / `flee_succeeds_once` | `FLEE` が**必ず成功** / **最初の1回だけ成功** | 7 / 1 | 8-6 |
| `unwinnable` + `crew_loss_on_attack` | **勝てない敵**。攻撃すると仲間が減るだけ | 1 | 8-5 |
| `mighty` | **強すぎる人**。攻撃すると**一撃で死ぬ** | 10 | 8-5 |
| `dialogue_cleared` | その部屋の敵を全部倒したあとの台詞 | 5 | 8-7 |
| `guide` | 新人向けの案内役(モイライ) | 1 | 11-1-5 |
| `ending` | このNPCと話すと迎えるエンディング | 4 | 9-3 |

- `ending`(アルゴ船編はペリアス王、トロイア編はアイネイアス、オデュッセイア編はペネロペイア、最終はモイライ)の中身:
  - `id`・`name`・`text`(本文の行のリスト)・`hint`(条件が足りないときの文)。
  - `requires_items` / `requires_quests` / `requires_endings`:迎えるための条件(最終エンディングは3つのエンディングが必要)。
  - `reward_item`:報酬の記念品。`blessing`:その神の祝福(`god`・`name`・`description`・`effect`・`value`、9-3)。

## 16-5 クエスト `quests`

```json
"quest.blind_the_cyclops": {
  "name": { ... }, "description": { ... },
  "giver_npc_id": "npc.trapped_sailor",
  "objective": { "type": "defeat_npc", "target_id": "npc.polyphemus", "count": 1 },
  "reward": { "hp": 15 }
}
```

- `giver_npc_id`:**依頼者**のNPC。このNPCに `QUEST` すると受けられます。
- `objective`:目的。`type` は `collect_item`(アイテムを持つ・10件)か `defeat_npc`(敵を倒す・6件)。`target_id` は対象のID、`count` は必要な数。
- `reward.hp`:報酬の大きさ。**最大HPの上昇量はこの値の1/5**(最低1)です(9-2、9-4)。名前が `hp` のままなのは設計の名残。
- クエストの達成は**自動**です。完了報告のコマンドは無く、サーバーが拾う・倒すを見て進めます(9-2)。

## 16-6 データを足すとき

例:**新しい部屋とアイテムを足す**。

1. `rooms` に新しい部屋を足し、既存の部屋の `exits` から**つなぐ**(出口の行き先がIDと一致しているか)。
2. `items` に新しいアイテムを足す(`room_id` は実在する部屋、`effects` も忘れずに)。
3. サーバーを起動する(`make run-server`)。**間違いがあれば、起動時に `validate`(5-7)が場所を示すエラーで止めます**。例:`room "loc.x" exit "east" points to unknown room "loc.y"`。
4. GUIで見たいなら、部屋の絵を `cmd/gui/assets/rooms/<部屋ID>.png`(960×576)に置く。**全部屋に絵が必要**で、無いと `go test`(`TestArtCoversWorld`、18章)が落ちます。

- **日本語と英語の両方**を書くこと。片方が空だと、その言語で文章が出ません(`LocalizedText.Get` は無ければ英語へ戻る、4-2)。
- **IDは変えない**こと。セーブデータ(`saves/`)がIDで部屋やアイテムを覚えているので、IDを変えると昔のセーブが読めなくなります。
- 数値(HP・効果・報酬)はすべてここにあるので、**バランス調整はこのファイルだけ**で済みます。

---

# 第17章 ビルドと設定(`Makefile`・`go.mod`・`.gitignore`・`saves/`)

**役割**:プログラムの作り方・動かし方と、Gitに入れないファイルを決める設定です。

## 17-1 `Makefile`

```make
GO      ?= go
BIN_DIR ?= bin
ADDR    ?= 127.0.0.1:4242

.DEFAULT_GOAL := help
.PHONY: help install build run-server run-client run-client-gui lint test clean
```

- `make 目的名` で、決まった手順を実行します。課題の要件(install・run-server・run-client・run-client-gui・lint・clean)を満たします。
- `GO ?= go`:変数。`?=` は「**すでに設定されていなければ**この値」の意味で、`make GO=go1.25 build` のように外から上書きできます。`ADDR` はCLIの接続先です。
- `.DEFAULT_GOAL := help`:`make` だけ打つと `help` が動きます。`.PHONY` は「これらはファイル名ではなく**目的の名前**」という宣言です(同じ名前のファイルがあっても動くように)。

| 目的 | やること |
|---|---|
| `make help` | 目的の一覧を表示(`## ` の後ろの説明を `grep` と `awk` で拾う) |
| `make install` | `go mod download`:依存ライブラリ(Fyneなど)をダウンロード |
| `make build` | サーバー・CLI・GUIを `bin/` にビルド |
| `make run-server` | `go run ./cmd/server`:サーバーを `:4242` で起動(ログはJSONで標準エラーへ) |
| `make run-client` | `go run ./cmd/cli $(ADDR)`:CLIで接続 |
| `make run-client-gui` | GUIを起動(下の注意) |
| `make lint` | `gofmt -l cmd` で**整形が必要なファイルを検出**し、あれば失敗。続けて `go vet ./...` |
| `make test` | `go test ./...` |
| `make clean` | `bin/` を消す(`saves/` は残す) |

- `run-client-gui` は、起動前に環境変数 `LANGUAGE` を整えます。`sed` で、先頭と末尾の `:` や連続する `:` を取り除きます。Fyneが言語の一覧を読むときに、空の項目で警告を出すのを避けるためです(`$$` は、`make` の中でシェルの `$` を書く書き方)。
- `lint` の `[ -n "$$unformatted" ]` は、「変数が空でなければ」。`gofmt -l` は**整形が必要なファイル名を表示するだけ**(何も表示されなければ全部整っている)なので、その出力を見て失敗にしています。
- **`make` はリポジトリのルートで実行する**こと:サーバーは `data/world.json` と `saves/` を**相対パス**で読むためです。

## 17-2 `go.mod` と `go.sum`

```
module github.com/anju0618/The_Answer_Protocol

go 1.25.0

require fyne.io/fyne/v2 v2.8.1
```

- `module`:このプロジェクトの**名前**(パッケージの住所)。
- `go 1.25.0`:必要なGoの最低バージョン。
- `require`:使うライブラリ。**直接使うのはGUIのFyneだけ**で、サーバーとCLIは標準ライブラリだけで動きます。下に並ぶ `// indirect` は、Fyneが使っている**間接的な**ライブラリです。
- `go.sum`:ダウンロードしたライブラリの**改ざん検出用のハッシュ**。手で触らず、`go mod tidy` などが更新します。

## 17-3 `.gitignore`

```
*.pdf
en.subject.txt
protocol-rfc.html
rfc.tar.gz
/saves/*
/server
/cli
/bin/
```

- **Gitに入れないもの**:
  - 学校から配られたファイル(課題のPDF・テキスト、RFC)。
  - **セーブデータ**(`/saves/*`):プレイヤーの保存と、アイテムの場所(`playerdata.json`、`itemdata.json`)。遊ぶたびに変わるので、リポジトリに入れません。
  - ビルドの成果物(`/server` `/cli` `/bin/`)。**実行ファイルはコミットしない**のが決まりです(作り直せるうえ、サイズも大きいため)。

## 17-4 `saves/` フォルダ

- サーバーが**実行時に作る**保存フォルダです(`os.MkdirAll`、12-2)。
  - `playerdata.json`:プレイヤーごとの状態(位置・HP・持ち物・クエスト・エンディング・`max_hp_bonus` など。第3章の `Player`)。
  - `itemdata.json`:**拾われて動いたアイテムの現在地**(12-2)。
- 課題は「再起動で状態が消えてよい」ですが、このゲームは**保存します**(READMEの「Architecture」に書いてあります)。消して最初からやり直したいときは、このフォルダの中身を消せば済みます。

---

# 第18章 テスト全体(`*_test.go`)

**役割**:コードが正しく動くことを**自動で確かめる**ファイルです。Goでは、`_test.go` で終わるファイルの `func TestXxx(t *testing.T)` が、`go test` で自動実行されます。**現在42ファイル、約6500行**。

## 18-1 テストの読み方

- `go test ./...`:全部のテストを実行します(`make test`)。`go test -run TestName -v ./cmd/server` で**1つだけ**実行できます。`-race` を付けると、**同時アクセスのバグ**(データ競合)も検出します。
- テストは `t.Errorf`(失敗を記録して続ける)や `t.Fatalf`(失敗して**その場で止める**)で、結果が期待どおりかを確かめます。
- 本物の通信はしません。サーバーのテストは **`net.Pipe()`**(メモリの中の疑似的な接続)で、サーバーとクライアントをつなぎます。

### サーバーのテストの共通部品(`server_test.go` ほか)

```go
func startTestClient(t *testing.T, server *Server) *testClient {
	serverConn, clientConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		server.handleClient(serverConn)
		close(done)
	}()
	...
	client.expect(t, "OK hello proto=1")
	return client
}
```

- `startTestClient`:**本物のクライアントのふり**をする接続を1つ作ります。サーバーの `handleClient`(7-21)を別のgoroutineで動かし、最初の挨拶 `OK hello proto=1` が来ることを確かめて、操作用の `testClient` を返します。`t.Cleanup` で、テストが終わったら接続を閉じます。
- `testClient` の便利メソッド:`connect`(`CONNECT` して成功を確かめる)、`cmd`(コマンドを送って**応答が期待どおり**か確かめる)、`cmdJSON`(応答のJSONを辞書で返す)、`waitEvent`(指定の `EVT` が来るまで待つ)など。
- `newServer(t.TempDir())`:**一時フォルダを保存先にした**サーバーを作ります。テストが終わると自動で消えるので、本物の `saves/` は汚れません。
- `logging_test.go` の `TestMain`:全テストの最初に呼ばれ、**ログの出力先を捨てる**(`io.Discard`)設定にします。テストの画面がログで埋まらないようにするためです。
- `en("...")`・`ens(...)`(`locale_test_helpers_test.go`):テスト用に**英語だけの `LocalizedText`** を手早く作る関数です。

## 18-2 サーバーのテスト(`cmd/server/`)

| ファイル | 何を確かめるか |
|---|---|
| `server_test.go`(964行) | 基本のコマンドすべて(`LOOK`・`TAKE`・`MOVE`・`QUIT`・接続/切断・保存と復元)の成功と失敗。同時に接続・同時に拾うなど**同時実行**の確認 |
| `auth_test.go` | 名前だけで登録・再接続できること。保存フォルダに書けないときの扱い |
| `resource_test.go` | `INVENTORY`・`DROP`・`TALK`、アイテムの置き場所と持ち主が**再起動しても残る**こと、保存に失敗したときに持ち物が壊れないこと |
| `world_test.go` | 実際の `world.json` が読めること、**不正な参照(存在しない部屋など)を検査が弾く**こと |
| `combat_test.go` | 神話の関門で即死、勝てない敵、クエスト達成で最大HPが上がり全回復、`FLEE`、部屋のハザード、生きた敵が出口を塞ぐこと |
| `combat_room_regression_test.go` | 逃げたあとの再攻撃・古い戦闘対象が残ったときの`FLEE`など、**過去に直した不具合の再発防止** |
| `defend_test.go` | `DEFEND` は戦闘中だけ使え、次の反撃を半分にする |
| `defeat_test.go` | **倒した敵の記録**(`LOOK` の `defeated`、突破後のセリフの切り替え)、死ぬと敵が戻ること、**一般人を殺すと自分が死ぬ・強い人は一撃・倒した敵は405** |
| `blessing_test.go` | 祝福がエンディングに応じて効くこと、反撃の軽減に**上限・最低1ダメージ**、回復ボーナス |
| `item_effects_test.go` | アイテムの効果が**持っている間だけ**効くこと、最大HPの**下限20**、**クエストの最大HPは死んでも残る**、悪い効果の上限、効果の検査 |
| `secret_exit_test.go` | 隠し出口が、エンディング達成**前は見えず、後は見える**こと |
| `endings_test.go`(530行) | 3つの編が独立していること、**全員が全編を最後まで遊べる**こと、条件が足りないエンディングは何も教えないこと、死亡で記念品以外を失うこと、仲間と勝利を分け合うこと、敵のHPがプレイヤーごとであること、自動回復 |
| `odyssey_integration_test.go` | **本物のデータ**でオデュッセイア編を最後まで通すこと、史実に反する選択がすべて即死になること |
| `argonauts_troy_integration_test.go` | 本物のデータで、アルゴ船編とトロイア編の神話の関門 |
| `ja_localization_integration_test.go` | 日本語の翻訳が、本物のデータで欠けていないこと |
| `locale_test.go` | `LANG` で物語の言語が切り替わること |
| `notify_test.go` | 死亡メッセージが本人に届くこと、**死因の文面が全部正しく書式化される**こと、初回接続の案内、クエストの案内と達成 |
| `hint_test.go` | ヒントが**正しい形**であること、存在しないIDのヒントを弾くこと、モイライが**直前の死因について1回だけ**教えること |
| `guide_integration_test.go` | 本物のデータで案内役・ヒント・依頼の案内がそろっていること |
| `hardcore_persistence_test.go` | 死んで失ったアイテムの持ち主が再起動後も正しいこと、保存に失敗したら持ち物を守ること |
| `saved_state_regression_test.go` | 壊れたセーブ(クエストが `null` など)で**他のプレイヤーが巻き込まれない**こと、保存済みプレイヤーの回復が再開すること |
| `no_softlock_test.go` | **倒せず逃げられない敵がいない**こと(行き詰まり=ソフトロックの防止) |
| `gui_state_test.go` | GUI向けの `STATE` の応答(仲間・招待)が正しいこと |
| `group_chat_test.go` | `GROUP` の作成・招待・参加・脱退、リーダー切断で引き継ぎ、`CHAT` の範囲、**受信者が1人壊れても他に配信が続く** |
| `flavor_line_test.go` | 名前が長くても、戦闘の実況が1行に収まって読めること |
| `client_conn_test.go` | 書き込みが止まったクライアントへの送信が**タイムアウトする**こと |
| `logging_test.go` | 接続・コマンド・応答・エラーのログ、世界の変化とクエストのログ、**不正利用(連打)の検知** |
| `locale_test_helpers_test.go` | 上の `en`・`ens`(テスト部品) |

## 18-3 CLIのテスト(`cmd/cli/main_test.go`)

- `QUIT` のエラー後も次のコマンドを打てること、`QUIT` が**自分の応答を待ってから**終わること(14章の送受信の流れの確認)。

## 18-4 GUIのテスト(`cmd/gui/`)

Fyneのテスト用の仮想ウインドウ(`fyne.io/fyne/v2/test`)で動かすので、**画面を出さずに**部品の配置や色を確かめられます。

| ファイル | 何を確かめるか |
|---|---|
| `art_test.go` | **全部屋・全NPCに絵がある**こと、絵の大きさ、同じ絵を使い回していないこと、部屋やNPCで場面が変わること、未知の部屋は `unknown` の絵になること、**倒した敵の絵が生きているときと違い、横長**であること |
| `item_photos_test.go` | 全アイテムに写真があること、`LOOK` に従って写真が変わること |
| `main_test.go` | 言語切替、マウスで `TAKE` がIDで送られること、描画、部屋の絵の更新、`QUIT` のエラー、**即死の部屋でゲームオーバー画面**が出ること |
| `protocol_test.go` | 通信層:イベントが挟まっても応答が対応づくこと、**空や複数行のコマンドを拒否**すること |
| `ui_bars_test.go` | HPバーの割合(はみ出し・0)と**色の切り替え** |
| `ui_combat_test.go` | 戦闘パネルの表示と非表示、逃げる・移動で閉じること、読めない応答は生の文で出すこと、`DEFEND` の説明、**倒した敵の絵がその場で変わる**こと、**一般人への攻撃の表示**(傷つけた・殺した・一撃) |
| `ui_effects_test.go` | フラッシュが、画面を作る前でも落ちず、アニメーションを始めること |
| `ui_endings_test.go` | エンディングの順番(最終が最後)、記念品を持つと図鑑に出ること、**見つけた即死の部屋の記録** |
| `ui_initial_room_test.go` | 日本語に切り替えたあとの最初の「まわり」が、タブを切り替えなくても**正しく並ぶ**こと |
| `ui_map_test.go` | 全部屋が別々のマスに置かれること、広い画面で地図が左の余白に出ること、**訪れた部屋とその隣だけ**が出ること |
| `ui_mouse_test.go` | マウス操作(アイテム・グループの招待と参加)、10個目以降の対象、`RFC` どおりのサーバーでも動く(`STATE` なし)こと、レスポンシブなレイアウト |
| `ui_overlap_test.go`(269行) | 日本語・英語、文字サイズ15・22、横並び・縦並びで、**文章やボタンが重ならず、絵が枠からはみ出さない**こと。戦闘中の配置、ゲームオーバー画面、アイテム画像の縮小 |
| `ui_story_test.go` | 冒険ログが**種類ごとの色**で出ること、色の名前がテーマで解決できること |

- GUIのテストは、**実際のウインドウを開く確認とは別**です(仮想ウインドウでの確認)。環境変数 `TAP_GUI_PREVIEW_DIR` を設定すると、確認用の画面をPNGで保存できます(15-19)。

## 18-5 テストを足すとき

1. 確かめたいことを1つに絞り、`TestXxx` という名前で `_test.go` に書く(名前は「何が起きるか」を表す文にする)。
2. サーバーなら `startTestClient` で接続し、`cmd` で「このコマンドを送ったら、この応答」と書く。内部の状態を見たいときは `server.mu.Lock()` して `server.players["alice"]` を直接読む(**ロックを取ること**)。
3. **乱数**(ダメージ)に頼るときは、`randDamage`(8-3)を固定値の関数に差し替えます(`var randDamage = func...` と変数にしてあるのは、このためです)。
4. 直した不具合には、**再発防止のテスト**を足します(`combat_room_regression_test.go` や `saved_state_regression_test.go` がその例)。

---

# おわりに

ここまでで、サーバー(第1〜13章)・CLI(第14章)・GUI(第15章)のコードと、ゲームのデータ(第16章)、ビルドと設定(第17章)、テスト(第18章)を、すべて上から順に読みました。

## 読み終わったあとの確認ポイント

評価で説明できるようにするため、次の問いに自分の言葉で答えられるか確認してください。

1. **接続が来てからコマンドが処理されるまでの流れ**を説明できる?
   → `Accept` → `go handleClient` → `Scan` で1行読む → `commandHandlers` から担当関数 → 応答を返す(第1章、7-21)
2. **遅いクライアントへの送信中に、共通ロックを保持し続けない理由**は?
   → 送信キューへ積み、共通ロックを外してから`waitResponse`で書き込み完了を待つ。実際のネットワーク書き込みは専用goroutineが行う(第6章、7-10)
3. **セーブが途中で壊れない理由**は?
   → 一時ファイルに書いて `Rename` で入れ替える(12-1、12-2)
4. **世界のデータを変えるときは、どこを触るか**?
   → `data/world.json` だけ。起動時に `validate` が矛盾を見つける(第5章、第16章)
5. **神話のゲート(即死)の仕組み**は?
   → `meetsMythRequirement`(3-5)と、 `handleAttack` / `handleTalk` の分岐(8-5, 7-19)
6. **新しいコマンドを足すには**?
   → `handleXxx` を書いて `commandHandlers` に1行足す(7-9)
7. **隠し部屋が、エンディングを見た人にだけ見えるのはなぜ**?
   → `Room.exitsFor` が、`SecretExits` のうち条件を満たしたものだけを足した出口の辞書を返し、`LOOK` と `MOVE` の両方がそれを使うから(2-6、7-12、7-13)
8. **クエストの報酬が「最大HPの上昇」になっているのは、どこで決まる**?
   → `quest.go` の達成処理が `questMaxHPGain`(報酬の1/5)を `MaxHPBonus` に足し、HPを全回復する。`MaxHPBonus` は死んでも残る(9-2、9-4、3-3)
9. **アイテムの良い効果・悪い効果は、どこで計算される**?
   → `world.json` の `effects` を、`itemEffectTotalLocked` が持ち物から合計し、祝福と足し合わせた `effectTotalLocked` を、戦闘と回復が呼ぶ(9-4、8-5、8-6)
10. **敵を倒すと世界が変わる仕組みは**?
   → 倒した記録は `EnemyHP`(プレイヤーごと)。`LOOK` の `defeated` で絵が変わり(15-5)、`dialogueFor` で部屋を突破した人にだけ別のセリフを返す(8-7、7-19)
11. **一般人を攻撃するとどうなる? 強い人は?**
   → 一般人は反撃しないが、倒すと自分が死ぬ(`attack_murder`)。`mighty` の人は一撃で死ぬ(`attack_mighty`)。`ERR 405` は倒した敵だけ(8-5)
12. **RFCと違う点は、どこに書いてある?**
   → README の「Protocol Implementation」の表(`ATTACK`・`LOOK`・`QUEST`・`STATUS`・`WHO`・`TALK`)
