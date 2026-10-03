package notify

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestNotifyArgsMatchesNotificationsSpec(t *testing.T) {
	got := notifyArgs(Notification{Summary: "alice in #go", Body: "hello"})
	want := []any{
		"dex",
		uint32(0),
		"",
		"alice in #go",
		"hello",
		[]string{},
		map[string]dbus.Variant{"category": dbus.MakeVariant("im.received")},
		int32(-1),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("notifyArgs() = %#v, want %#v", got, want)
	}
}

func TestNotifyArgsEscapesBodyOnly(t *testing.T) {
	args := notifyArgs(Notification{Summary: "a&b <c>", Body: "x & <b>y</b>"})
	if got, want := args[3], "a&b <c>"; got != want {
		t.Fatalf("summary = %q, want %q", got, want)
	}
	if got, want := args[4], "x &amp; &lt;b&gt;y&lt;/b&gt;"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestFormatBodyTruncation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "at limit is unchanged",
			body: strings.Repeat("é", maxBodyRunes),
			want: strings.Repeat("é", maxBodyRunes),
		},
		{
			name: "over limit is cut to limit with ellipsis",
			body: strings.Repeat("é", maxBodyRunes+1),
			want: strings.Repeat("é", maxBodyRunes-1) + "…",
		},
		{
			name: "truncation happens before escaping",
			body: strings.Repeat("&", maxBodyRunes+1),
			want: strings.Repeat("&amp;", maxBodyRunes-1) + "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBody(tt.body); got != tt.want {
				t.Fatalf("formatBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNotifyReportsConnectionErrorAsUnavailable(t *testing.T) {
	connectErr := errors.New("no session bus")
	d := &DBus{err: connectErr}

	err := d.Notify(context.Background(), Notification{Summary: "alice"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Notify() error = %v, want ErrUnavailable", err)
	}
	if !errors.Is(err, connectErr) {
		t.Fatalf("Notify() error = %v, want wrapped connection error", err)
	}
}

func TestCloseWithoutConnection(t *testing.T) {
	if err := (&DBus{err: errors.New("no session bus")}).Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}
