package main

import "testing"

func TestEnsureHyphenPrefix(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"added things", "- added things"},
		{"- already", "- already"},
		{"-already", "- already"},
	}
	for _, tt := range tests {
		if got := ensureHyphenPrefix(tt.in); got != tt.want {
			t.Fatalf("ensureHyphenPrefix(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
