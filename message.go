package main

import (
	"fmt"
	"strings"
)

func splitMessageLines(text string) []string {
	text = strings.TrimRight(text, "\n\r")
	if text == "" {
		return nil
	}
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, len(raw))
	for _, l := range raw {
		lines = append(lines, strings.TrimRight(l, "\r"))
	}
	return lines
}

func isMultiLine(msg Message) bool {
	return strings.Contains(msg.Text, "\n")
}

func messageFirstLine(msg Message) string {
	lines := splitMessageLines(msg.Text)
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

func hasHyphenPrefix(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(trimmed, "-")
}

func ensureHyphenPrefix(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "- ") {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "-") {
		rest := strings.TrimLeft(strings.TrimPrefix(trimmed, "-"), " \t")
		return "- " + rest
	}
	return "- " + trimmed
}

func formatSubLineForStorage(line string, noHyphen bool) string {
	line = strings.TrimRight(line, " \t\r")
	if noHyphen {
		return line
	}
	if hasHyphenPrefix(line) {
		return ensureHyphenPrefix(line)
	}
	return "- " + strings.TrimLeft(line, " \t")
}

func processWriteText(text string, noHyphen bool) string {
	text = strings.TrimRight(text, " \t\r\n")
	text = strings.TrimLeft(text, " \t\r")
	lines := splitMessageLines(text)
	if len(lines) == 0 {
		return ""
	}
	if len(lines) == 1 {
		return lines[0]
	}
	result := []string{lines[0]}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		result = append(result, formatSubLineForStorage(line, noHyphen))
	}
	return strings.Join(result, "\n")
}

func subLinesForCopy(msg Message) []string {
	lines := splitMessageLines(msg.Text)
	if len(lines) <= 1 {
		return nil
	}
	bullets := make([]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		bullets = append(bullets, ensureHyphenPrefix(line))
	}
	return bullets
}

func allLinesAsBulletsForCopy(msg Message) []string {
	lines := splitMessageLines(msg.Text)
	bullets := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		bullets = append(bullets, ensureHyphenPrefix(line))
	}
	return bullets
}

func bulletsForOtherMessage(msg Message) []string {
	if isMultiLine(msg) {
		return allLinesAsBulletsForCopy(msg)
	}
	return []string{ensureHyphenPrefix(msg.Text)}
}

func stripHyphenForRead(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "- ") {
		return strings.TrimLeft(trimmed[2:], " \t")
	}
	if strings.HasPrefix(trimmed, "-") {
		return strings.TrimLeft(trimmed[1:], " \t")
	}
	return line
}

func printReadMessageBody(prefix string, msg Message) {
	lines := splitMessageLines(msg.Text)
	if len(lines) <= 1 {
		fmt.Printf("%s%q\n", prefix, msg.Text)
		return
	}

	indent := strings.Repeat(" ", len(prefix)+1)
	fmt.Printf("%s\"%s\n", prefix, lines[0])
	for j := 1; j < len(lines); j++ {
		line := lines[j]
		if msg.NoHyphen {
			line = stripHyphenForRead(line)
		}
		suffix := ""
		if j == len(lines)-1 {
			suffix = `"`
		}
		fmt.Printf("%s%s%s\n", indent, line, suffix)
	}
}
