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

## 7. ゲームシステム設計(戦闘・クエスト固有ギミック)
RFC 6.1.1/6.1.2で「未定義=チームが決めてREADMEで正当化する部分」とされている、戦闘・クエストの中身の設計方針。**2026-09-29時点で実装完了**: `combat.go`(ATTACK/FLEE)、`quest.go`(QUEST/QUESTS)、`hazard.go`(部屋ハザード)、`odyssey.go`(オデュッセイア固有のアイテム処理)。3部作(アルゴナウタイ/トロイア/オデュッセイア)全編の神話ゲートと`world.json`の日本語訳(後述8章)も実装済み。

### 7.1 コンセプト
**「神話に忠実な行動を取らないと死ぬ」**を戦闘・クエスト設計の軸にする。単純なHP削り合いではなく、各NPC・各場面ごとに神話上の正しい対策(アイテム所持・事前クエスト達成・正しい選択)を要求し、満たしていなければ即死させる「神話ゲート」を仕込む。

### 7.2 基本メカニクス(3層)
1. **戦闘ゲート**: `role: "enemy"` のNPCに `myth_requirement`(必要アイテム/達成済みクエスト)を持たせる。ATTACK時に満たしていなければ即死(HP→0、リスポーン処理へ)。満たしていれば通常の数値戦闘(固定ダメージ+小さな乱数、プレイヤー→NPCの順で攻撃、NPC生存時はカウンター)。
2. **FLEEの神話整合性**: NPCごとに `flee_is_myth_accurate: true/false` を持たせる。神話で「退いた/逃げた」話ならFLEE成功、「正面から戦い切った」話ならFLEE失敗(反撃を受ける)。
3. **経路ハザード**: MOVEの出口(exit)に危険度を持たせる。`hazard: "lethal"`(即死)、`hazard: "crew_loss"`(乗組員が代わりに死ぬ、プレイヤー本人は無傷)など。

### 7.3 神話ゲート一覧(戦闘)
| NPC/場面 | 必要条件 | 未達成時 | FLEE |
|---|---|---|---|
| 青銅の雄牛 (khalkotauroi) | `item.medeas_ointment` 所持 | 即死(焼かれる) | 失敗 |
| コルキスの竜 (colchis_dragon) | `item.medeas_draught` 所持 | 即死(丸呑み) | 失敗 |
| タロス (talos) | `quest.golden_fleece` 達成済み(メデイアの魔術による援護、`myth_requirement_quest`で判定) | 即死(投石) | 失敗 |
| ポリュペモス (polyphemus) | `item.olive_stake` 所持 | 即死(食われる) | **成功**(羊の下に隠れて脱出) |
| ライストリュゴネス族 | なし(対抗不可) | ATTACKは常に即死→代わりに**乗組員が死ぬ**扱いに変更予定(7.5参照) | **成功**(オデュッセウスの船だけ湾外に停泊、これが正解) |
| ヘクトール (hector) | `item.shield_of_achilles` 所持 | 即死 | **そのNPCから初めて逃げた1回だけ成功**(3周城壁を逃げた話を再現、`Player.FledFrom`で永続管理)、以降は何度再戦しても失敗(アテナに唆され引き戻される) |
| 求婚者たち (antinous) | `item.odysseus_bow` 所持(弓を張った状態=`quest.string_the_bow`達成) | 即死(素手で100人は不可能) | 失敗 |
| アミュコス (amycus) / ハルピュイア (harpy) | なし(通常戦闘) | — | 失敗 |

### 7.4 神話ゲート一覧(滞在・対話型、ATTACK以外)
LOOK/TALKなど「その場にいる/話す」だけで発動する危険。ATTACK系とは別枠の第4の仕組みとして扱う。
- **セイレーン** (`loc.ody_sirens`): `item.beeswax`(**新規アイテム、要追加**)所持が必須。未所持でLOOK/TALK等その場に留まると即死(歌に誘われ入水)。所持していれば安全に通過でき、歌の演出(フレーバーテキスト)も見られるボーナス扱い。
- **キルケー** (`npc.circe`): `item.moly` 所持が必須(既存アイテム、`quest.moly_herb`と連動)。未所持でTALK/滞在すると豚に変えられ即死扱い。所持していれば安全に進行し、ダイアログで「彼女に誓いを立てさせた」というフレーバーを入れる(神話でヘルメスがオデュッセウスに授けた助言が「傷つけないと誓わせろ」だったことに準拠)。

### 7.5 経路ハザード: スキュラとカリュブディス
現状 `loc.ody_sirens` → `east` → `loc.ody_scylla` の一本道になっているが、これを分岐させる。
- `loc.ody_sirens` → `east`: 既存の `loc.ody_scylla`。通過時に**乗組員6人を自動で失う**(プレイヤーのHPは無傷)。通過には後述のcrew閾値判定が必要。
- `loc.ody_sirens` → `south`(**新規**): `loc.ody_charybdis`(**新規部屋、要追加**)。進入した瞬間ほぼ確実に即死。安全策は用意しない(神話通り、カリュブディス側に生還の道はない)。

