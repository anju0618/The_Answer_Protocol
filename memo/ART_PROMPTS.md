# 絵の依頼書: Codex 用の指示文

次の2種類、計12枚を2026-10-02に生成・配置済みです。背景は960×576、敵の画像はすべて312×192の背景透明PNGです。再生成用に指示文を残しています。

1. 織られざる機の背景(1枚。仮絵を差し替え済み)
2. 倒された敵の立ち絵(11枚。専用画像を追加済み)

---

## 1. 織られざる機の背景(仮絵の差し替え用)

真エンディング「答え」のあと、運命の間の南から行ける部屋です。

### 置き方

- 形式: **PNG、960 × 576 ピクセル**ちょうど(他の部屋の絵と同じ。違うとテストが落ちます)
- 置き場所と名前: `cmd/gui/assets/rooms/loc.unwoven_loom.png`(いまの仮絵に**同名で上書き**)
- 他の部屋の絵と同じ画像にしない(同じだとテストが落ちます)
- 文字・ロゴ・枠・UI は絵に入れない

### 指示文

```
16-bit pixel art illustration, 960x576 pixels, wide landscape scene.
Same style as a retro Greek-mythology text adventure: chunky visible pixels,
limited painterly palette, dark navy night sky, ancient Greek ruins.
Calm, awe-filled, hopeful mood.
No text, no letters, no UI, no border, no watermark.
No close-up faces. Small distant figures only.

Scene: a quiet, sacred chamber hidden behind the three great tapestries of
the Hall of the Fates. In the center stands a huge wooden loom that nobody
has ever used: the warp threads are strung in glowing gold, but the weft
threads hang loose and unfinished, drifting as if waiting for a hand.
Only the bottom few rows are woven. Faint golden motes of light float where
the unwritten future would be. Two bronze braziers burn on either side.
Dark navy night sky with stars visible through an open colonnade behind.
```

---

## 2. 倒された敵の立ち絵

プレイヤーが敵を倒すと、その部屋では敵が「倒れた絵」に変わります。専用の絵があればそれを使い、なければゲームが、普通の立ち絵を暗くして横倒しにした絵を自動で作ります(動作確認済み)。そのため、1枚ずつ好きな順に置いて構いません。

### 置き方

- 形式: **PNG(背景は透明)、312 × 192 ピクセル**を目安にした**横長**の絵。サイズが違っても、場面に合わせて拡大縮小されます。
- 置き場所: `cmd/gui/assets/npcs_defeated/` に、下の表のファイル名で置く
- 向き: 床に倒れた姿を**真横から**見た絵。頭は**右**側にする(自動生成の絵も同じ向きです)
- 地面の影や背景は入れない(部屋の絵の上にそのまま重ねます)
- 文字・枠・UI は入れない
- 血はにじませる程度にとどめる(ゲーム全体が神話の語り口なので、静かに倒れた印象にする)
- 元になる立ち絵は `cmd/gui/assets/npcs/<同じファイル名>` です。**服・色・持ち物をそろえてください**

| ファイル名 | 敵 | 倒れた姿のイメージ |
|---|---|---|
| `npc.amycus.png` | アミュコス王 | 籠手をはめたまま仰向けに倒れた、大柄なボクサーの王 |
| `npc.harpy.png` | ハルピュイア | 翼を広げて地面に落ちた、半人半鳥の怪物 |
| `npc.khalkotauroi.png` | 青銅の雄牛 | 青銅の脚を投げ出して倒れた雄牛。鼻からの炎は消えている |
| `npc.earthborn.png` | 大地生まれの戦士 | 槍と盾を落として崩れた戦士。体が土に還りかけている |
| `npc.colchis_dragon.png` | 眠らぬ竜 | 長い体をだらりと伸ばして、ついに目を閉じた大蛇 |
| `npc.apsyrtus.png` | アプシュルトス | 剣を手放して倒れた若い王子 |
| `npc.talos.png` | タロス | 青銅の巨体が横倒しになり、足首の封じ目から光が漏れている |
| `npc.hector.png` | ヘクトール | 盾を傍らに、穏やかな顔で横たわる英雄(尊厳のある姿に) |
| `npc.polyphemus.png` | ポリュペモス | 大きな一つ目を閉じて倒れた巨人。傍らに焼けた杭 |
| `npc.scylla.png` | スキュラ | 六つの首がぐったりと垂れた怪物 |
| `npc.antinous.png` | アンティノオス | 酒杯を落として倒れた求婚者。宴の皿が散らばる |

### 指示文(1枚ごとに、最後の `Character:` の行だけ差し替えて使う)

```
16-bit pixel art character sprite, wide horizontal canvas about 312x192 pixels,
transparent background.
Same style as a retro Greek-mythology text adventure: chunky visible pixels,
limited painterly palette, dark navy outlines, ancient Greek.
The character is DEFEATED, lying on the ground seen from the side, head on the
right, completely still. Quiet and solemn, not gory: at most a small stain.
Keep the same clothes, colors and equipment as the standing version.
No ground, no shadow, no background, no text, no UI, no border, no watermark.

Character: <上の表の「倒れた姿のイメージ」を英語にしたもの>
```

使い方の例(アミュコス王):

```
Character: King Amycus, a huge boastful boxer-king in leather gloves, lying on
his back, arms limp at his sides, eyes closed.
```

### 絵を置いたあとにやること

1. ファイルを `cmd/gui/assets/npcs_defeated/` に置く
2. `go test ./cmd/gui/` を実行する
3. 敵を倒して、倒れた絵に変わることを GUI で確認する
