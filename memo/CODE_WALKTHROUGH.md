# The Answer Protocol コード1行ずつ解説(Go初心者向け)

ソースを**上から順に、1文ずつ**読んでいくための解説です。

## このドキュメントの読み方

- 各章は「1ファイル」です。まずそのファイルの役割を1〜2行で書き、そのあとコードを**数行ずつのかたまり**に切って、各行を説明します。
- コードの下の `-` の箇条書きが、上のコードの**1行ずつの解説**です。
- 初めて出てくる Go の文法は、そこで説明します(後ろの章では説明を省略します)。
- 読む順番は「分からないものが少ない順」に並べています。前の章で説明済みのものは、後ろの章で使えるようになっています。
- 対象は `cmd/server/`、`cmd/cli/`、`cmd/gui/` の実装コードです。`*_test.go` は関連する処理を確認するためのテストとして参照します。
- コード中の単独の `...` は省略の印です。完全な関数は対応するソースファイルを参照します。7-10の `handleXxx` は共通の流れを説明するための例です。

## 読む順番(全体の予定)

| 章 | ファイル | 内容 | 状態 |
|---|---|---|---|
| 1 | `cmd/server/main.go` | サーバーの起動 | 書いた |
| 2 | `cmd/server/room.go` | 部屋の型 | 書いた |
| 3 | `cmd/server/player.go` | プレイヤーの型・HP回復 | 書いた |
| 4 | `cmd/server/locale.go` | 日英切替 | 書いた |
| 5 | `cmd/server/world.go` | 世界データの読み込み・検査 | 書いた |
| 6 | `cmd/server/client_conn.go` | 接続ごとの送信キュー | 書いた |
| 7 | `cmd/server/server.go` | サーバー本体・各コマンド | 書いた |
| 8 | `combat.go` `hazard.go` `hardcore.go` | 戦闘・危険・死亡 | 書いた |
| 9 | `quest.go` `endings.go` `odyssey.go` | クエスト・エンディング | 書いた |
| 10 | `chat.go` `group.go` | チャット・グループ | 書いた |
| 11 | `notify.go` `flavor.go` | 通知文・実況文 | 書いた |
| 12 | `player_store.go` `item_store.go` | セーブ | 書いた |
| 13 | `logging.go` | ログ・不正検知 | 書いた |
| 14 | `cmd/cli/main.go` | CLIクライアント | 書いた |
| 15 | `cmd/gui/*.go` | GUIクライアント | 書いた |

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

# 第2章 `cmd/server/room.go`(16行)

**役割**:「部屋」を表すデータの型を定義するだけのファイルです。処理(関数)は1つもありません。

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
	Hazard      *RoomHazard       `json:"hazard,omitempty"`
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
	Hazard      *RoomHazard       `json:"hazard,omitempty"`
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
  - `MOVE east` と打ったとき、サーバーは `room.Exits["east"]` で行き先を引きます。
- `Hazard *RoomHazard`
  - 部屋の危険です。型の前の `*` は**ポインタ**(「`RoomHazard` そのものではなく、その置き場所を指す矢印」)です。
  - ポインタには「**何も指していない状態 = `nil`**」があります。そのため「**危険が無い部屋は `nil`**」と表せます。
  - ポインタにせず `RoomHazard` をそのまま持つと、危険の無い部屋も「種類が空の危険」を持つことになり、「あるか無いか」を区別しにくくなります。
  - `omitempty` は、`nil` のときJSONに書き出さない、という意味にもなります。
  - 使う側は `if room.Hazard == nil { 危険なし }` と書きます(`hazard.go` で出てきます)。

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

# 第3章 `cmd/server/player.go`(88行)

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
	Name            string                  `json:"name"`
	HP              int                     `json:"hp"`
	RoomID          string                  `json:"room_id"`
	Inventory       []string                `json:"inventory"`
```

- `Name`:プレイヤー名。
- `HP`:今の体力。
- `RoomID`:今いる部屋のID。
- `Inventory []string`
  - 持ち物のIDの一覧です。`[]string` は**スライス**(長さが変えられる配列)で「文字列の並び」です。
  - 例:`["item.olive_stake", "item.beeswax"]`。

```go
	Crew            int                     `json:"crew,omitempty"`
	CrewInitialized bool                    `json:"crew_initialized,omitempty"`
	IntroSeen       bool                    `json:"intro_seen,omitempty"`
	CombatTargetID  string                  `json:"combat_target_id,omitempty"`
```

- `Crew`:オデュッセイア編の「仲間の人数」。
- `CrewInitialized`:仲間を付与済みか。`bool` は `true`/`false` の型です。
- `IntroSeen`:導入の案内文をもう見たか。
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
	lastRegen       time.Time
	exiting         bool
}
```

