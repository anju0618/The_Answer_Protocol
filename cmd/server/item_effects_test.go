package main

import "testing"

func effectTestServer(items map[string]*Item) *Server {
	server := newServer("")
	server.world = &World{StartRoomID: "loc.start", Rooms: map[string]*Room{"loc.start": {ID: "loc.start"}}, Items: items}
	return server
}

func TestItemEffectsOnlyCountWhileCarried(t *testing.T) {
	server := effectTestServer(map[string]*Item{
		"item.good": {Effects: []ItemEffect{{blessingDamageBonus, 3}, {effectMaxHP, 10}}},
		"item.bad":  {Effects: []ItemEffect{{blessingDamageBonus, -2}, {effectMaxHP, -5}}},
	})
	player := &Player{}
	if got := server.maxHPLocked(player); got != maxPlayerHP {
		t.Fatalf("max HP with nothing carried = %d, want %d", got, maxPlayerHP)
	}
	player.Inventory = []string{"item.good", "item.bad"}
	if got := server.effectTotalLocked(player, blessingDamageBonus); got != 1 {
		t.Errorf("damage total = %d, want 1 (3 - 2)", got)
	}
	if got := server.maxHPLocked(player); got != maxPlayerHP+5 {
		t.Errorf("max HP = %d, want %d", got, maxPlayerHP+5)
	}
}

func TestMaxHPNeverDropsBelowFloorAndKeepsQuestGains(t *testing.T) {
	server := effectTestServer(map[string]*Item{"item.curse": {Effects: []ItemEffect{{effectMaxHP, -500}}}})
	player := &Player{MaxHPBonus: 12, Inventory: []string{"item.curse"}}
	if got := server.maxHPLocked(player); got != minPlayerMaxHP {
		t.Errorf("max HP = %d, want the floor %d", got, minPlayerMaxHP)
	}
	player.Inventory = nil
	if got := server.maxHPLocked(player); got != maxPlayerHP+12 {
		t.Errorf("max HP = %d, want %d", got, maxPlayerHP+12)
	}
}

func TestQuestMaxHPSurvivesDeath(t *testing.T) {
	server := effectTestServer(nil)
	server.players["alice"] = &Player{Name: "alice", HP: 50, MaxHPBonus: 8, RoomID: "loc.start"}
	server.respawnPlayerLocked(server.players["alice"], "alice", "attack_counter", "npc.x", "X")
	if got := server.players["alice"].MaxHPBonus; got != 8 {
		t.Fatalf("max HP bonus after death = %d, want 8", got)
	}
}

func TestQuestMaxHPGain(t *testing.T) {
	for hp, want := range map[int]int{10: 2, 15: 3, 25: 5, 3: 1} {
		if got := questMaxHPGain(&Quest{Reward: QuestReward{HP: hp}}); got != want {
			t.Errorf("gain for %d = %d, want %d", hp, got, want)
		}
	}
}

func TestNegativeCounterReductionHurtsMoreButIsCapped(t *testing.T) {
	if got := reduceCounter(10, -20); got != 12 {
		t.Errorf("reduceCounter(10, -20) = %d, want 12", got)
	}
	if got := reduceCounter(10, -400); got != 15 {
		t.Errorf("reduceCounter(10, -400) = %d, want 15 (capped at +50%%)", got)
	}
}

func TestValidateItemEffects(t *testing.T) {
	for name, effect := range map[string]ItemEffect{"unknown": {"luck", 1}, "zero": {effectMaxHP, 0}} {
		w := &World{Items: map[string]*Item{"item.x": {Effects: []ItemEffect{effect}}}}
		if w.validateItemEffects() == nil {
			t.Errorf("%s effect was accepted", name)
		}
	}
}
