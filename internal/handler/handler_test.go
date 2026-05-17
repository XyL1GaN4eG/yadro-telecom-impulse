package handler

import (
	"errors"
	"testing"
	"time"

	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/player"
)

func TestReadmeSuccessWorkflow(t *testing.T) {
	resetState(t)

	steps := []event.Command{
		{Time: 14*time.Hour + 40*time.Minute, PlayerID: 1, EventID: 1},
		{Time: 14*time.Hour + 40*time.Minute, PlayerID: 1, EventID: 2},
		{Time: 14*time.Hour + 41*time.Minute, PlayerID: 1, EventID: 3},
		{Time: 14*time.Hour + 44*time.Minute, PlayerID: 1, EventID: 11, Arg: "50"},
		{Time: 14*time.Hour + 45*time.Minute, PlayerID: 1, EventID: 3},
		{Time: 14*time.Hour + 48*time.Minute, PlayerID: 1, EventID: 4},
		{Time: 14*time.Hour + 48*time.Minute, PlayerID: 1, EventID: 6},
		{Time: 14*time.Hour + 49*time.Minute, PlayerID: 1, EventID: 11, Arg: "25"},
		{Time: 14*time.Hour + 49*time.Minute + 2*time.Second, PlayerID: 1, EventID: 10, Arg: "80"},
		{Time: 14*time.Hour + 50*time.Minute, PlayerID: 1, EventID: 11, Arg: "65"},
		{Time: 14*time.Hour + 59*time.Minute, PlayerID: 1, EventID: 7},
		{Time: 15*time.Hour + 4*time.Minute, PlayerID: 1, EventID: 8},
	}

	for _, step := range steps {
		if _, err := HandleCommand(step); err != nil && !errors.Is(err, ErrNoOutput) {
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
	if p.TotalTime() != 24*time.Minute {
		t.Fatalf("total time = %v, want 24m", p.TotalTime())
	}
	if p.AverageFloorClearTime() != 5*time.Minute {
		t.Fatalf("average floor clear time = %v, want 5m", p.AverageFloorClearTime())
	}
	if p.Dungeon.BossKillTime != 11*time.Minute {
		t.Fatalf("boss kill time = %v, want 11m", p.Dungeon.BossKillTime)
	}
}

func TestReadmeFailWorkflow(t *testing.T) {
	resetState(t)

	steps := []event.Command{
		{Time: 14 * time.Hour, PlayerID: 2, EventID: 1},
		{Time: 14*time.Hour + 10*time.Minute, PlayerID: 2, EventID: 2},
		{Time: 14*time.Hour + 11*time.Minute, PlayerID: 2, EventID: 5},
		{Time: 14*time.Hour + 14*time.Minute, PlayerID: 2, EventID: 3},
		{Time: 14*time.Hour + 27*time.Minute, PlayerID: 2, EventID: 11, Arg: "60"},
		{Time: 14*time.Hour + 29*time.Minute, PlayerID: 2, EventID: 11, Arg: "50"},
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
	if p.TotalTime() != 19*time.Minute {
		t.Fatalf("total time = %v, want 19m", p.TotalTime())
	}
}

func TestReadmeDisqualifiesUnregisteredPlayer(t *testing.T) {
	resetState(t)

	if _, err := HandleCommand(event.Command{Time: 14*time.Hour + 10*time.Minute, PlayerID: 3, EventID: 2}); err != nil {
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

func TestUnregisteredActionDisqualifiesPlayer(t *testing.T) {
	resetState(t)

	if _, err := HandleCommand(event.Command{Time: time.Hour, PlayerID: 4, EventID: 3}); err != nil {
		t.Fatalf("unregistered action returned error: %v", err)
	}

	if player.Players[4].Status != player.StatusDisqual {
		t.Fatalf("player status = %v, want %v", player.Players[4].Status, player.StatusDisqual)
	}
}

func TestFinishedPlayerEventsAreIgnored(t *testing.T) {
	resetState(t)

	steps := []event.Command{
		{Time: time.Hour, PlayerID: 1, EventID: 1},
		{Time: time.Hour, PlayerID: 1, EventID: 2},
		{Time: time.Hour + time.Minute, PlayerID: 1, EventID: 8},
	}
	for _, step := range steps {
		if _, err := HandleCommand(step); err != nil {
			t.Fatalf("HandleCommand(%+v) returned error: %v", step, err)
		}
	}

	if _, err := HandleCommand(event.Command{Time: time.Hour + 2*time.Minute, PlayerID: 1, EventID: 4}); !errors.Is(err, ErrNoOutput) {
		t.Fatalf("post-finish event error = %v, want ErrNoOutput", err)
	}
}

func TestCloseActivePlayers(t *testing.T) {
	resetState(t)

	steps := []event.Command{
		{Time: time.Hour, PlayerID: 1, EventID: 1},
		{Time: time.Hour, PlayerID: 1, EventID: 2},
	}
	for _, step := range steps {
		if _, err := HandleCommand(step); err != nil {
			t.Fatalf("HandleCommand(%+v) returned error: %v", step, err)
		}
	}

	CloseActivePlayers(3 * time.Hour)

	p := player.Players[1]
	if !p.Finished {
		t.Fatal("player was not finished")
	}
	if p.TotalTime() != 2*time.Hour {
		t.Fatalf("total time = %v, want 2h", p.TotalTime())
	}
}

func resetState(t *testing.T) {
	t.Helper()

	player.Players = make(map[uint8]player.Player)
	game.Cfg = game.Config{
		Floors:   2,
		Monsters: 2,
		OpenAt:   "14:05:00",
		Duration: 2,
	}
}
