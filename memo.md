# memo.md — The Answer Protocol 知識・手法メモ

## 1. RFC 42TAP の要点(protocol-rfc.html を読んだ結果のまとめ)
RFC 2119キーワード(MUST/SHOULD/MAY)で書かれた本物のRFC形式。要点だけ抜粋。

### 状態遷移
`DISCONNECTED → CONNECTED(TCP確立、未認証) → AUTHENTICATED(CONNECTコマンド送信済み) → TERMINATED`

### 接続シーケンス
```
Server -> Client: OK hello proto=1        # 接続直後、サーバーから必ず送る
Client -> Server: CONNECT alice
Server -> Client: OK connected            # 成功時
Server -> Client: ERR 201 NAME_IN_USE     # 名前重複時
```

### メッセージ全体のABNF構造
```
message      = command-line / response-line / event-line
command-line = command-name [SP arguments] LF
response-line = ("OK" / error-response) [SP response-data] LF
event-line   = "EVT" SP event-type SP event-data LF
error-response = "ERR" SP error-code SP error-message   ; error-code = 3DIGIT
```
→ **サーバーからの応答は必ず OK/ERR で始まる。イベント(非同期通知)は必ず EVT で始まる**という3分類を守るのが実装の骨格。

### コマンド一覧と代表レスポンス
| コマンド | 用途 | 主な成功レスポンス | 主なエラー |
|---|---|---|---|
| CONNECT <name> | 認証 | OK connected | ERR 201 NAME_IN_USE |
| LOOK | 現在地情報取得(JSON) | OK {room, players, items, npcs} | - |
| MOVE <dir> | 移動 | OK room=<id> | ERR 301 NO_EXIT |
| QUIT | 切断 | OK bye | - |
| CHAT <scope> <msg> | GLOBAL/ROOM/GROUPへ発言 | OK | - |
| WHO | 人数照会 | OK players=<n> | - |
| GROUP CREATE/INVITE/JOIN/LEAVE | グループ管理 | OK group=<id> 等 | ERR 401/402 |
| TAKE <item> | 取得 | OK taken=<id> | ERR 404 ITEM_NOT_FOUND |
| DROP <item> | 手放す | OK dropped=<id> | ERR 404 ITEM_NOT_IN_INVENTORY |
| INVENTORY | 所持品一覧(JSON配列) | OK [...] | - |
| TALK <npc> | NPC会話 | OK <dialogue text> | ERR 404 NPC_NOT_FOUND |
| ATTACK <npc> | 戦闘開始 | OK {attacker_hp, target_hp, damage, status} | ERR 404 / 405 NPC_NOT_HOSTILE |
| STATUS | 自分のHP確認 | OK {hp, max_hp, status} | - |
| QUEST <npc> | クエスト受注 | OK {quest_id, description, reward, status} | ERR 404 / 406 NO_QUEST_AVAILABLE |
| QUESTS | 自分のクエスト一覧 | OK [{quest_id, status, progress}, ...] | - |

### イベント一覧(サーバー→クライアント、非同期プッシュ)
- `EVT ROOM PRESENCE ENTER <name>` / `EVT ROOM PRESENCE LEAVE <name>`
- `EVT ROOM CHAT <sender> <msg>`
- `EVT GLOBAL CHAT <sender> <msg>`
- `EVT GROUP INVITE|JOIN|LEAVE|CHAT ...`
- `EVT STATS players=<n>`

### エラーコード一覧
| コード | 意味 |
|---|---|
| 201 | NAME_IN_USE |
| 301 | NO_EXIT |
| 401 | NOT_IN_GROUP |
| 402 | ALREADY_IN_GROUP |
| 404 | ITEM_NOT_FOUND / ITEM_NOT_IN_INVENTORY / NPC_NOT_FOUND(文脈で使い分け) |
| 405 | NPC_NOT_HOSTILE |
| 406 | NO_QUEST_AVAILABLE |
| 900 | CONNECTION_FAILED |
| 901 | SEND_FAILED |
4xxは継続可能なエラー、9xxは再接続が必要になりうるエラー、という位置づけ。

### 意図的に未定義(=チームで設計してREADMEに正当化を書く箇所)
- 戦闘: ターン管理・初期先攻順、ダメージ計算式、戦闘の開始/終了/状態遷移、DEFEND/FLEE/USE_ITEM等の追加コマンド、状態異常(毒・バフ・デバフ)
- クエスト: 進行状況の追跡・検証方法、自動/手動の完了判定、報酬の付与方法、前提クエスト(クエストチェーン)、COMPLETE_QUEST/ABANDON_QUEST等の追加コマンド