- `lastRegen time.Time`
  - 最後にHP回復を計算した時刻です。
- `exiting bool`
  - 退室処理の最中かどうかです。
- この2つは**名前が小文字で始まる**ので、JSONに変換されません(=セーブされません)。しかも `json:"..."` のタグもありません。メモリ上だけで使う一時的な状態です。

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
func (p *Player) regenLocked(now time.Time) {
```

- 名前の最後の `Locked` は、このプロジェクトの決まりで「**呼ぶ前にサーバーのロック(`s.mu`)を取っていること**」を意味します(並行処理の話は第7章で詳しく)。
- 引数 `now` は「今の時刻」です。時刻を引数で受け取る形にすると、テストで好きな時刻を渡せます。

```go
	if p.HP >= maxPlayerHP || p.lastRegen.IsZero() {
		p.lastRegen = now
		return
	}
```

- `maxPlayerHP` は `server.go` で定義された定数(100)です。
- 「すでに満タン」または「`lastRegen` が一度も設定されていない(`IsZero()`)」とき:
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
	p.HP += ticks * regenAmount
	if p.HP >= maxPlayerHP {
		p.HP = maxPlayerHP
		p.lastRegen = now
		return
	}
	p.lastRegen = p.lastRegen.Add(time.Duration(ticks) * regenInterval)
}
```

- `p.HP += ticks * regenAmount`
  - 経過した回数分だけ回復します(`+=` は「足して代入」)。
- 上限を超えたら100に切り詰めて、基準時刻を今にして終わります。
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

# 第5章 `cmd/server/world.go`(226行)

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

	HomeRoomID string `json:"-"`
}
```

- `Name` / `Description`:名前と説明(言語別)。
- `RoomID`:**今、そのアイテムがある部屋のID**。誰かが拾うと `""` になる(どの部屋にも無い)。
- `Obtainable`:`TAKE` で取れるか。
- `Renewable`:`true` なら**取っても部屋から無くならない**(何度でも手に入る)。
- `RewardOnly`:`true` なら、エンディングの報酬としてだけ手に入る記念品(部屋には置かれない)。
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
```

- `Role`:役割。`"enemy"`(戦える敵)、`"quest_giver"`、`"dialogue"` のどれか。
- `RoomID`:いる部屋。
- `HP`:敵の初期HP。
- `Dialogue []LocalizedText`:**台詞のリスト**。`[]LocalizedText` は「LocalizedTextのスライス」。`TALK` では先頭(`Dialogue[0]`)を返し、ガイドNPCは続きも順に送る。

```go
	MythRequirementItem  string          `json:"myth_requirement_item,omitempty"`
	MythRequirementQuest string          `json:"myth_requirement_quest,omitempty"`
	FleeAccurate         bool            `json:"flee_accurate,omitempty"`
	FleeSucceedsOnce     bool            `json:"flee_succeeds_once,omitempty"`
	Unwinnable           bool            `json:"unwinnable,omitempty"`
	CrewLossOnAttack     int             `json:"crew_loss_on_attack,omitempty"`
	Guide                bool            `json:"guide,omitempty"`
	Ending               *Ending         `json:"ending,omitempty"`
}
```

- `MythRequirementItem` / `MythRequirementQuest`:この敵を攻撃(や会話)する前に**必要なアイテム/達成済みクエスト**。無いまま挑むと即死(ゲームの核)。
- `FleeAccurate`:`FLEE` が**必ず成功**する。
- `FleeSucceedsOnce`:`FLEE` が**最初の1回だけ成功**する。
- `Unwinnable`:**勝てない敵**。攻撃すると仲間が減るだけ。
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
- `QuestReward`(報酬):`HP` は回復量。
- `Quest`:`GiverNPCID` は依頼者NPCのID。`Objective` と `Reward` は**構造体を丸ごと中に持つ**(ポインタではない)。

## 5-5 `World`(世界全体)

```go
type World struct {
	StartRoomID string            `json:"start_room_id"`
	Rooms       map[string]*Room  `json:"rooms"`
	Items       map[string]*Item  `json:"items"`
	NPCs        map[string]*NPC   `json:"npcs"`
	Quests      map[string]*Quest `json:"quests"`
}
```

