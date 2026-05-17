package replay

import (
	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/player"
	"testing"
	"time"
)

func TestProcessCommandSuccessfulEventReturnsIncomingOutput(t *testing.T) {
	resetState()

	lines := ProcessCommand(event.Command{
		Time:     14 * time.Hour,
		PlayerID: 1,
		EventID:  1,
	})

	want := "[14:00:00] Player [1] registered"
	if len(lines) != 1 || lines[0] != want {
		t.Fatalf("ProcessCommand() = %#v, want %#v", lines, []string{want})
	}
}

func TestProcessCommandImpossibleMoveReturnsOutgoing33(t *testing.T) {
	resetState()
	ProcessCommand(event.Command{PlayerID: 1, EventID: 1})
	ProcessCommand(event.Command{PlayerID: 1, EventID: 2})

	lines := ProcessCommand(event.Command{
		Time:     14*time.Hour + time.Minute,
		PlayerID: 1,
		EventID:  4,
	})

	want := "[14:01:00] Player [1] makes imposible move [4]"
	if len(lines) != 1 || lines[0] != want {
		t.Fatalf("ProcessCommand() = %#v, want %#v", lines, []string{want})
	}
}

func TestProcessCommandDisqualifiedUnregisteredPlayerReturnsOutgoing31(t *testing.T) {
	resetState()

	lines := ProcessCommand(event.Command{
		Time:     14*time.Hour + 2*time.Minute,
		PlayerID: 3,
		EventID:  2,
	})

	want := "[14:02:00] Player [3] is disqualified"
	if len(lines) != 1 || lines[0] != want {
		t.Fatalf("ProcessCommand() = %#v, want %#v", lines, []string{want})
	}
}

func TestProcessCommandDeathAfterDamageReturnsIncomingDamageAndOutgoing32(t *testing.T) {
	resetState()
	ProcessCommand(event.Command{PlayerID: 1, EventID: 1})
	ProcessCommand(event.Command{PlayerID: 1, EventID: 2})

	lines := ProcessCommand(event.Command{
		Time:     14*time.Hour + 3*time.Minute,
		PlayerID: 1,
		EventID:  11,
		Arg:      "100",
	})

	want := []string{
		"[14:03:00] Player [1] recieved [100] of damage",
		"[14:03:00] Player [1] is dead",
	}
	if len(lines) != len(want) {
		t.Fatalf("ProcessCommand() = %#v, want %#v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("ProcessCommand()[%d] = %q, want %q", i, lines[i], want[i])
		}
	}
}

func resetState() {
	ResetPlayers()
	game.Cfg = game.Config{
		Floors:   2,
		Monsters: 2,
		OpenAt:   "14:05:00",
		Duration: 2,
	}
	player.Players = make(map[uint8]player.Player)
}
