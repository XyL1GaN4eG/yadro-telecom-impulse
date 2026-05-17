package tui

import (
	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/player"
	"impulse/internal/replay"
	"testing"
	"time"
)

func TestModelStepResetAndStepBack(t *testing.T) {
	game.Cfg = game.Config{
		Floors:   2,
		Monsters: 2,
		OpenAt:   "14:05:00",
		Duration: 2,
	}
	replay.ResetPlayers()

	entries := []Entry{
		{Line: "[14:00:00] 1 1", Cmd: event.Command{Time: 14 * time.Hour, PlayerID: 1, EventID: 1}},
		{Line: "[14:01:00] 1 2", Cmd: event.Command{Time: 14*time.Hour + time.Minute, PlayerID: 1, EventID: 2}},
		{Line: "[14:02:00] 1 3", Cmd: event.Command{Time: 14*time.Hour + 2*time.Minute, PlayerID: 1, EventID: 3}},
	}
	model := NewModel(entries)

	if !model.Step() || model.Index != 1 {
		t.Fatalf("after first step index = %d, want 1", model.Index)
	}
	if _, ok := player.Players[1]; !ok {
		t.Fatal("player was not registered after first step")
	}

	if !model.Step() || model.Index != 2 {
		t.Fatalf("after second step index = %d, want 2", model.Index)
	}
	if !player.Players[1].EnteredDungeon {
		t.Fatal("player did not enter dungeon after second step")
	}

	model.StepBack()
	if model.Index != 1 {
		t.Fatalf("after step back index = %d, want 1", model.Index)
	}
	if player.Players[1].EnteredDungeon {
		t.Fatal("step back did not replay to previous state")
	}

	model.Reset()
	if model.Index != 0 {
		t.Fatalf("after reset index = %d, want 0", model.Index)
	}
	if len(player.Players) != 0 {
		t.Fatalf("after reset players = %d, want 0", len(player.Players))
	}
	if len(model.Log) != 0 {
		t.Fatalf("after reset log = %d lines, want 0", len(model.Log))
	}
}