- `StartRoomID`:新規プレイヤーが始まる部屋。
- 4つの辞書:**IDから実体を引く**。`Rooms["loc.hall_of_fates"]` で運命の間の `*Room` が得られる。値が**ポインタ**なので、取り出して書き換えれば辞書の中身も変わる(アイテムの `RoomID` を書き換える処理で使う)。
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
	return w.validateEndings()
}
```

- 依頼者NPCが実在するか。目標の個数が1以上か。
- 目的の種類によって、**対象がアイテムかNPCか**を使い分けて実在確認する。
- 最後の `return w.validateEndings()`:エンディング定義の検査(第9章)の結果をそのまま返す。ここまで問題が無ければ `nil`。

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

# 第7章 `cmd/server/server.go`(841行)

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
  - 体力の上限100、復活時20。
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
	s.players[name] = player
	return nil
}
```

- **第3段階:登録**。
- `defer s.mu.Unlock()`:関数を抜けるとき自動で鍵を返す。以降どこで `return` してもよい。
- **もう一度、重複を確認する**。第1段階とここまでの間は鍵を外していたので、別の接続が同じ名前で先に入った可能性がある。 (「**二重チェック**」と呼ばれる定番の書き方)
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
	player.regenLocked(time.Now())
	return player
}
```

- 状態を変えるコマンドの先頭で使う。**退室中の人は `nil` 扱い**にして、 `regenLocked` で**HP自動回復を反映**してから返す。
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
	}{room, players, items, npcs})
```

- **無名の構造体**(名前を付けない型)をその場で定義して、すぐ値を入れてJSONにする。その場限りの形なので、型に名前を付けるまでもない。
- `}{room, players, items, npcs}`:フィールドの**定義順**に値を渡している。
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
	destination, ok := room.Exits[strings.ToLower(parts[1])]
	if !ok {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 301 NO_EXIT")
		return false
	}
	if s.world.Rooms[destination] == nil { ...500... }
```

- `room.Exits[方向]` で行き先を引く。方向は**小文字にしてから**引く(`NORTH` でも通る)。
- `destination, ok := ...`:辞書の2値形式。 `ok` が `false` = その方向に出口が無い → `301 NO_EXIT`。
- 行き先の部屋が世界に無ければ(データ異常)`500`。

```go
	if _, blocker := s.blockingEnemyLocked(player, player.RoomID); blocker != nil {
		encounterRoomID := player.RoomID
		locale := clientLocale(conn)
		client := conn.(*serverClient)
		response, err := client.enqueueResponse("OK room=" + destination)
		if err == nil {
			s.respawnPlayerLocked(player, *name, "slip_past", blocker.Name.Get(locale))
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
- `_, blocker := ...`:戻り値のうちIDは使わないので捨てる。 `blocker != nil` なら敵がいる。
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
			s.respawnPlayerLocked(player, *name, "talk_unprepared", npc.Name.Get(locale))
			s.broadcastFlavorLocked(encounterRoomID, flavor{key: "talk_unprepared", player: *name, npc: npc})
		}
		s.mu.Unlock()
		...
	}
```

- **敵ではないNPC**に神話の前提条件があり、満たしていないときは、**話しかけただけで死亡**する。応答は `OK dead`。

```go
	dialogue := ""
	if len(npc.Dialogue) > 0 {
		dialogue = npc.Dialogue[0].Get(locale)
	}
	if dialogue == "" || strings.TrimSpace(dialogue) == "" || !utf8.ValidString(dialogue) ||
		strings.IndexFunc(dialogue, unicode.IsControl) >= 0 || len("OK ")+len(dialogue) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
```

- 台詞の先頭(`Dialogue[0]`)をその言語で取り出す。
- **台詞が使えない場合は `500`**:空・空白だけ・不正なUTF-8・制御文字を含む・長すぎる。データの不備でプロトコルが壊れないための安全装置。

```go
	response, err := client.enqueueResponse("OK " + dialogue)
	if err == nil {
		logger.Info("npc_interaction", "player", *name, "npc", npcID, "room", player.RoomID)
		if npc.Guide {

			s.sendGuideLocked(*name, 1)
		}
		s.sendQuestHintLocked(player, npcID)
		s.talkEndingLocked(player, npc)
	}
```

- 応答は `OK 台詞`。
- そのあと:
  - **案内役**なら台詞の続き(2つ目以降)を通知として順に送る(第11章)。
  - **依頼者**なら、クエストの案内や進捗を通知(第9章)。
  - **エンディング**のあるNPCなら、エンディングの判定(第9章)。

## 7-20 `handleStatus`

```go
	player.regenLocked(time.Now())
	status := "healthy"
	if player.CombatTargetID != "" {
		status = "combat"
	}
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		MaxHP  int    `json:"max_hp"`
		Status string `json:"status"`
	}{player.HP, maxPlayerHP, status})
