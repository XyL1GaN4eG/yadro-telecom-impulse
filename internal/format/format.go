package format

import (
	"fmt"
	"impulse/internal/event"
	"impulse/internal/player"
	"sort"
	"strings"
	"time"
)

func Command(cmd event.Command) string {
	switch cmd.EventID {
	case 1:
		return fmt.Sprintf("%s Player [%d] registered", Time(cmd.Time), cmd.PlayerID)
	case 2:
		return fmt.Sprintf("%s Player [%d] entered the dungeon", Time(cmd.Time), cmd.PlayerID)
	case 3:
		return fmt.Sprintf("%s Player [%d] killed the monster", Time(cmd.Time), cmd.PlayerID)
	case 4:
		return fmt.Sprintf("%s Player [%d] went to the next floor", Time(cmd.Time), cmd.PlayerID)
	case 5:
		return fmt.Sprintf("%s Player [%d] went to the previous floor", Time(cmd.Time), cmd.PlayerID)
	case 6:
		return fmt.Sprintf("%s Player [%d] entered the boss's floor", Time(cmd.Time), cmd.PlayerID)
	case 7:
		return fmt.Sprintf("%s Player [%d] killed the boss", Time(cmd.Time), cmd.PlayerID)
	case 8:
		return fmt.Sprintf("%s Player [%d] left the dungeon", Time(cmd.Time), cmd.PlayerID)
	case 9:
		return fmt.Sprintf("%s Player [%d] cannot continue due to [%s]", Time(cmd.Time), cmd.PlayerID, cmd.Arg)
	case 10:
		return fmt.Sprintf("%s Player [%d] has restored [%s] of health", Time(cmd.Time), cmd.PlayerID, cmd.Arg)
	case 11:
		return fmt.Sprintf("%s Player [%d] recieved [%s] of damage", Time(cmd.Time), cmd.PlayerID, cmd.Arg)
	default:
		return ""
	}
}

func Disqualified(cmd event.Command) string {
	return fmt.Sprintf("%s Player [%d] is disqualified", Time(cmd.Time), cmd.PlayerID)
}

func Dead(cmd event.Command) string {
	return fmt.Sprintf("%s Player [%d] is dead", Time(cmd.Time), cmd.PlayerID)
}

func ImpossibleMove(cmd event.Command) string {
	return fmt.Sprintf("%s Player [%d] makes imposible move [%d]", Time(cmd.Time), cmd.PlayerID, cmd.EventID)
}

func Time(d time.Duration) string {
	return fmt.Sprintf("[%s]", Duration(d))
}

func Duration(d time.Duration) string {
	total := int(d.Seconds())
	h := total / 3600
	m := total % 3600 / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func FinalReport(players map[uint8]player.Player) string {
	ids := make([]int, 0, len(players))
	for id := range players {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	lines := make([]string, 0, len(players)+1)
	lines = append(lines, "Final report:")
	for _, id := range ids {
		lines = append(lines, Report(players[uint8(id)]))
	}

	return strings.Join(lines, "\n") + "\n"
}

func Report(p player.Player) string {
	return fmt.Sprintf(
		"[%s] %d [%s, %s, %s] HP:%d",
		Status(p.Status),
		p.ID,
		Duration(p.TotalTime()),
		Duration(p.AverageFloorClearTime()),
		Duration(p.Dungeon.BossKillTime),
		p.Health,
	)
}

func Status(status player.Status) string {
	switch status {
	case player.StatusSuccess:
		return "SUCCESS"
	case player.StatusDisqual:
		return "DISQUAL"
	default:
		return "FAIL"
	}
}
