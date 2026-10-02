package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func arcOf(roomID string) string {
	switch {
	case strings.HasPrefix(roomID, "loc.argo_"):
		return "argo"
	case strings.HasPrefix(roomID, "loc.troy_"):
		return "troy"
	case strings.HasPrefix(roomID, "loc.ody_"):
		return "odyssey"
	}
	return "hub"
}

func TestThreeArcsAreIndependent(t *testing.T) {
	world := loadRealWorld(t)
	itemArc := func(id string) string { return arcOf(world.Items[id].HomeRoomID) }
	questArc := func(id string) string { return arcOf(world.NPCs[world.Quests[id].GiverNPCID].RoomID) }

	for id, npc := range world.NPCs {
		arc := arcOf(npc.RoomID)
		if npc.MythRequirementItem != "" && itemArc(npc.MythRequirementItem) != arc {
			t.Errorf("NPC %s (%s) needs %s from %s", id, arc, npc.MythRequirementItem, itemArc(npc.MythRequirementItem))
		}
		if npc.MythRequirementQuest != "" && questArc(npc.MythRequirementQuest) != arc {
			t.Errorf("NPC %s (%s) needs quest %s from %s", id, arc, npc.MythRequirementQuest, questArc(npc.MythRequirementQuest))
		}
		if e := npc.Ending; e != nil && len(e.RequiresEndings) == 0 {
			for _, itemID := range e.RequiresItems {
				if itemArc(itemID) != arc {
					t.Errorf("ending %s (%s) needs item %s from %s", e.ID, arc, itemID, itemArc(itemID))
				}
			}
			for _, questID := range e.RequiresQuests {
				if questArc(questID) != arc {
					t.Errorf("ending %s (%s) needs quest %s from %s", e.ID, arc, questID, questArc(questID))
				}
			}
		}
	}
	for id, room := range world.Rooms {
		if h := room.Hazard; h != nil && h.RequiredItemID != "" && itemArc(h.RequiredItemID) != arcOf(id) {
			t.Errorf("room %s needs item %s from another arc", id, h.RequiredItemID)
		}
	}
	for id, quest := range world.Quests {
		arc := questArc(id)
		var targetArc string
		if quest.Objective.Type == "collect_item" {
			targetArc = itemArc(quest.Objective.TargetID)
		} else {
			targetArc = arcOf(world.NPCs[quest.Objective.TargetID].RoomID)
		}
		if targetArc != arc {
			t.Errorf("quest %s is given in %s but its target is in %s", id, arc, targetArc)
		}
	}
}

func useBestCaseDice(t *testing.T) {
	t.Helper()
	original := randDamage
	randDamage = func(min, max int) int {
		if min == combatMinDamage {
			return max
		}
		return min
	}
	t.Cleanup(func() { randDamage = original })
}

func (client *testClient) run(t *testing.T, command string) []string {
	t.Helper()
	if _, err := fmt.Fprintln(client.conn, command); err != nil {
		t.Fatalf("send %q: %v", command, err)
	}
	return client.readLinesUntilResponse(t)
}

type walker struct {
	t      *testing.T
	server *Server
	client *testClient
	name   string
	events []string
}

func (w *walker) setHP(hp int) {
	w.server.mu.Lock()
	w.server.players[w.name].HP = hp
	w.server.mu.Unlock()
}

func (w *walker) teleport(roomID string) {
	w.server.mu.Lock()
	w.server.players[w.name].RoomID = roomID
	w.server.mu.Unlock()
}

func (w *walker) room() string {
	w.server.mu.Lock()
	defer w.server.mu.Unlock()
	return w.server.players[w.name].RoomID
}