```

- ここだけ `playerForUpdateLocked` ではなく、 `s.players[*name]` で取って**直接** `regenLocked` を呼ぶ(`STATUS` は退室中でも答えてよいため)。
- 戦闘中(`CombatTargetID` が空でない)なら `"combat"`、そうでなければ `"healthy"`。
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
		s.respawnPlayerLocked(player, name, "hazard_lethal", roomName)
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
		s.respawnPlayerLocked(player, name, "hazard_item", roomName)
		return &flavor{key: "hazard_item", player: name, room: room}
```

- **`item_gate`**:必要アイテムを**持っていれば何も起きない**(`nil`)。持っていなければ死亡。

```go
	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name, "hazard_crew", roomName)
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

## 8-2 `hardcore.go`(119行)— 仲間の効果と、死亡のペナルティ

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
	player.Inventory = kept
	var returned []string
	for _, itemID := range lost {
		if !s.world.Items[itemID].Renewable {
			returned = append(returned, itemID)
		}
	}
	s.returnItemsHomeLocked(name, returned)
	return outcomeLost
}
```

- 持ち物を「残るもの」だけにする。
- 失ったもののうち**再生しないもの**を、元の置き場所に戻す(再生アイテムは部屋に残っているので戻す必要が無い)。
- `outcomeLost` を返す。

```go
func (s *Server) returnItemsHomeLocked(name string, itemIDs []string) {
	if len(itemIDs) == 0 {
		return
	}
	for _, itemID := range itemIDs {
		item := s.world.Items[itemID]
		item.RoomID = item.HomeRoomID
		if item.RoomID == "" {
			item.RoomID = s.world.StartRoomID
		}
		delete(s.unsavedTakes[name], itemID)
	}
```

- 失ったアイテムを**元の部屋**(`HomeRoomID`)に戻す。元が無い(報酬アイテムなど)なら開始部屋に。
- 未保存の取得記録からも消す。 `delete` は、 **辞書が `nil` でも安全**に呼べる。

```go
	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	locations, err := s.loadItemLocations()
	if err != nil {
		logger.Error("return_lost_items_failed", "player", name, "error", err.Error())
		return
	}
	for _, itemID := range itemIDs {
		locations[itemID] = s.world.Items[itemID].RoomID
	}
	if err := s.writeItemLocations(locations); err != nil {
		logger.Error("save_item_locations_failed", "player", name, "error", err.Error())
	}
}
```

- アイテム位置のファイル(`itemdata.json`)にも反映する(再起動後も元の部屋にあるように)。
- ここの `defer s.ioMu.Unlock()` で、どこで `return` してもファイルの鍵を返す。
- 失敗してもゲームは続けたいので、**エラーはログに残すだけ**。

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

## 8-3 `combat.go`(228行) 前半 — 乱数と死亡処理

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

## 8-4 `respawnPlayerLocked`(死亡と復活)

```go
func (s *Server) respawnPlayerLocked(player *Player, name, cause string, args ...any) {
	outcome := s.applyDeathPenaltyLocked(player, name)
	s.notifyDeathLocked(name, cause, outcome, args...)
	oldRoomID := player.RoomID
	logger.Info("player_died", "player", name, "cause", cause, "room", oldRoomID, "belongings", string(outcome))
```

- `args ...any`:**可変長引数**。 `...any` は「任意の型の値を何個でも」。 死因ごとに必要な情報(敵の名前など)が違うので、このように受け取る。渡すときも `args...` と展開して `notifyDeathLocked` に渡す。
- 順番:① 持ち物のペナルティ(8-2) → ② 本人へ死亡メッセージを通知 → ③ ログ。

```go
	destination := defaultStartRoomID
	if s.world != nil {
		destination = s.world.StartRoomID
	}
	player.HP = respawnHP
	player.CombatTargetID = ""
	player.RoomID = destination
	if oldRoomID == destination {
		return
	}
```

- 復活先は開始部屋。HPは20、戦闘状態を解除。
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

- `ATTACK` の応答JSONの形(RFCで決まっている)。

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
	if npc.Role != "enemy" || enemyHP <= 0 {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 405 NPC_NOT_HOSTILE")
		return false
	}
	encounterRoomID := player.RoomID
