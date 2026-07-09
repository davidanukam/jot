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

const noPendingMessages = "No messages have been made for the next commit"

var reservedCommands = map[string]bool{
	"read":              true,
	"log":               true,
	"main":              true,
	"copy":              true,
	"paste":             true,
	"update":            true,
	"rm":                true,
	"remove":            true,
	"edit":              true,
	"undo":              true,
	"clear":             true,
	"init":              true,
	"write":             true,
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

	fmt.Fprintln(os.Stderr, `usage: jot write "message"`)
	os.Exit(1)
}

func dispatch(args []string) int {
	switch args[0] {
	case "help":
		printHelp()
		return 0
	case "version":
		fmt.Println(version)
		return 0
	case "read":
		return runRead(args[1:])
	case "log":
		return runJotLog(args[1:])
	case "copy":
		return runCopy(args[1:])
	case "paste":
		return runPaste(args[1:])
	case "update":
		return runUpdate()
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
	case "write":
		return runWrite(args[1:])
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

func runWrite(args []string) int {
	noHyphen := false
	var msgArgs []string
	for _, a := range args {
		if a == "--no-hyphen" {
			noHyphen = true
			continue
		}
		msgArgs = append(msgArgs, a)
	}

	if len(msgArgs) == 0 {
		fmt.Fprintln(os.Stderr, `usage: jot write "message"`)
		return 1
	}
	if len(msgArgs) > 1 {
		fmt.Fprintln(os.Stderr, `message must be surrounded by quotation marks (usage: jot write "message")`)
		return 1
	}

	text := processWriteText(msgArgs[0], noHyphen)
	return appendMessage(text, noHyphen)
}

func appendMessage(text string, noHyphen bool) int {
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

	store.Messages = append(store.Messages, Message{
		Text:     text,
		Time:     time.Now(),
		NoHyphen: noHyphen,
	})
	if len(store.Messages) == 1 {
		store.MainIndex = 0
	}

	if err := saveStore(gitDir, store); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Printf("Wrote #%d: %q\n", len(store.Messages)-1, text)
	return 0
}

func runRead(args []string) int {
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
		fmt.Println(noPendingMessages)
		return 0
	}

	maxPrefix := 0
	for i, msg := range store.Messages {
		if i == store.MainIndex {
			continue
		}
		prefix := fmt.Sprintf("#%d  %s  ", i, formatGitLogDate(msg.Time))
		if len(prefix) > maxPrefix {
			maxPrefix = len(prefix)
		}
	}

	for i, msg := range store.Messages {
		timeStr := formatGitLogDate(msg.Time)
		if i == store.MainIndex {
			prefix := fmt.Sprintf("#%d  %s [main]  ", i, timeStr)
			printReadMessageBody(prefix, msg)
			continue
		}
		prefix := fmt.Sprintf("#%d  %s  ", i, timeStr)
		padded := fmt.Sprintf("%-*s", maxPrefix, prefix)
		printReadMessageBody(padded, msg)
	}
	return 0
}

func runMain(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: jot main <n>")
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

func runPaste(args []string) int {
	preview := hasFlag(args, "-p") || hasFlag(args, "--preview")

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
		fmt.Println(noPendingMessages)
		return 0
	}

	fmt.Println(formatted)

	if preview {
		fmt.Print("commit with this message? [y/N] ")
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

	if err := copyToClipboard(formatted); err != nil {
		fmt.Fprintf(os.Stderr, "clipboard: %v\n", err)
		fmt.Fprintln(os.Stderr, "message was printed above; install a clipboard tool or copy it manually")
	}

	message, ok := commitMessageForGit(store)
	if !ok {
		fmt.Println(noPendingMessages)
		return 0
	}

	code = gitCommit(message)
	if code != 0 {
		return code
	}
	if err := clearStore(gitDir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
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
		fmt.Println(noPendingMessages)
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
		fmt.Fprintln(os.Stderr, "usage: jot rm <n>")
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
		fmt.Fprintln(os.Stderr, "usage: jot edit <n> <new text...>")
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
		fmt.Println(noPendingMessages)
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

func runJotLog(args []string) int {
	if _, code := requireGitDir(); code != 0 {
		return code
	}
	return runGitLog(args)
}

func runUpdate() int {
	if _, code := requireGitDir(); code != 0 {
		return code
	}
	return runGitAdd()
}

func printFormatted(store Store) int {
	formatted, ok := formatCommitMessage(store)
	if !ok {
		fmt.Println(noPendingMessages)
		return 0
	}
	fmt.Println(formatted)
	return 0
}

func formatGitLogDate(t time.Time) string {
	return t.Format("Mon Jan 2 15:04:05 2006 -0700")
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
	fmt.Printf(`jot %s — commit-message journal

Usage:
  jot write "message"      Log a note (supports multi-line quoted text)
  jot write --no-hyphen "message"
                           Log without auto-hyphens on sub-lines (copy always uses them)
  jot read                 List pending notes for the next commit
  jot read --preview       Print formatted commit message without clipboard
  jot log                  Show past commit messages (runs git log)
  jot main <n>             Set which note becomes the commit subject line
  jot copy                 Format and copy commit message to clipboard
  jot copy --preview       Print formatted commit message without clipboard
  jot -c                   Alias for jot copy
  jot paste                Copy message to clipboard and git commit
  jot paste -p             Preview message, confirm, then copy and commit
  jot update               Stage all changes (runs git add .)
  jot rm <n>               Remove note at index n
  jot remove <n>           Alias for jot rm
  jot edit <n> <text...>   Replace note text at index n
  jot undo                 Remove the most recently added note
  jot clear                Wipe all pending notes (prompts unless -y/--yes)
  jot init                 Install or repair the post-commit hook
  jot help                 Show this help
  jot version              Print version

Notes are stored per-repo in .git/jot/log.json and cleared automatically
after each successful commit via the post-commit hook.
`, version)
}