func (w *walker) do(step string) {
	w.t.Helper()
	if roomID, ok := strings.CutPrefix(step, "teleport "); ok {
		w.teleport(roomID)
		return
	}
	if target, ok := strings.CutPrefix(step, "kill "); ok {
		for i := 0; i < 60; i++ {
			w.setHP(maxPlayerHP)
			lines := w.client.run(w.t, "ATTACK "+target)
			w.events = append(w.events, lines...)
			resp := lines[len(lines)-1]
			var result combatResult
			if err := json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &result); err != nil {
				w.t.Fatalf("step %q: response %q: %v", step, resp, err)
			}
			switch result.Status {
			case "victory":
				return
			case "dead":
				w.t.Fatalf("step %q: %s died: %s", step, w.name, resp)
			}
		}
		w.t.Fatalf("step %q: never won", step)
	}
	lines := w.client.run(w.t, step)
	w.events = append(w.events, lines...)
	if resp := lines[len(lines)-1]; !strings.HasPrefix(resp, "OK") {
		w.t.Fatalf("step %q: %s", step, resp)
	}
	if strings.HasPrefix(step, "TALK") {

		w.events = append(w.events, w.client.drain(w.t)...)
		return
	}
	if w.room() == w.server.world.StartRoomID {
		w.t.Fatalf("step %q: %s was sent back to the start room", step, w.name)
	}
}

func (client *testClient) drain(t *testing.T) []string {
	t.Helper()
	var lines []string
	for {
		client.conn.SetReadDeadline(time.Now().Add(120 * time.Millisecond))
		line, err := client.reader.ReadString(0x0a)
		if err != nil {
			client.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			return lines
		}
		lines = append(lines, strings.TrimSuffix(line, "\n"))
	}
}

func (w *walker) eventsContain(text string) bool {
	for _, line := range w.events {
		if strings.Contains(line, text) {
			return true
		}
	}
	return false
}

func (w *walker) has(itemID string) bool {
	w.server.mu.Lock()
	defer w.server.mu.Unlock()
	return w.server.players[w.name].hasItem(itemID)
}

func (w *walker) reached(endingID string) bool {
	w.server.mu.Lock()
	defer w.server.mu.Unlock()
	return w.server.players[w.name].Endings[endingID]
}

var argoRoute = []string{
	"teleport loc.argo_iolcus", "QUEST npc.tiphys", "MOVE south", "kill npc.amycus",
	"MOVE east", "QUEST npc.phineus", "kill npc.harpy",
	"MOVE east", "MOVE east", "MOVE north", "QUEST npc.aeetes", "TAKE item.medeas_ointment",
	"MOVE east", "kill npc.khalkotauroi", "kill npc.earthborn",
	"MOVE north", "QUEST npc.medea", "TAKE item.medeas_draught", "kill npc.colchis_dragon", "TAKE item.golden_fleece",
	"MOVE east", "kill npc.apsyrtus", "MOVE north", "MOVE east", "kill npc.talos",
	"MOVE north", "TALK npc.pelias",
}

var troyRoute = []string{
	"teleport loc.troy_ida", "MOVE east", "MOVE east", "QUEST npc.calchas", "TAKE item.aulis_offering",
	"MOVE east", "QUEST npc.agamemnon", "MOVE north", "QUEST npc.achilles", "TAKE item.shield_of_achilles",
	"MOVE south", "MOVE east", "MOVE east", "kill npc.hector",
	"MOVE north", "QUEST npc.priam", "TAKE item.hectors_ransom",
	"MOVE east", "QUEST npc.odysseus", "TAKE item.trojan_horse",
	"MOVE north", "TALK npc.aeneas",
}

var odysseyRoute = []string{
	"MOVE east", "MOVE east", "MOVE east", "MOVE east",
	"QUEST npc.trapped_sailor", "TAKE item.olive_stake", "kill npc.polyphemus",
	"MOVE east", "MOVE east", "FLEE", "MOVE east", "TAKE item.beeswax",
	"MOVE south", "MOVE east", "FLEE", "MOVE east", "MOVE east", "MOVE east", "MOVE east",
	"QUEST npc.eumaeus", "MOVE north", "QUEST npc.penelope", "TAKE item.odysseus_bow", "kill npc.antinous",
	"TALK npc.penelope",
}

func (w *walker) play(route []string) {
	w.t.Helper()
	w.teleport(w.server.world.StartRoomID)
	for _, step := range route {
		w.do(step)
	}
}

