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

func TestDirectMessagesFlushAndLoad(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	directMessages := &DirectMessages{}
	if !directMessages.Add("Libera", "AlertBot") {
		t.Fatal("first direct message was not added")
	}
	if directMessages.Add("libera", "alertbot") {
		t.Fatal("duplicate direct message was added")
	}
	if err := directMessages.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	loaded, err := LoadDirectMessages()
	if err != nil {
		t.Fatalf("LoadDirectMessages() error = %v", err)
	}
	if len(loaded.Users) != 1 || loaded.Users[0].User != "AlertBot" {
		t.Fatalf("loaded direct messages = %+v", loaded.Users)
	}
}