```

- いつもの骨格(7-10)の前半。NPCを特定する。
- **敵でない、またはこの人にとってすでに倒されている**なら `405 NPC_NOT_HOSTILE`。
- `encounterRoomID`:戦った部屋を控える(死んで部屋が変わっても、**戦った部屋に実況を送る**ため)。

```go
	var event flavor
	var result combatResult

	switch {
```

- あとで埋める2つの変数:実況データと応答の中身。
- `switch {` で、上から順に条件を見る。

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
		s.respawnPlayerLocked(player, *name, "attack_unprepared", npc.Name.Get(locale))
		event = flavor{key: "attack_unprepared", player: *name, npc: npc}
		result = combatResult{0, enemyHP, 0, "dead"}
```

- 必要なアイテムやクエストが無いのに攻撃すると、**即死**。ゲームの核心。
- 応答の `attacker_hp` は0(死んだので)。復活後のHP20とは別に、結果として0を返している。

### ケース3:通常の戦闘

```go
	default:

		allies := s.alliesInRoomLocked(*name)
		bonus := allyBonusCount(allies)
		damage := randDamage(combatMinDamage, combatMaxDamage) + bonus*allyDamageBonus
		enemyHP -= damage
		if enemyHP < 0 {
			enemyHP = 0
		}
		player.setEnemyHP(npcID, enemyHP)
```

- 同室の仲間の数(最大3)で `bonus` を決める。
- 与ダメージ = 8〜14の乱数 + 仲間1人につき5。
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

- **倒した場合**:戦闘解除、 「敵を倒す」クエストを進行、仲間に撃破を共有。 `status` は `"victory"`。

```go
		} else {
			player.CombatTargetID = npcID
			counter := randDamage(counterMinDamage, counterMaxDamage)
			counter = max(1, counter*(100-bonus*allyCounterReductionPercent)/100)
			player.HP -= counter
```

- **倒せなかった場合**:戦闘中の相手を記録し、反撃を受ける。
- 反撃のダメージは7〜14。仲間1人につき20%軽減。 `counter * (100 - 20×人数) / 100` で割合を掛ける(整数の計算)。
- `max(1, ...)`:**最低でも1ダメージ**を受ける(組み込みの `max` 関数)。

```go
			if player.HP <= 0 {
				s.respawnPlayerLocked(player, *name, "attack_counter", npc.Name.Get(locale))
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

## 8-6 `handleFlee`(逃げる)

```go
	targetID := player.CombatTargetID
	npc := s.world.NPCs[targetID]
	if targetID == "" || npc == nil {

		player.CombatTargetID = ""
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
		if player.FledFrom == nil {
			player.FledFrom = make(map[string]bool)
		}
		player.FledFrom[targetID] = true
	} else {
		counter := randDamage(counterMinDamage, counterMaxDamage)
		player.HP -= counter
		if player.HP <= 0 {
			s.respawnPlayerLocked(player, *name, "flee_failed", npc.Name.Get(locale))
			result = "failure_dead"
			event = flavor{key: "flee_dead", player: *name, npc: npc}
		} else {
			result = "failure"
			event = flavor{key: "flee_hit", player: *name, npc: npc, n: counter}
		}
	}
```

- **成功**:戦闘解除、 `FledFrom[敵] = true` と記録(以後その敵は出口を塞がない)。 辞書が `nil` なら先に `make`。
- **失敗**:反撃を受ける。HPが尽きれば死亡。

```go
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		Result string `json:"result"`
	}{player.HP, result})
```

- 応答は `OK {"hp":..,"result":"success|failure|failure_dead"}`。残りは骨格どおり。

---

# 第9章 クエスト・エンディング・オデュッセイア編(`quest.go`, `endings.go`, `odyssey.go`)

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
		s.respawnPlayerLocked(player, name, "lotus")
		return &flavor{key: "lotus", player: name}
	case itemSacredCattle:
		s.respawnPlayerLocked(player, name, "cattle")
		return &flavor{key: "cattle", player: name}
	}
	return nil
}
```

- **拾った結果の効果**:ロトスの実(故郷を忘れる)や聖なる牛(ヘリオスの怒り)は、**手に取った瞬間に死亡**する(神話の通り)。
- `respawnPlayerLocked(player, name, "lotus")`:死因 `"lotus"` を渡す。 `args` は省略(可変長引数なので、渡さなくてよい)。
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

## 9-2 `quest.go`(235行)— クエスト

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
	logger.Info("quest_completed", "player", player.Name, "quest", questID, "reward_hp", quest.Reward.HP)
	player.HP += quest.Reward.HP
	if player.HP > maxPlayerHP {
		player.HP = maxPlayerHP
	}
	s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{ ... }.Format(locale, quest.Name.Get(locale), quest.Reward.HP, player.HP))
}
```

- 達したら `"completed"` にして、 **報酬のHPを足す**(上限100まで)。
- 「クエスト達成! 報酬: HP+10(現在HP ◯)」と本人に通知。

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
		(「◯◯を頼みたい(報酬 HP+◯)。QUEST でうけられる」を通知)
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
	}{questID, quest.Description.Get(locale), quest.Reward.HP, "available"})
	...
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil && newlyAccepted && s.objectiveAlreadyMetLocked(player, quest) {

		s.advanceQuestLocked(player, questID, quest, player.Quests[questID])
	}
```