func newWalker(t *testing.T, server *Server, name string) *walker {
	t.Helper()
	client := startTestClient(t, server)
	client.connect(t, name)
	return &walker{t: t, server: server, client: client, name: name}
}

func TestEveryArcIsWinnableByEveryPlayer(t *testing.T) {
	useBestCaseDice(t)
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)

	arcs := []struct {
		name    string
		route   []string
		ending  string
		trophy  string
		keyItem string
	}{
		{"argo", argoRoute, "ending.argo", "item.laurel_of_the_argo", "item.golden_fleece"},
		{"troy", troyRoute, "ending.troy", "item.ember_of_troy", "item.trojan_horse"},
		{"odyssey", odysseyRoute, "ending.odyssey", "item.olive_branch_of_ithaca", "item.odysseus_bow"},
	}
	for _, playerName := range []string{"alice", "bob"} {
		w := newWalker(t, server, playerName)
		for _, arc := range arcs {
			w.play(arc.route)
			if !w.reached(arc.ending) {
				t.Fatalf("%s did not reach %s (events: %q)", playerName, arc.ending, w.events)
			}
			if !w.has(arc.trophy) || !w.has(arc.keyItem) {
				t.Errorf("%s: trophy %s / key item %s missing after the ending", playerName, arc.trophy, arc.keyItem)
			}
			if !w.eventsContain("EVT PLAYER ENDING === ENDING:") {
				t.Errorf("%s: no ending banner for %s", playerName, arc.name)
			}
		}

		w.teleport(server.world.StartRoomID)
		w.do("TALK npc.moirai")
		if !w.reached("ending.final") || !w.has("item.thread_of_fate") || !w.eventsContain("Atropos") {
			t.Fatalf("%s did not get the final ending (events: %q)", playerName, w.events)
		}
	}
}

func TestEndingNeedsItsRequirementsAndSaysWhatIsMissing(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	w := newWalker(t, server, "alice")
	w.teleport("loc.argo_iolcus")

	lines := append(w.client.run(t, "TALK npc.pelias"), w.client.drain(t)...)
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "EVT PLAYER ENDING ") || strings.Contains(joined, "You still need") {
		t.Fatalf("expected only a refusal with no list of requirements, got %q", lines)
	}
	if w.reached("ending.argo") {
		t.Fatal("ending reached without its requirements")
	}
}

func TestRenewableKeyItemsAndUniqueItems(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := newWalker(t, server, "alice")
	bob := newWalker(t, server, "bob")

	alice.teleport("loc.ody_circe")
	bob.teleport("loc.ody_circe")
	alice.do("TAKE item.beeswax")
	bob.do("TAKE item.beeswax")
	if !alice.has("item.beeswax") || !bob.has("item.beeswax") {
		t.Fatal("both players should hold their own beeswax")
	}
	look := alice.client.run(t, "LOOK")
	if strings.Contains(look[len(look)-1], "item.beeswax") {
		t.Fatalf("alice already holds beeswax, LOOK should not offer another: %s", look[len(look)-1])
	}
	if lines := alice.client.run(t, "TAKE item.beeswax"); !strings.HasPrefix(lines[len(lines)-1], "ERR 404") {
		t.Fatalf("second TAKE of a held key item = %v, want ERR 404", lines)
	}
	alice.do("DROP item.beeswax")
	alice.do("TAKE item.beeswax")

	alice.teleport("loc.argo_lemnos")
	bob.teleport("loc.argo_lemnos")
	alice.do("TAKE item.hospitality_gift")
	if lines := bob.client.run(t, "TAKE item.hospitality_gift"); !strings.HasPrefix(lines[len(lines)-1], "ERR 404") {
		t.Fatalf("unique item taken twice: %v", lines)
	}
	alice.do("DROP item.hospitality_gift")
	bob.do("TAKE item.hospitality_gift")
}

