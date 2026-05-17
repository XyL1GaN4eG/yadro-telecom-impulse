package parser

import (
	"testing"
	"time"
)

func TestSplitParsesReadmeLine(t *testing.T) {
	cmd, err := Split("[14:00:01] 7 9 out of mana")
	if err != nil {
		t.Fatal(err)
	}

	if cmd.Time != 14*time.Hour+time.Second {
		t.Fatalf("time = %v, want 14:00:01", cmd.Time)
	}
	if cmd.PlayerID != 7 {
		t.Fatalf("player id = %d, want 7", cmd.PlayerID)
	}
	if cmd.EventID != 9 {
		t.Fatalf("event id = %d, want 9", cmd.EventID)
	}
	if cmd.Arg != "out of mana" {
		t.Fatalf("arg = %q, want %q", cmd.Arg, "out of mana")
	}
}
