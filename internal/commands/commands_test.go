package commands

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Command
		isCommand bool
	}{
		{
			name:  "regular message",
			input: "hello /leave",
		},
		{
			name:      "leave without reason",
			input:     "/leave",
			want:      Command{Name: "leave", Args: []string{}},
			isCommand: true,
		},
		{
			name:      "leave with reason",
			input:     "/LEAVE see you later",
			want:      Command{Name: "leave", Args: []string{"see", "you", "later"}},
			isCommand: true,
		},
		{
			name:      "join channel",
			input:     "/join #go",
			want:      Command{Name: "join", Args: []string{"#go"}},
			isCommand: true,
		},
		{
			name:      "join channel with key",
			input:     "/JOIN #private secret",
			want:      Command{Name: "join", Args: []string{"#private", "secret"}},
			isCommand: true,
		},
		{
			name:      "list channels",
			input:     "/list",
			want:      Command{Name: "list", Args: []string{}},
			isCommand: true,
		},
		{
			name:      "list specific channel",
			input:     "/LIST #go",
			want:      Command{Name: "list", Args: []string{"#go"}},
			isCommand: true,
		},
		{
			name:      "unknown slash command",
			input:     "/unknown value",
			want:      Command{Name: "unknown", Args: []string{"value"}},
			isCommand: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, isCommand := Parse(tt.input)
			if isCommand != tt.isCommand {
				t.Fatalf("Parse() command = %v, want %v", isCommand, tt.isCommand)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