func TestDeathLosesEverythingExceptTrophies(t *testing.T) {
	useBestCaseDice(t)
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := newWalker(t, server, "alice")

	alice.teleport("loc.argo_lemnos")
	alice.do("TAKE item.hospitality_gift")
	alice.teleport("loc.ody_circe")
	alice.do("TAKE item.beeswax")
	server.mu.Lock()
	server.players["alice"].Inventory = append(server.players["alice"].Inventory, "item.thread_of_fate")
	server.players["alice"].setEnemyHP("npc.harpy", 5)
	server.mu.Unlock()

	alice.teleport("loc.ody_cyclops")
	lines := alice.client.run(t, "ATTACK npc.polyphemus")
	death := strings.Join(lines, "\n")
	if !strings.Contains(death, "You lost everything you were carrying except your trophies") {
		t.Fatalf("death message does not explain the loss: %s", death)
	}
	if alice.has("item.hospitality_gift") || alice.has("item.beeswax") {
		t.Fatal("ordinary belongings should be lost on death")
	}
	if !alice.has("item.thread_of_fate") {
		t.Fatal("trophies must survive death")
	}
	server.mu.Lock()
	returned := server.world.Items["item.hospitality_gift"].RoomID
	woundedHP, wounded := server.players["alice"].EnemyHP["npc.harpy"]
	server.mu.Unlock()
	if returned != "loc.argo_lemnos" {
		t.Errorf("lost unique item returned to %q, want its home room", returned)
	}
	if wounded {
		t.Errorf("enemies should recover after a death, harpy HP record = %d", woundedHP)
	}

	bob := newWalker(t, server, "bob")
	for _, cmd := range []struct {
		client *testClient
		line   string
	}{{alice.client, "GROUP CREATE"}, {alice.client, "GROUP INVITE bob"}, {bob.client, "GROUP JOIN alice"}} {
		if lines := cmd.client.run(t, cmd.line); !strings.HasPrefix(lines[len(lines)-1], "OK") {
			t.Fatalf("%s: %v", cmd.line, lines)
		}
	}
	alice.teleport("loc.ody_circe")
	bob.teleport("loc.ody_cyclops")
	alice.do("TAKE item.beeswax")
	alice.teleport("loc.ody_cyclops")
	lines = alice.client.run(t, "ATTACK npc.polyphemus")
	if !alice.has("item.beeswax") || !strings.Contains(strings.Join(lines, "\n"), "companions dragged your belongings to safety") {
		t.Fatalf("an ally in the room should protect belongings: %v", lines)
	}
}

func TestGroupAlliesShareCombatVictory(t *testing.T) {
	useBestCaseDice(t)
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := newWalker(t, server, "alice")
	bob := newWalker(t, server, "bob")
	for _, cmd := range []struct {
		client *testClient
		line   string
	}{{alice.client, "GROUP CREATE"}, {alice.client, "GROUP INVITE bob"}, {bob.client, "GROUP JOIN alice"}} {
		if lines := cmd.client.run(t, cmd.line); !strings.HasPrefix(lines[len(lines)-1], "OK") {
			t.Fatalf("%s: %v", cmd.line, lines)
		}
	}
	alice.teleport("loc.argo_salmydessus")
	bob.teleport("loc.argo_salmydessus")
	bob.do("QUEST npc.phineus")

	lines := alice.client.run(t, "ATTACK npc.harpy")
	var result combatResult
	if err := json.Unmarshal([]byte(strings.TrimPrefix(lines[len(lines)-1], "OK ")), &result); err != nil {
		t.Fatal(err)
	}
	if want := combatMaxDamage + allyDamageBonus; result.Damage != want {
		t.Fatalf("damage with one ally = %d, want %d", result.Damage, want)
	}
	alice.setHP(maxPlayerHP)
	alice.do("kill npc.harpy")

	bob.client.run(t, "STATUS")
	server.mu.Lock()
	bobHarpy := server.players["bob"].enemyHP("npc.harpy", server.world.NPCs["npc.harpy"])
	questState := server.players["bob"].Quests["quest.phineus_harpies"]
	server.mu.Unlock()
	if bobHarpy != 0 || questState == nil || questState.Status != "completed" {
		t.Fatalf("ally should share the victory: harpy hp=%d quest=%+v", bobHarpy, questState)
	}
}

