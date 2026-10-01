# ゲームオーバー部屋の絵: Codex 用の指示文

史実に反する選択をして死んだときに表示する絵です(GUI のゲームオーバー画面)。

## 置き方(これを守ると、そのまま表示される)

- 形式: **PNG、960 × 576 ピクセル**ちょうど(既存の部屋の絵と同じ。違うとテストが落ちます)
- 置き場所: `cmd/gui/assets/rooms/` に、下の表のファイル名で置く
- 7枚はすべて別の絵にする(同じ画像を使い回すとテストが落ちます)
- 文字・ロゴ・枠・UI は絵に入れない(文字はGUIが重ねて表示します)

| ファイル名 | 場面 |
|---|---|
| `loc.ody_ismarus_feast.png` | イスマロスの宴 |
| `loc.ody_lotus_garden.png` | 蓮の園 |
| `loc.ody_sealed_cave.png` | ふさがれた洞窟 |
| `loc.ody_laestrygonian_depths.png` | 港の奥 |
| `loc.ody_pigsty.png` | キルケーの豚小屋 |
| `loc.ody_eternal_ogygia.png` | 果てしなきオギュギア |
| `loc.ody_charybdis.png` | カリュブディスの大渦(すでにある穏やかな絵の描き直し。任意) |

## 共通の指示文(毎回、最初に付ける)

```
16-bit pixel art illustration, 960x576 pixels, wide landscape scene.
Same style as a retro Greek-mythology text adventure: chunky visible pixels,
limited painterly palette, dark navy night sky, ancient Greek ruins.
Eerie, unsettling, slightly gory atmosphere told through implication:
bloodstains, bones, broken wood, wrong shadows, sickly glowing light.
No text, no letters, no UI, no border, no watermark.
No close-up faces. Small distant figures only.
```

## 個別の指示文(共通の指示文のあとに続ける)

### loc.ody_ismarus_feast.png(イスマロスの宴)

```
Scene: a Mediterranean beach at dawn after a massacre. A half-eaten feast is
still laid out on the sand: overturned wine jars, spilled red wine mixed with
blood, butchered sheep. Dozens of motionless sailors lie among the tables.
At the treeline, a dark crowd of Cicones warriors with spears is just
arriving. Beached black ships behind, sails torn. Grey-red dawn light.
```

### loc.ody_lotus_garden.png(蓮の園)

```
Scene: a dreamy island garden that is slowly turning wrong. Huge pale lotus
flowers glow with a sickly pink-violet light. Sailors sit and lie among them
with blank smiles and glassy empty eyes, vines and lotus roots creeping over
their legs and into their mouths. A far-off black ship rots at the shore,
forgotten. Soft fog, drifting petals, a sweet but nauseating mood.
```

### loc.ody_sealed_cave.png(ふさがれた洞窟)

```
Scene: the inside of a giant's cave seen from deep within, in near darkness.
A colossal boulder seals the cave mouth; only thin cracks of cold light leak
around it. In the foreground a huge sleeping one-eyed giant lies curled up,
and around him gnawed bones, skulls and dried blood. A few tiny emaciated
sailors with swords sit against the walls, trapped. Claustrophobic, damp,
torch-less, deep blue-black with a faint red glow from a dying fire.
```

### loc.ody_laestrygonian_depths.png(港の奥)

```
Scene: a narrow deep harbor enclosed by sheer cliffs, seen from the water.
On the cliff tops, rows of enormous giants hurl house-sized boulders down.
The water is full of splintered ships, floating wreckage, drowned sailors and
dark red water. One boulder is just about to hit the viewer's ship at the
bottom of the frame. Storm-grey sky, dread, hopeless.
```

### loc.ody_pigsty.png(キルケーの豚小屋)

```
Scene: a stone pigsty beside a witch's marble hall at dusk. Inside, a crowd of
pigs with unsettlingly human eyes, scraps of torn sailors' clothing and
bronze armor in the mud, and a ring on one trotter. Muddy, filthy, flies
buzzing. In the background, a tall silhouette of a sorceress with a wand
stands on the hall's steps with a faint smile. Sickly green and amber light.
```

### loc.ody_eternal_ogygia.png(果てしなきオギュギア)

```
Scene: a beautiful but suffocating island cave-palace of a nymph, with a
calm sea and an endless sunset that never ends. Everything is too perfect and
too still. A lone man sits on the shore with his back to the viewer,
staring at the horizon; around him, countless identical empty wine cups and
a heap of unfinished, rotting raft planks, grey roots slowly growing over
him. A ghostly pale woman's silhouette stands at the cave mouth. Uncanny,
quiet horror: endless time, no escape.
```

### loc.ody_charybdis.png(任意の描き直し)

```
Scene: a gigantic whirlpool in the open sea at night, a black spiral throat
dropping into darkness. Shattered ship hulls, snapped masts, and pale bodies
are spiraling down into it. A faint red glow at the bottom. Jagged rocks on
the cliffs, a rotten fig tree overhead. Terrifying and hopeless.
```

## 絵を置いたあとにやること

1. ファイルを `cmd/gui/assets/rooms/` に置く
2. `go test ./cmd/gui/` を実行する(サイズ・重複・読み込みを確認します)
3. `cmd/gui/art_test.go` の「即死の部屋は絵がなくてもよい」という緩めた条件を、「全部の部屋に絵が必要」に戻してもらう

## アテナの立ち絵(仮絵の差し替え用)

`cmd/gui/assets/npcs/npc.athena.png` は今、プログラムで作った仮の絵です。本絵に差し替えるときは次を守ってください。

- 形式: **PNG(透過あり)、192 × 312 ピクセル**ちょうど。他のNPCの立ち絵と同じ形式。
- 内容: 灰色の瞳の女神アテナ。若い羊飼いに化けた姿で、兜・槍・盾を持つ。文字や枠は入れない。
