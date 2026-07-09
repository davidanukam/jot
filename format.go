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

	mainMsg := store.Messages[mainIdx]
	var subject string
	var bullets []string

	if isMultiLine(mainMsg) {
		subject = messageFirstLine(mainMsg)
		bullets = append(bullets, subLinesForCopy(mainMsg)...)
		for i, msg := range store.Messages {
			if i == mainIdx {
				continue
			}
			bullets = append(bullets, bulletsForOtherMessage(msg)...)
		}
	} else {
		subject = mainMsg.Text
		for i, msg := range store.Messages {
			if i == mainIdx {
				continue
			}
			bullets = append(bullets, bulletsForOtherMessage(msg)...)
		}
	}

	return formatQuotedCommit(subject, bullets)
}

// commitMessageForGit returns the formatted commit message without shell quoting.
func commitMessageForGit(store Store) (string, bool) {
	formatted, ok := formatCommitMessage(store)
	if !ok {
		return "", false
	}
	if len(formatted) >= 2 && formatted[0] == '"' && formatted[len(formatted)-1] == '"' {
		return formatted[1 : len(formatted)-1], true
	}
	return formatted, true
}

func formatQuotedCommit(subject string, bullets []string) (string, bool) {
	if len(bullets) == 0 {
		return `"` + subject + `"`, true
	}
	return `"` + subject + "\n" + strings.Join(bullets, "\n") + `"`, true
}