func TestEnemiesAreDefeatedPerPlayer(t *testing.T) {
	useBestCaseDice(t)
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := newWalker(t, server, "alice")
	bob := newWalker(t, server, "bob")
	alice.teleport("loc.argo_salmydessus")
	bob.teleport("loc.argo_salmydessus")

	alice.do("kill npc.harpy")
	bob.do("QUEST npc.phineus")
	bob.do("kill npc.harpy")
	server.mu.Lock()
	state := server.players["bob"].Quests["quest.phineus_harpies"]
	server.mu.Unlock()
	if state == nil || state.Status != "completed" {
		t.Fatalf("bob should complete the harpy quest after his own fight: %+v", state)
	}

	alice.do("QUEST npc.phineus")
	server.mu.Lock()
	state = server.players["alice"].Quests["quest.phineus_harpies"]
	server.mu.Unlock()
	if state == nil || state.Status != "completed" {
		t.Fatalf("a quest accepted after the kill should count it: %+v", state)
	}
}

func TestCombatFlavorIsLocalizedPerRecipient(t *testing.T) {
	useBestCaseDice(t)
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	english := startTestClient(t, server)
	english.connect(t, "alice")
	japanese := startTestClient(t, server)
	japanese.command(t, "LANG ja", "OK lang=ja")
	japanese.command(t, "CONNECT bob", "OK connected")
	japanese.name = "bob"
	english.drain(t)
	japanese.drain(t)
	for _, name := range []string{"alice", "bob"} {
		server.mu.Lock()
		server.players[name].RoomID = "loc.argo_salmydessus"
		server.mu.Unlock()
	}

	if _, err := fmt.Fprintln(english.conn, "ATTACK npc.harpy"); err != nil {
		t.Fatal(err)
	}
	en := english.waitEvent(t, "EVT ROOM COMBAT ")
	ja := japanese.waitEvent(t, "EVT ROOM COMBAT ")
	if !strings.Contains(en, "attacks Harpy for") {
		t.Errorf("English recipient got %q", en)
	}
	if !strings.Contains(ja, "ハルピュイア") || !strings.Contains(ja, "ダメージ") {
		t.Errorf("Japanese recipient got %q", ja)
	}
}

func TestRegenRestoresHPOverTime(t *testing.T) {
	start := time.Now()
	p := &Player{HP: 20}
	p.regenLocked(start, 0, maxPlayerHP)
	p.regenLocked(start.Add(10*time.Second), 0, maxPlayerHP)
	if want := 20 + int(10*time.Second/regenInterval)*regenAmount; p.HP != want {
		t.Fatalf("HP after 10s = %d, want %d", p.HP, want)
	}
	p.regenLocked(start.Add(time.Hour), 0, maxPlayerHP)
	if p.HP != maxPlayerHP {
		t.Fatalf("HP after an hour = %d, want %d", p.HP, maxPlayerHP)
	}
}

func TestLaestrygoniansCanBeFledWithoutFighting(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	w := newWalker(t, server, "alice")
	w.teleport("loc.ody_laestrygonians")
	w.do("FLEE")
	w.do("MOVE east")
	if w.room() != "loc.ody_circe" {
		t.Fatalf("after fleeing the giants alice is in %q, want loc.ody_circe", w.room())
	}
}

func TestCrewResetsOnEachFreshVoyage(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	w := newWalker(t, server, "alice")
	w.do("MOVE east")
	server.mu.Lock()
	server.players["alice"].Crew = 1
	server.mu.Unlock()
	w.teleport(server.world.StartRoomID)
	w.do("MOVE east")
	server.mu.Lock()
	crew := server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew {
		t.Fatalf("crew after re-entering from the hall = %d, want %d", crew, startingCrew)
	}

	server.mu.Lock()
	server.players["alice"].Crew = 3
	server.mu.Unlock()
	w.do("MOVE east")
	w.do("MOVE west")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 1 {
		t.Fatalf("crew after stepping east and back = %d, want 1 (3 minus Cicones toll 2; no refill)", crew)
	}
}
