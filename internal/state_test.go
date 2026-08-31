package internal

import (
	"os"
	"testing"
)

func TestStateRoundTrip(t *testing.T) {
	old := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", old)
	os.Setenv("XDG_CONFIG_HOME", t.TempDir())

	state := State{MinID: 1, MaxID: 100, Start: 7, Current: 42, Step: 17, Emitted: 35, Done: false}
	if err := SaveState(state); err != nil {
		t.Fatal(err)
	}
	got, err := LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if got != state {
		t.Fatalf("got %#v", got)
	}
	if err := ResetState(); err != nil {
		t.Fatal(err)
	}
}