### セキュリティ/堅牢性の要求(9章)
- **メッセージ分割**: 1つのTCPパケットに1コマンドが収まる保証はない。サーバーは不完全な行をバッファし、`\n`が来るまで待つ必要がある(=TCPは「メッセージ境界を保証しないバイトストリーム」であることを実装で意識する)。
- **メッセージ結合**: 逆に1パケットに複数行が来ることもある。行ごとに分割して個別処理する必要がある。
- UTF-8を正しく扱う(ユーザー名・メッセージ中のマルチバイト文字でクラッシュしない)。
- 制御文字(ターミナル操作系のエスケープシーケンス等)は拒否 or 安全に処理(ターミナルインジェクション対策)。
- 入力バリデーション必須(インジェクション対策)。
- 推奨リソース上限: 1行あたり1024バイト、接続数、プレイヤーあたりの所持品数、チャット頻度。

## 2. TCPソケットプログラミングの基礎(言語別ヒント)
TCPはストリーム指向で「メッセージの境界」を保証しない。**必ず行バッファリング(`\n`が来るまで蓄積)を自分で書く**必要がある。
- **C**: `socket()/bind()/listen()/accept()` + `select()`か`poll()`/`epoll()`(Linux)でマルチクライアント処理。もしくはクライアントごとに`pthread`。
- **C++**: 生ソケットAPIを使うか、Boost.Asioで非同期I/O。
- **Go**: `net.Listen("tcp", ...)` + `go func(){ handleConn(conn) }()` でgoroutine1本/クライアントが素直。`bufio.Scanner`で行読み込みが楽。
- **Rust**: `std::net::TcpListener` + スレッド、または`tokio`で非同期(async/await)。`tokio::io::BufReader::lines()`で行読み込み。
- **Zig**: `std.net.StreamServer`(バージョンによりAPI名が変わるので使用するZigのバージョンのドキュメントを都度確認)。

並行モデルの選択(要README説明):
- スレッド/goroutineベース: 実装がシンプル、大量接続でリソースを食う
- イベントループ(epoll/kqueue/select、または非同期ランタイム): スケールするが実装が複雑

## 3. ワールドデータ(YAML/JSON)のパース
- Go: `gopkg.in/yaml.v3`, `encoding/json`
- Rust: `serde_yaml` / `serde_json` + `serde::Deserialize`
- C++: `yaml-cpp`, `nlohmann/json`
- C: `libyaml`, `cJSON`(手作業が多くなりがち)
- Zig: 標準の`std.json`。YAMLは外部ライブラリが少ないのでJSON採用も検討。

## 4. GUIクライアントのツールキット選択肢
- Qt(C++、`QTcpSocket`でネットワークも統合しやすい)
- GTK+(C/C++、GTK4推奨)
- Rust: `egui`(即時モードGUI、実装が速い)、`iced`
- Go: `Fyne`, `gioui`
- Web系(Electron等、"web-based"は許可されているがcursesはNG)
- **注意**: curses/ncursesはテキストUI扱いで「GUIではない」と明記されているので、GUIクライアントには使えない(CLIクライアント側でリッチ表示に使うのはOK)。

## 5. 構造化ログの実装
- Go: `log/slog`(標準ライブラリ、JSON出力対応)
- Rust: `tracing` + `tracing-subscriber`(JSON layer)
- C++: `spdlog`(JSON sinkを組み合わせ)
- C: 自作のprintf系JSON整形、または`json-c`で組み立ててから出力

## 6. 開発の勧め方
1. まずRFCの状態遷移と3種のメッセージ形式(command/response/event)をそのまま素朴に実装し、`nc`(netcat)や`telnet`で手動接続してCONNECT→LOOKが通ることを確認する。
2. 行バッファリング(分割/結合パケット耐性)を早い段階でテストする(意図的に1文字ずつ送るテストスクリプトなどで確認すると良い)。
3. コマンドを1つずつ実装し、都度エラーコードの妥当性を確認。
4. 戦闘・クエストの設計は最後に固めても良いが、README用の「設計判断とその理由」は開発しながらメモを取っておくと後で楽。
5. 他チームとの相互接続テストを早めに一度やっておく(プロトコル解釈のズレを早期発見できる)。
