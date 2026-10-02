package main

import (
	"testing"
	"time"
)

func TestBlessingsFollowReachedEndings(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	player := &Player{Name: "alice", Endings: map[string]bool{}}

	for _, effect := range []string{blessingCounterReduction, blessingRegenBonus, blessingDamageBonus} {
		if got := server.blessingTotalLocked(player, effect); got != 0 {
			t.Fatalf("%s before any ending = %d, want 0", effect, got)
		}
	}
	player.Endings["ending.odyssey"] = true // Athena
	if got := server.blessingTotalLocked(player, blessingCounterReduction); got != 20 {
		t.Fatalf("Athena's counter reduction = %d, want 20", got)
	}
	player.Endings["ending.argo"] = true // Hera
	player.Endings["ending.troy"] = true // Apollo
	if server.blessingTotalLocked(player, blessingRegenBonus) != 1 || server.blessingTotalLocked(player, blessingDamageBonus) != 3 {
		t.Fatal("Hera and Apollo blessings are missing")
	}
}

func TestReduceCounterIsCappedAndNeverBelowOne(t *testing.T) {
	cases := []struct{ counter, percent, want int }{
		{10, 0, 10}, {10, 20, 8}, {10, 80, 2}, {10, 200, 2}, {1, 80, 1},
	}
	for _, c := range cases {
		if got := reduceCounter(c.counter, c.percent); got != c.want {
			t.Errorf("reduceCounter(%d, %d) = %d, want %d", c.counter, c.percent, got, c.want)
		}
	}
}

func TestRegenBonusSpeedsUpHealing(t *testing.T) {
	start := time.Now()
	plain := &Player{HP: 20}
	blessed := &Player{HP: 20}
	plain.regenLocked(start, 0, maxPlayerHP)
	blessed.regenLocked(start, 1, maxPlayerHP)
	plain.regenLocked(start.Add(10*time.Second), 0, maxPlayerHP)
	blessed.regenLocked(start.Add(10*time.Second), 1, maxPlayerHP)
	if plain.HP != 25 || blessed.HP != 30 {
		t.Fatalf("after 10s: plain=%d (want 25), blessed=%d (want 30)", plain.HP, blessed.HP)
	}
}

func TestAthenaStandsOnIthacasShoreAndBlessingsAreValidated(t *testing.T) {
	world := loadRealWorld(t)
	athena := world.NPCs["npc.athena"]
	if athena == nil || athena.RoomID != "loc.ody_ithaca_shore" || athena.Role != "dialogue" {
		t.Fatalf("Athena = %+v", athena)
	}
	world.NPCs["npc.penelope"].Ending.Blessing.Effect = "fly"
	if err := world.validate(); err == nil {
		t.Fatal("an unknown blessing effect must fail validation")
	}
}
