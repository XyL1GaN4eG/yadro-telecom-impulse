package format

import (
	"impulse/internal/event"
	"impulse/internal/player"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	cmd := event.Command{
		Time:     14*time.Hour + 29*time.Minute,
		PlayerID: 2,
		EventID:  11,
		Arg:      "50",
	}

	want := "[14:29:00] Player [2] recieved [50] of damage"
	if got := Command(cmd); got != want {
		t.Fatalf("Command() = %q, want %q", got, want)
	}
}

func TestFinalReport(t *testing.T) {
	players := map[uint8]player.Player{
		2: {
			ID:             2,
			Health:         0,
			Status:         player.StatusFail,
			EnteredDungeon: true,
			EnteredAt:      14*time.Hour + 10*time.Minute,
			Finished:       true,
			FinishedAt:     14*time.Hour + 29*time.Minute,
		},
		1: {
			ID:             1,
			Health:         35,
			Status:         player.StatusSuccess,
			EnteredDungeon: true,
			EnteredAt:      14*time.Hour + 40*time.Minute,
			Finished:       true,
			FinishedAt:     15*time.Hour + 4*time.Minute,
			Dungeon: player.DungeonRun{
				BossKillTime: 11 * time.Minute,
				Floors: []player.Floor{
					{Cleared: true, TimeSpent: 5 * time.Minute},
					{IsBoss: true, Cleared: true},
				},
			},
		},
	}

	want := "Final report:\n" +
		"[SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35\n" +
		"[FAIL] 2 [00:19:00, 00:00:00, 00:00:00] HP:0\n"
	if got := FinalReport(players); got != want {
		t.Fatalf("FinalReport() = %q, want %q", got, want)
	}
}