### 7.6 乗組員(crew)リソースシステム
NPCではなく「プレイヤーに同行する乗組員」を一種のリソースとして持たせる、オデュッセイア編限定の仕組み。
- Player構造体に `Crew int` を追加(サーバー内部状態のみ)。
- オデュッセイア編開始地点 `loc.ody_troy_shore` でCrew初期値 **12**をセット。
- 各イベントでcrewが増減(たたき台、要詰め):
  - キコネス族の襲撃で略奪に長居: **-2**
  - ポリュペモスの洞窟突入時点で: **-2**(固定、既存dialogueの「もう2人食われた」を反映)
  - ライストリュゴネス族: ATTACK(誤答)で**-8**、FLEE(正解)で**±0**
  - ロトパゴイ: 新規アイテム`item.lotus_fruit`(`loc.ody_lotus`に配置、入手可能)をTAKE(食べる)すると**-2**(捜索に出した乗組員も蜜に絆され、引き戻すのに手間取り数名失う)。TAKEしなければノーコスト。
  - アイオロスの風袋: 既存`item.bag_of_winds`を`loc.ody_ithaca_shore`以外の場所でDROPすると**-3**(早まって開封→嵐で押し戻される)。イタケの浜に着くまで持ち続ければノーコスト。新規コマンドは追加せず既存DROPを流用する。
  - スキュラ通過時: 自動で**-6**(必須、避けられない)
- **Scylla通過条件**: 通過直前の時点で `Crew + 1(プレイヤー自身) >= 7` でなければ、6人を失いきれず**プレイヤーも即死**扱いにする。

### 7.7 RFCプロトコル互換性への配慮
RFCはJSONレスポンスへの独自フィールド追加を明示的に禁止も許可もしていない(グレーゾーン)。他チームのクライアントの厳密なJSONパーサーを壊すリスクを避けるため、**`crew`のような独自ゲーム状態はSTATUS/ATTACK等RFC規定のJSON構造には含めない**方針とする。crewの増減は以下のみで表現する:
- サーバー内部の`Player.Crew`フィールド(判定用)
- LOOK/CHAT/EVTのフリーテキストでのブロードキャスト(例: `"Scylla's six heads snatch six of your crew from the deck."`)
- 構造化ログ(JSON)への記録

### 7.8 クエスト完了判定
- `world.json`の各クエストの`objective`(`collect_item`/`defeat_npc`)を、TAKE/ATTACK成功のタイミングでサーバー側が自動照合し、達成なら自動で`reward`(HP回復)を付与する(手動COMPLETE_QUESTコマンドは追加しない)。
- クエストの内容自体が「神話上の正しい行動」と一致するよう設計する(例: `quest.blind_the_cyclops`の裏でolive_stake所持がATTACK生存条件になっている、など7.3と連動)。

### 7.9 world.jsonに追加したデータ(実装済み)
- アイテム追加: `item.medeas_ointment`(雄牛の火傷除け)、`item.beeswax`(セイレーン用耳栓)、`item.lotus_fruit`(`loc.ody_lotus`、obtainable)
- 部屋追加: `loc.ody_charybdis`
- `loc.ody_sirens`のexitsに`south`(→`loc.ody_charybdis`)を追加
- `Player.Crew`、`Player.CombatTargetID`、`Player.FledFrom`(NPC単位でのFLEE一度きり管理)、`Player.Quests`を追加
- 全戦闘NPCに`myth_requirement_item`/`myth_requirement_quest`/`flee_accurate`/`flee_succeeds_once`/`unwinnable`/`crew_loss_on_attack`を設定(該当するもののみ)

## 8. 多言語対応(日本語版)
最初に`LANG ja`を送ってからCONNECTすると、以降そのコネクションのLOOK/TALK/QUESTのテキストが日本語で返る。デフォルトは英語。

### 8.1 プロトコル面の設計判断
- RFCの`CONNECT <name>`はそのまま変更しない(他チームサーバー/クライアントとの相互接続を壊さないため、引数を増やさない)。
- 代わりに、認証前(CONNECT前)にだけ送れる独自コマンド`LANG <code>`を追加。対応コードは`en`/`ja`。CONNECT後に送るとERR 400。未知のコードもERR 400。
- RFC規定のJSON構造(LOOK/ATTACK/STATUS/QUEST等)のキー・形は変更しない。`Room.Name`等の`LocalizedText`(`map[string]string`、言語コード→テキスト)はサーバー内部のデータモデルのみで使い、実際にワイヤーに乗せる直前(`roomView`など)でリクエストした接続の言語に解決してから通常の`string`として送る。他チームのクライアントが言語システムを知らなくても、常に見慣れた形式のJSONが届く。
- 日本語訳が存在しないフィールドは自動的に英語にフォールバックする(`LocalizedText.Get(locale)`)。

### 8.2 実装ファイル
- `locale.go`: `LocalizedText`型、`LANG`コマンド、`roomView`(LOOK応答用のロケール解決済みDTO)
- `world.go`/`room.go`: `Room.Name`/`Description`、`Item.Name`/`Description`、`NPC.Name`/`Description`/`Dialogue`、`Quest.Name`/`Description`を`LocalizedText`化
- TAKE/DROP/TALK/ATTACK/QUESTでのNPC・アイテム名によるマッチングも、接続の言語での表記に対して行う(例: 日本語ロケールなら`TALK ポリュペモス`が通る)
- `data/world.json`: 3部作全ての`name`/`description`/`dialogue`に`"ja"`キーを追加済み

### 8.3 スコープ外にしたもの
- ATTACK/FLEE/MOVEハザードの戦闘フレーバーテキスト(`EVT ROOM COMBAT ...`のブロードキャスト)は英語固定。行為者本人の言語で組み立てているが、同じ部屋にいる他言語プレイヤーへの翻訳配信はしていない(複数受信者に別々の文言を送る仕組みが必要になるため、今回は見送り)。
