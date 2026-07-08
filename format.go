package main

import "strings"

func formatCommitMessage(store Store) (string, bool) {
	if len(store.Messages) == 0 {
		return "", false
	}

	mainIdx := store.MainIndex
	if mainIdx < 0 || mainIdx >= len(store.Messages) {
		mainIdx = 0
	}

	subject := store.Messages[mainIdx].Text
	if len(store.Messages) == 1 {
		return subject, true
	}

	var bullets []string
	for i, msg := range store.Messages {
		if i == mainIdx {
			continue
		}
		bullets = append(bullets, "- "+msg.Text)
	}

	return subject + "\n\n" + strings.Join(bullets, "\n"), true
}
