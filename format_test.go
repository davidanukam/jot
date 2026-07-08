package main

import (
	"strings"
	"testing"
	"time"
)

func TestFormatCommitMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		store   Store
		want    string
		wantOK  bool
	}{
		{
			name:   "empty store",
			store:  Store{MainIndex: 0},
			want:   "",
			wantOK: false,
		},
		{
			name: "single message",
			store: Store{
				Messages:  []Message{{Text: "Only subject", Time: now}},
				MainIndex: 0,
			},
			want:   "Only subject",
			wantOK: true,
		},
		{
			name: "multiple messages main at zero",
			store: Store{
				Messages: []Message{
					{Text: "Subject line", Time: now},
					{Text: "Detail one", Time: now},
					{Text: "Detail two", Time: now},
				},
				MainIndex: 0,
			},
			want:   "Subject line\n\n- Detail one\n- Detail two",
			wantOK: true,
		},
		{
			name: "main not at index zero",
			store: Store{
				Messages: []Message{
					{Text: "Added circles", Time: now},
					{Text: "Refactored pipeline", Time: now},
				},
				MainIndex: 1,
			},
			want:   "Refactored pipeline\n\n- Added circles",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := formatCommitMessage(tt.store)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