- 応答はRFCで決まった形 `{quest_id, description, reward, status}`。
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
		count := 1
		if quest := s.world.Quests[id]; quest != nil {
			count = quest.Objective.Count
		}
		list = append(list, questEntry{id, state.Status, fmt.Sprintf("%d/%d", state.Progress, count)})
	}
```

- 各クエストについて `{quest_id, status, progress: "2/3"}` を作る。
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

## 9-3 `endings.go`(190行)— エンディング

### 9-3-1 `Ending` の型

```go
type Ending struct {
	ID              string          `json:"id"`
	Name            LocalizedText   `json:"name"`
	RequiresItems   []string        `json:"requires_items,omitempty"`
	RequiresQuests  []string        `json:"requires_quests,omitempty"`
	RequiresEndings []string        `json:"requires_endings,omitempty"`
	RewardItem      string          `json:"reward_item,omitempty"`
	Hint            LocalizedText   `json:"hint"`
	Text            []LocalizedText `json:"text"`
}
```

- 各NPCに付けられるエンディングの定義。
- `RequiresItems` / `RequiresQuests` / `RequiresEndings`:迎えるために**持っているべきアイテム・達成すべきクエスト・先に迎えるべき別のエンディング**(最終エンディングは、3つの物語のエンディングを全部迎えることが条件)。
- `RewardItem`:報酬の記念品(アイテムID)。
- `Hint`:条件が足りないときに表示する文。
- `Text []LocalizedText`:エンディングの本文(複数行)。

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
		if e.RewardItem != "" {
			item := w.Items[e.RewardItem]
			if item == nil || !item.RewardOnly {
				return fmt.Errorf("ending %q reward %q must be an existing reward_only item", e.ID, e.RewardItem)
			}
		}
	}
```

- 本文があるか、必要なアイテム・クエストが実在するか、 **報酬が「報酬専用アイテム」であるか**を検査。

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
	s.sendEndingProgressLocked(player)
}
```

- **初めて迎える**:記録して報酬を付与し、ログに残す。
- 「=== エンディング: ◯◯ ===」の見出し → 本文 → 報酬の案内 → 全体の進捗、の順に通知する。

---

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

## 11-1 `notify.go`(148行)— 個人向けの通知

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

## 11-2 `flavor.go`(114行)— 部屋への実況

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
			recipient.enqueueEvent("EVT ROOM COMBAT " + f.text(s.localeOfLocked(playerName)))
		}
	}
}
```

- 指定の部屋にいる**全員**に、実況を送る。
- **受け取る人ごとに言語が違うかもしれない**ので、ループの中で `f.text(その人の言語)` を作る。日本語の人には日本語、英語の人には英語。

---

# 第12章 セーブ(`player_store.go`, `item_store.go`)

セーブは2つのJSONファイルです。

| ファイル | 中身 |
|---|---|
| `saves/playerdata.json` | 全プレイヤー(名前→`Player`) |
| `saves/itemdata.json` | 普通のアイテムの置き場所(アイテムID→部屋ID) |

## 12-1 `player_store.go`(109行)

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
	}
	return players, nil
}
```

- JSONを辞書に変換。
- ファイルの中身が `null` だと `players` が `nil` になるので、それは不正として弾く。
- 辞書のキーと、中の `Name` が**一致しているか**確認(手で書き換えて壊れたデータに気づくため)。

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
func (s *Server) loadItemLocations() (map[string]string, error) {
	...
}

func (s *Server) writeItemLocations(locations map[string]string) error {
	...
}
```

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
  10. `ui_bars.go` / `ui_combat.go`:HP・仲間のバーと戦闘パネル
  11. `ui_story.go` / `ui_effects.go`:色付きの冒険ログと画面のフラッシュ
  12. `ui_map.go` / `ui_endings.go`:地図とエンディング一覧
  13. `cmd/server/gui_state.go`:GUI向けSTATE拡張の応答

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

## 15-2 `model.go`(155行)— 受け取るデータの型

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

- サーバーが返す JSON を受け取るための型。 `LOOK`、 `STATUS`、 `STATE`、 `QUESTS` の応答に対応する。サーバー側で定義した形(第4章の `roomView` など)と同じ。
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

## 15-3 `retro.go`(90行)— 見た目

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
)
```

- 配色。ほぼ黒の紺(`ink`)、紺(`navy`)、象牙色(`ivory`)、金(`gold`)。 `R, G, B, A` は赤・緑・青・不透明度(0〜255)。

```go
type retroTheme struct{ base fyne.Theme }

