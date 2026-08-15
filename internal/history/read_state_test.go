package history

import "testing"

func TestReadStateFlushAndLoad(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	state, err := LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState() error = %v", err)
	}
	if _, ok := state.Marker("libera", "#go"); ok {
		t.Fatal("missing marker unexpectedly found")
	}

	want := ReadMarker{ServerTime: 42, MsgID: "abc"}
	state.MarkRead("Libera", "#Go", want)
	if err := state.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	loaded, err := LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState() after flush error = %v", err)
	}
	if got, ok := loaded.Marker("libera", "#go"); !ok || got != want {
		t.Fatalf("Marker() = (%+v, %v), want (%+v, true)", got, ok, want)
	}
}
