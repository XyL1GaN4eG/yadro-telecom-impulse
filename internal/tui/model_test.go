package tui

import (
	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/player"
	"impulse/internal/replay"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

func TestDungeonRoomUsesASCIIGraphics(t *testing.T) {
	p := player.New(7, 2, 3)
	p.Floor = 0
	p.Health = 42

	lines := dungeonRoom(NewModel(nil), p.Dungeon.Floors[0], 0, 24, false)
	rendered := strings.Join(lines, "\n")
	for _, want := range []string{"+----------------------+", "|  00 FLOOR", "M:MMM", "P:-"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("dungeonRoom() =\n%s\nwant substring %q", rendered, want)
		}
	}
}

func TestDungeonViewShowsFullMapAndPlayers(t *testing.T) {
	game.Cfg = game.Config{
		Floors:   2,
		Monsters: 2,
		OpenAt:   "14:05:00",
		Duration: 2,
	}
	replay.ResetPlayers()

	model := NewModel([]Entry{
		{Line: "[14:00:00] 1 1", Cmd: event.Command{Time: 14 * time.Hour, PlayerID: 1, EventID: 1}},
		{Line: "[14:00:00] 2 1", Cmd: event.Command{Time: 14 * time.Hour, PlayerID: 2, EventID: 1}},
		{Line: "[14:01:00] 2 2", Cmd: event.Command{Time: 14*time.Hour + time.Minute, PlayerID: 2, EventID: 2}},
	})
	model.CloseAt = 0
	model.Step()
	model.Step()
	model.Step()
	model.selectedID = 2

	view := model.dungeonView(44, 24)
	for _, want := range []string{"START", "00 FLOOR", "01 BOSS", "EXIT", "players:@1", "P:[@2]"} {
		if !strings.Contains(view, want) {
			t.Fatalf("dungeonView() =\n%s\nwant substring %q", view, want)
		}
	}
}

func TestUpdateUsesVimNavigationAndHelp(t *testing.T) {
	game.Cfg = game.Config{
		Floors:   2,
		Monsters: 2,
		OpenAt:   "14:05:00",
		Duration: 2,
	}
	replay.ResetPlayers()

	model := NewModel([]Entry{
		{Line: "[14:00:00] 1 1", Cmd: event.Command{Time: 14 * time.Hour, PlayerID: 1, EventID: 1}},
		{Line: "[14:01:00] 1 2", Cmd: event.Command{Time: 14*time.Hour + time.Minute, PlayerID: 1, EventID: 2}},
	})

	next, _ := model.Update(keyMsg('k'))
	model = next.(Model)
	if model.Index != 1 {
		t.Fatalf("k key index = %d, want 1", model.Index)
	}

	next, _ = model.Update(keyMsg('j'))
	model = next.(Model)
	if model.Index != 0 {
		t.Fatalf("j key index = %d, want 0", model.Index)
	}

	next, _ = model.Update(keyMsg('?'))
	model = next.(Model)
	if !model.ShowHelp {
		t.Fatal("? key did not open help")
	}
	help := model.View()
	if !strings.Contains(help, "l/k") || !strings.Contains(help, "h/j") {
		t.Fatalf("help view does not show vim navigation:\n%s", help)
	}

	next, _ = model.Update(keyMsg('l'))
	model = next.(Model)
	if model.Index != 0 {
		t.Fatalf("help view should ignore l key, index = %d, want 0", model.Index)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = next.(Model)
	if model.ShowHelp {
		t.Fatal("esc did not close help")
	}
}

func keyMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}