func (t retroTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := storyColor(name); ok {
		return c
	}
	switch name {
	case theme.ColorNameBackground, theme.ColorNameMenuBackground:
		return ink
	...
	}
	return t.base.Color(name, theme.VariantDark)
}
```

- `fyne.Theme` という interface(`Color`、`Font`、`Icon`、`Size` の4メソッド)を**自分で実装**して、見た目を変える。
- 背景は `ink`、ボタンは `navy`、文字は `ivory`、強調は `gold`…と色を割り当て、決めていないものは元のテーマ(`t.base`)に任せる。
- `Font` は日本語フォントを返し、 `Size` は文字の大きさや余白を返す(角丸は0 = 四角く)。

```go
func framed(title string, content fyne.CanvasObject) fyne.CanvasObject {
	frame := canvas.NewRectangle(ink)
	frame.StrokeColor = ivory
	frame.StrokeWidth = 2
	caption := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	inside := container.NewBorder(caption, nil, nil, nil, content)
	return container.NewStack(frame, container.NewPadded(inside))
}
```

- **枠付きのパネル**を作る部品。四角形(枠)の上に、見出し+中身を重ねる。
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

## 15-5 `art_assets.go`(87行)— 画像

```go
const (
	artWidth     = 960
	artHeight    = 576
	spriteWidth  = 192
	spriteHeight = 312
)

