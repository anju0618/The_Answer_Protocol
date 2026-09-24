package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestJapaneseTranslationsAgainstRealWorldData(t *testing.T) {
	world, err := loadWorld(filepath.Join("..", "..", "data", "world.json"))
	if err != nil {
		t.Fatalf("load real world: %v", err)
	}

	server := newServer(t.TempDir())
	server.world = world
	alice := startTestClient(t, server)
	alice.cmd(t, "LANG ja", "OK lang=ja")
	alice.connect(t, "alice")

	look := alice.cmdJSON(t, "LOOK")
	room := look["room"].(map[string]any)
	if room["name"] != "運命の間" {
		t.Fatalf("hall of fates ja name = %v, want 運命の間", room["name"])
	}

	teleport := func(roomID string) {
		server.mu.Lock()
		server.players["alice"].RoomID = roomID
		server.mu.Unlock()
	}

	teleport("loc.argo_bull_field")
	look = alice.cmdJSON(t, "LOOK")
	if look["room"].(map[string]any)["name"] != "青銅の雄牛の野" {
		t.Fatalf("bull field ja name = %v", look["room"])
	}
	alice.cmd(t, "TALK 青銅の雄牛", "OK 雄牛たちは火を吹き、地を蹴りつけ、誰であろうと軛をつけてみよと挑発している。")

	teleport("loc.troy_gate")
	alice.cmd(t, "TALK ヘクトール", "OK この門を守り抜く一日一日が、我が都市がまだ立っていることの証だ。")

	teleport("loc.ody_cyclops")
	alice.cmd(t, "TALK ポリュペモス", "OK また客が、自分から俺の食料庫に迷い込んできたか。じっとしてろ、すぐ終わる。")

	teleport("loc.ody_cyclops")
	quest := alice.cmdJSON(t, "QUEST 囚われた水夫")
	// 説明文の本文に続いて、どこで何をすればよいかのヒントが付く。
	description, _ := quest["description"].(string)
	if !strings.HasPrefix(description, "ポリュペモスの目を潰し、洞窟に囚われた乗組員を救い出そう。") || !strings.Contains(description, "ヒント: ") {
		t.Fatalf("quest ja description = %v", quest)
	}

	teleport("loc.ody_cyclops")
	alice.cmd(t, "TAKE 研がれたオリーブの杭", "OK taken=item.olive_stake")
	alice.cmd(t, "DROP 研がれたオリーブの杭", "OK dropped=item.olive_stake")
}
