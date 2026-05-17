package format

import (
	"fmt"
	"impulse/internal/event"
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
	total := int(d.Seconds())
	h := total / 3600
	m := total % 3600 / 60
	s := total % 60
	return fmt.Sprintf("[%02d:%02d:%02d]", h, m, s)
}