//go:embed assets/rooms/*.png assets/npcs/*.png
var artAssets embed.FS
```

- 背景画像のサイズ(960×576)とNPC画像のサイズ(192×312)。
- `embed.FS`:埋め込んだファイルを**ファイルシステムのように**読める型。 `assets/rooms/*.png` のように `*` で複数をまとめて埋め込める。

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
func composeScene(roomID string, npcs []string) *image.RGBA {
	background := image.NewRGBA(image.Rect(0, 0, artWidth, artHeight))
	if room := loadArt("rooms", roomID); room != nil {
		if room.Bounds().Dx() == artWidth && room.Bounds().Dy() == artHeight {
			draw.Draw(background, background.Bounds(), room, room.Bounds().Min, draw.Src)
		} else {
			xdraw.CatmullRom.Scale(background, background.Bounds(), room, room.Bounds(), draw.Src, nil)
		}
	}
```

- **部屋の背景にNPCの絵を重ねて、1枚の絵を作る**。
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
		sprite := loadArt("npcs", id)
		if sprite == nil {
			continue
		}
		dest := image.Rect(x, y, x+width, y+height)
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
	application := app.New()
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
			words = ui.catalog.label("quest", quest.QuestID, ui.locale) + "\n" + quest.Description + fmt.Sprintf("\nHP +%d", quest.Reward)
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

## 15-8 `ui_layout.go`(398行)— 画面とウインドウサイズ

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
	ui.quitButton.Disable()
	ui.statusLabel = compactLabel(ui.tr("Not connected", "未接続"))
	ui.roomCount = compactLabel(ui.tr("Here: -", "部屋: - 人"))
	ui.totalCount = compactLabel(ui.tr("Online: -", "全体: - 人"))
	ui.hpLabel = compactLabel("HP: -")
	ui.crewLabel = compactLabel(ui.tr("Crew: -", "仲間: - 人"))
	ui.hpBar, ui.crewBar = newStatBar(), newStatBar()
	ui.groupLabel = compactLabel(ui.tr("Group: -", "グループ: -"))

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

## 15-10 `ui_journal.go`(238行)— まわり・持ち物・クエスト

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
	if ui.scene.Image == nil || previous.Room.ID != view.Room.ID || !slices.Equal(previous.NPCs, view.NPCs) {
		ui.scene.Resource = nil
		ui.scene.Image = composeScene(view.Room.ID, view.NPCs)
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
		if !known || ui.catalog.NPCs[id].Role == "enemy" {
			content.Add(container.NewGridWithColumns(2,
				ui.commandButton(ui.tr("Attack", "戦う"), func() { ui.send("ATTACK " + id) }),
				ui.commandButton(ui.tr("Flee", "逃げる"), func() { ui.send("FLEE") }),
			))
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

一覧の差し替えが終わったら、「まわり」全体のスクロール領域もRefreshします。子の一覧だけを更新すると、既に表示している親に古い高さが残り、接続直後の項目が重なることがありました。親まで更新することで、タブ切替やサイズ変更をしなくても初回のLOOKから正しい高さになります。

- 部屋IDが変われば古い選択画面を閉じます。部屋またはNPCが変わったときに背景を作り直し、道具が変わったときにアイテムの絵を更新します。移動先・NPC・道具・プレイヤーを操作ボタン付きで表示します。カタログで分かるNPCには対応する依頼・戦闘だけを表示し、未知のNPCでは各操作を残します。同じLOOKなら一覧やスクロール位置を保ちます。

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
	return journalCard(container.NewBorder(nil, nil, picture, button, journalName(ui.catalog.label("item", id, ui.locale))))
}
```

- 道具の小さな絵・表示名・取る／置くボタンをまとめます。表示は名前、送信はIDを使います。

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

## 15-11 `cmd/server/gui_state.go`— STATE拡張

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

## 15-12 戦闘・地図・演出と表示の検証

最新mainから取り込んだGUIの機能は次のファイルへ分かれています。

| ファイル | 担当する表示と処理 |
| --- | --- |
| `ui_bars.go` | HPと仲間のラベルの背景へ残量を描画する |
| `ui_combat.go` | 戦闘中の敵名・HP・「戦う／構える／逃げる」を表示し、ATTACK・DEFEND・FLEEの応答を反映する |
| `ui_story.go` | 冒険の履歴を種類付きで保存し、種類別の色を使ったRichTextへ変換する |
| `ui_effects.go` | 移動・被害・死亡時のフラッシュを作り、時間経過で透明にする |
| `ui_map.go` | 訪れた部屋と接続関係、危険・即死等の凡例を表示する。タブと絵の左の余白の描画部品を別々に更新し、左へ表示できる場合は地図タブを取り除く。狭い画面の地図タブは地図と凡例全体をスクロールできる |
| `ui_endings.go` | 持ち物の記念品から取得したエンディングと祝福を表示し、設定へ記録した即死部屋の数も表示する |

長い敵名、エンディングの祝福文にも`textVBox`を使います。地図の凡例は色見本以外へ残りの幅を割り当て、言語による文字幅の違いに対応します。

`ui_overlap_test.go`は日本語・英語、文字サイズ15・22、画面の横並び・縦並びの切替で、文章・ボタンの重なりと画像の枠外表示を確認します。4種類の一覧、戦闘中のパネル、ゲームオーバー画面、アイテム画像の縮小と1枚ずつのスクロールも確認します。実際の部屋・NPC・クエストを使い、長い日本語名を追加して折り返しを検証します。

`ui_initial_room_test.go`では、未接続の画面を表示し、英語から日本語へ切り替えた後に、最初の部屋を読み込みます。タブ切替・サイズ変更をしない状態の「まわり」の配置を検証し、その後タブを切り替えても一覧の高さが変わらないことを確認します。文字サイズ15・22、1280×900・640×900で確認します。

`ui_map_test.go`では、日本語・英語、文字サイズ15・22で、広い画面の地図が左の余白に収まり、右の地図タブを取り除くことを確認します。幅を狭めると同じ地図タブを戻すこと、持ち物・クエストの選択とチャットの下書きを保つこと、言語切替後も広い画面では地図タブを出さないことも確認します。2か所の地図は描画部品を共有せずに訪問済みの部屋を更新します。

Fyneのテスト用ウインドウは実ウインドウの最小サイズ制約を自動では適用しないため、テスト側で現在の最小サイズ以上へ調整します。`TAP_GUI_PREVIEW_DIR`に既存の出力ディレクトリを指定すると、同じ検証画面をPNGへ保存できます。通常のテストではPNGをファイルへ書きません。これはFyneのテスト用Canvasでの確認であり、OSの実ウインドウの起動確認とは別です。

# おわりに

ここまでで、サーバー(第1〜13章)・CLI(第14章)・GUI(第15章)のコードをすべて、上から順に読みました。

## 読み終わったあとの確認ポイント

評価で説明できるようにするため、次の問いに自分の言葉で答えられるか確認してください。

1. **接続が来てからコマンドが処理されるまでの流れ**を説明できる?
   → `Accept` → `go handleClient` → `Scan` で1行読む → `commandHandlers` から担当関数 → 応答を返す(第1章、7-21)
2. **遅いクライアントへの送信中に、共通ロックを保持し続けない理由**は?
   → 送信キューへ積み、共通ロックを外してから`waitResponse`で書き込み完了を待つ。実際のネットワーク書き込みは専用goroutineが行う(第6章、7-10)
3. **セーブが途中で壊れない理由**は?
   → 一時ファイルに書いて `Rename` で入れ替える(12-1)
4. **世界のデータを変えるときは、どこを触るか**?
   → `data/world.json` だけ。起動時に `validate` が矛盾を見つける(第5章)
5. **神話のゲート(即死)の仕組み**は?
   → `meetsMythRequirement`(3-5)と、 `handleAttack` / `handleTalk` の分岐(8-5, 7-19)
6. **新しいコマンドを足すには**?
   → `handleXxx` を書いて `commandHandlers` に1行足す(7-9)
