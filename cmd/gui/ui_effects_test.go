package main

import (
	"testing"
	"time"
)

func TestFlashSceneStartsAnimationAndToleratesMissingLayer(t *testing.T) {
	(&gui{}).flashScene(flashHurt, time.Millisecond)

	ui := newCombatTestUI(t)
	ui.flashScene(flashHurt, 50*time.Millisecond)
	if ui.flashAnim == nil {
		t.Fatal("flashScene must start an animation")
	}
	ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":90,"target_hp":28,"damage":12,"status":"combat"}`)
	ui.handleResponse("MOVE", "MOVE east", "OK room=loc.ody_troy_shore")
}
