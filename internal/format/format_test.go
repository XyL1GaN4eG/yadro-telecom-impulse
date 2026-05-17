package format

import (
	"impulse/internal/event"
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
