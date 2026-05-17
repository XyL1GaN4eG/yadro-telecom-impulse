package handler

import (
	"io"
	"log"
	"testing"

	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/player"
)

func TestReadmeSuccessWorkflow(t *testing.T) {
	resetState(t)

	steps := []event.Command{
		{PlayerID: 1, EventID: 1},
		{PlayerID: 1, EventID: 2},
		{PlayerID: 1, EventID: 3},
		{PlayerID: 1, EventID: 11, Arg: "50"},
		{PlayerID: 1, EventID: 3},
		{PlayerID: 1, EventID: 4},
		{PlayerID: 1, EventID: 6},
		{PlayerID: 1, EventID: 11, Arg: "25"},
		{PlayerID: 1, EventID: 10, Arg: "80"},
		{PlayerID: 1, EventID: 11, Arg: "65"},
		{PlayerID: 1, EventID: 7},
		{PlayerID: 1, EventID: 8},
	}

	for _, step := range steps {
		if _, err := HandleCommand(step); err != nil {
			t.Fatalf("HandleCommand(%+v) returned error: %v", step, err)
		}
	}

	p := player.Players[1]
	if p.Status != player.StatusSuccess {
		t.Fatalf("player status = %v, want %v", p.Status, player.StatusSuccess)
	}
	if p.Health != 35 {
		t.Fatalf("player health = %d, want 35", p.Health)
	}
	if p.Floor != 1 {
		t.Fatalf("player floor = %d, want 1", p.Floor)
	}
	if !p.Dungeon.BossFloorEntered {
		t.Fatal("boss floor was not entered")
	}
	if !p.Dungeon.Floors[0].Cleared {
		t.Fatal("first floor was not cleared")
	}
	if !p.Dungeon.Floors[1].Cleared {
		t.Fatal("boss floor was not cleared")
	}
}

func TestReadmeFailWorkflow(t *testing.T) {
	resetState(t)

	steps := []event.Command{
		{PlayerID: 2, EventID: 1},
		{PlayerID: 2, EventID: 2},
		{PlayerID: 2, EventID: 5},
		{PlayerID: 2, EventID: 3},
		{PlayerID: 2, EventID: 11, Arg: "60"},
		{PlayerID: 2, EventID: 11, Arg: "50"},
	}

	for _, step := range steps {
		_, _ = HandleCommand(step)
	}

	p := player.Players[2]
	if p.Status != player.StatusFail {
		t.Fatalf("player status = %v, want %v", p.Status, player.StatusFail)
	}
	if p.Health != 0 {
		t.Fatalf("player health = %d, want 0", p.Health)
	}
	if p.Floor != 0 {
		t.Fatalf("player floor = %d, want 0", p.Floor)
	}
	if p.Dungeon.Floors[0].MonstersLeft != 1 {
		t.Fatalf("monsters left on first floor = %d, want 1", p.Dungeon.Floors[0].MonstersLeft)
	}
	if p.Dungeon.Floors[0].Cleared {
		t.Fatal("first floor is cleared, want uncleared")
	}
}

func TestReadmeDisqualifiesUnregisteredPlayer(t *testing.T) {
	resetState(t)

	if _, err := HandleCommand(event.Command{PlayerID: 3, EventID: 2}); err != nil {
		t.Fatalf("unregistered enter returned error: %v", err)
	}

	p := player.Players[3]
	if p.Status != player.StatusDisqual {
		t.Fatalf("player status = %v, want %v", p.Status, player.StatusDisqual)
	}
	if p.Health != player.DefaultHealth {
		t.Fatalf("player health = %d, want %d", p.Health, player.DefaultHealth)
	}
}

func resetState(t *testing.T) {
	t.Helper()

	log.SetOutput(io.Discard)
	player.Players = make(map[uint8]player.Player)
	game.Cfg = game.Config{
		Floors:   2,
		Monsters: 2,
		OpenAt:   "14:05:00",
		Duration: 2,
	}
}
