package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "0.1.0"

var reservedCommands = map[string]bool{
	"show":              true,
	"main":              true,
	"copy":              true,
	"rm":                true,
	"remove":            true,
	"edit":              true,
	"undo":              true,
	"clear":             true,
	"init":              true,
	"help":              true,
	"version":           true,
	"_post-commit-clear": true,
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	args := os.Args[1:]

	if args[0] == "--help" || args[0] == "-h" {
		printHelp()
		return
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Println(version)
		return
	}
	if args[0] == "-c" {
		os.Exit(runCopy(args[1:]))
	}

	if reservedCommands[args[0]] {
		os.Exit(dispatch(args))
		return
	}

	os.Exit(runLog(args))
}

func dispatch(args []string) int {
	switch args[0] {
	case "help":
		printHelp()
		return 0
	case "version":
		fmt.Println(version)
		return 0
	case "show":
		return runShow(args[1:])
	case "copy":
		return runCopy(args[1:])
	case "main":
		return runMain(args[1:])
	case "rm", "remove":
		return runRemove(args[1:])
	case "edit":
		return runEdit(args[1:])
	case "undo":
		return runUndo()
	case "clear":
		return runClear(args[1:])
	case "init":
		return runInit()
	case "_post-commit-clear":
		return runPostCommitClear()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		return 1
	}
}

func requireGitDir() (string, int) {
	dir, err := gitDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return "", 1
	}
	return dir, 0
}

func runLog(args []string) int {
	text := strings.TrimSpace(strings.Join(args, " "))
	if text == "" {
		fmt.Fprintln(os.Stderr, "message text cannot be empty")
		return 1
	}

	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	if err := ensureHook(gitDir); err != nil {
		fmt.Fprintf(os.Stderr, "install hook: %v\n", err)
		return 1
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	store.Messages = append(store.Messages, Message{Text: text, Time: time.Now()})
	if len(store.Messages) == 1 {
		store.MainIndex = 0
	}

	if err := saveStore(gitDir, store); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Printf("Logged #%d: %q\n", len(store.Messages)-1, text)
	return 0
}

func runShow(args []string) int {
	preview := hasFlag(args, "--preview")
	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if preview {
		return printFormatted(store)
	}

	if len(store.Messages) == 0 {
		fmt.Println("no pending gut notes since the last commit")
		return 0
	}

	for i, msg := range store.Messages {
		marker := ""
		if i == store.MainIndex {
			marker = " [main]"
		}
		fmt.Printf("#%d  %s%s  %q\n", i, formatRelativeTime(msg.Time), marker, msg.Text)
	}
	return 0
}

func runMain(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: gut main <n>")
		return 1
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid index: %q (must be a number)\n", args[0])
		return 1
	}

	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if n < 0 || n >= len(store.Messages) {
		fmt.Fprintf(os.Stderr, "index %d out of range (have %d message(s))\n", n, len(store.Messages))
		return 1
	}

	store.MainIndex = n
	if err := saveStore(gitDir, store); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Printf("main message set to #%d\n", n)
	return 0
}

func runCopy(args []string) int {
	preview := hasFlag(args, "--preview")

	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	formatted, ok := formatCommitMessage(store)
	if !ok {
		fmt.Println("no pending gut notes since the last commit")
		return 0
	}

	fmt.Println(formatted)
	if preview {
		return 0
	}

	if err := copyToClipboard(formatted); err != nil {
		fmt.Fprintf(os.Stderr, "clipboard: %v\n", err)
		fmt.Fprintln(os.Stderr, "message was printed above; install a clipboard tool or copy it manually")
		return 0
	}
	return 0
}

