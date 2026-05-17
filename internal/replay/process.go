package replay

import (
	"encoding/json"
	"errors"
	"impulse/internal/event"
	outfmt "impulse/internal/format"
	"impulse/internal/game"
	"impulse/internal/handler"
	"impulse/internal/player"
	"os"
)

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &game.Cfg)
}

func ResetPlayers() {
	player.Players = make(map[uint8]player.Player)
}

func ProcessCommand(cmd event.Command) []string {
	wasBossFloorEntered := false
	if p, ok := player.Players[cmd.PlayerID]; ok {
		wasBossFloorEntered = p.Dungeon.BossFloorEntered
	}

	p, err := handler.HandleCommand(cmd)
	if err != nil {
		if errors.Is(err, handler.ErrNoOutput) {
			return nil
		}
		if cmd.EventID >= 4 && cmd.EventID <= 7 {
			return []string{outfmt.ImpossibleMove(cmd)}
		}
		return nil
	}

	if p.Status == player.StatusDisqual {
		return []string{outfmt.Disqualified(cmd)}
	}

	lines := []string{outfmt.Command(cmd)}
	if cmd.EventID == 4 && !wasBossFloorEntered && p.Dungeon.BossFloorEntered {
		lines = append(lines, outfmt.Command(event.Command{
			Time:     cmd.Time,
			PlayerID: cmd.PlayerID,
			EventID:  6,
		}))
	}
	if cmd.EventID == 11 && p.Health == 0 {
		lines = append(lines, outfmt.Dead(cmd))
	}
	return lines
}
