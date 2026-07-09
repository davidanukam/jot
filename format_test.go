package main

import (
	"testing"
	"time"
)

func TestFormatCommitMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		store  Store
		want   string
		wantOK bool
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
			want:   `"Only subject"`,
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
			want:   "\"Subject line\n- Detail one\n- Detail two\"",
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
			want:   "\"Refactored pipeline\n- Added circles\"",
			wantOK: true,
		},
		{
			name: "multi-line main with other messages",
			store: Store{
				Messages: []Message{
					{Text: "Testing this out\n- added some nice things\n- each new thing is now cooler than before", Time: now},
					{Text: "Updated jot message command", Time: now},
					{Text: "Updated README file", Time: now},
				},
				MainIndex: 0,
			},
			want: "\"Testing this out\n- added some nice things\n- each new thing is now cooler than before\n- Updated jot message command\n- Updated README file\"",
			wantOK: true,
		},
		{
			name: "single-line main with multi-line other",
			store: Store{
				Messages: []Message{
					{Text: "Updated jot message command", Time: now},
					{Text: "Testing this out\n- added some nice things\n- each new thing is now cooler than before", Time: now},
					{Text: "Updated README file", Time: now},
				},
				MainIndex: 0,
			},
			want: "\"Updated jot message command\n- Testing this out\n- added some nice things\n- each new thing is now cooler than before\n- Updated README file\"",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := formatCommitMessage(tt.store)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommitMessageForGit(t *testing.T) {
	now := time.Now()
	store := Store{
		Messages: []Message{
			{Text: "Subject line", Time: now},
			{Text: "Detail one", Time: now},
		},
		MainIndex: 0,
	}

	got, ok := commitMessageForGit(store)
	if !ok {
		t.Fatal("expected ok")
	}
	want := "Subject line\n- Detail one"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestProcessWriteText(t *testing.T) {
	got := processWriteText("Testing this out\nadded some nice things\neach new thing is now cooler than before", false)
	want := "Testing this out\n- added some nice things\n- each new thing is now cooler than before"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	gotNoHyphen := processWriteText("Testing this out\nadded some nice things", true)
	wantNoHyphen := "Testing this out\nadded some nice things"
	if gotNoHyphen != wantNoHyphen {
		t.Fatalf("got %q, want %q", gotNoHyphen, wantNoHyphen)
	}
}