func runRemove(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: gut rm <n>")
		return 1
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid index: %q (must be a number)\n", args[0])
		return 1
	}

	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if n < 0 || n >= len(store.Messages) {
		fmt.Fprintf(os.Stderr, "index %d out of range (have %d message(s))\n", n, len(store.Messages))
		return 1
	}

	mainReset := false
	if n == store.MainIndex {
		store.MainIndex = 0
		mainReset = true
	} else if n < store.MainIndex {
		store.MainIndex--
	}

	store.Messages = append(store.Messages[:n], store.Messages[n+1:]...)
	if len(store.Messages) == 0 {
		store.MainIndex = 0
	} else if store.MainIndex >= len(store.Messages) {
		store.MainIndex = 0
	}

	if err := saveStore(gitDir, store); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Printf("removed message #%d\n", n)
	if mainReset {
		fmt.Println("main message reset to #0")
	}
	return 0
}

func runEdit(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gut edit <n> <new text...>")
		return 1
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid index: %q (must be a number)\n", args[0])
		return 1
	}

	text := strings.TrimSpace(strings.Join(args[1:], " "))
	if text == "" {
		fmt.Fprintln(os.Stderr, "message text cannot be empty")
		return 1
	}

	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if n < 0 || n >= len(store.Messages) {
		fmt.Fprintf(os.Stderr, "index %d out of range (have %d message(s))\n", n, len(store.Messages))
		return 1
	}

	store.Messages[n].Text = text
	if err := saveStore(gitDir, store); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Printf("updated message #%d\n", n)
	return 0
}

func runUndo() int {
	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if len(store.Messages) == 0 {
		fmt.Fprintln(os.Stderr, "no messages to undo")
		return 1
	}

	n := len(store.Messages) - 1
	return runRemove([]string{strconv.Itoa(n)})
}

func runClear(args []string) int {
	yes := hasFlag(args, "--yes") || hasFlag(args, "-y")

	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	store, err := loadStore(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if len(store.Messages) == 0 {
		fmt.Println("no pending gut notes since the last commit")
		return 0
	}

	if !yes {
		fmt.Printf("clear %d pending message(s)? [y/N] ", len(store.Messages))
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "read confirmation:", err)
			return 1
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			fmt.Println("cancelled")
			return 0
		}
	}

	if err := clearStore(gitDir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Println("cleared all pending messages")
	return 0
}

func runInit() int {
	gitDir, code := requireGitDir()
	if code != 0 {
		return code
	}

	result, err := installHook(gitDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Println(initHookMessage(result))
	return 0
}

func runPostCommitClear() int {
	gitDir, err := gitDir()
	if err != nil {
		return 0
	}
	_ = clearStore(gitDir)
	return 0
}

func printFormatted(store Store) int {
	formatted, ok := formatCommitMessage(store)
	if !ok {
		fmt.Println("no pending gut notes since the last commit")
		return 0
	}
	fmt.Println(formatted)
	return 0
}

func formatRelativeTime(t time.Time) string {
	elapsed := time.Since(t)
	if elapsed < time.Minute {
		return "just now"
	}
	if elapsed < time.Hour {
		m := int(elapsed / time.Minute)
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	}
	if elapsed < 24*time.Hour {
		h := int(elapsed / time.Hour)
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	}
	return t.Format("2006-01-02 15:04")
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func printHelp() {
	fmt.Printf(`gut %s — commit-message journal

Usage:
  gut "<message>"          Log a note (multiple words without quotes are joined)
  gut show                 List pending notes
  gut show --preview       Print formatted commit message without clipboard
  gut main <n>             Set which note becomes the commit subject line
  gut copy                 Format and copy commit message to clipboard
  gut copy --preview       Print formatted commit message without clipboard
  gut -c                   Alias for gut copy
  gut rm <n>               Remove note at index n
  gut remove <n>           Alias for gut rm
  gut edit <n> <text...>   Replace note text at index n
  gut undo                 Remove the most recently added note
  gut clear                Wipe all pending notes (prompts unless -y/--yes)
  gut init                 Install or repair the post-commit hook
  gut help                 Show this help
  gut version              Print version

Notes are stored per-repo in .git/gut/log.json and cleared automatically
after each successful commit via the post-commit hook.
`, version)
}
