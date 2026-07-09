package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	hookMarker     = "jot _post-commit-clear"
	hookBlockStart = "# >>> jot hook >>>"
	hookBlockEnd   = "# <<< jot hook <<<"
	legacyHookMarker = "gut _post-commit-clear"
)

func gitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--absolute-git-dir")
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "rev-parse", "--git-dir")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("not inside a git repository")
		}
		abs, absErr := filepath.Abs(strings.TrimSpace(string(out)))
		if absErr != nil {
			return "", fmt.Errorf("resolve git directory: %w", absErr)
		}
		return abs, nil
	}
	return strings.TrimSpace(string(out)), nil
}

func hookPath(gitDir string) string {
	return filepath.Join(gitDir, "hooks", "post-commit")
}

func hookInstalled(content string) bool {
	return strings.Contains(content, hookMarker) || strings.Contains(content, legacyHookMarker)
}

func hookInvocation() string {
	exe, err := os.Executable()
	if err != nil {
		return "jot _post-commit-clear 2>/dev/null || true"
	}
	exe = filepath.ToSlash(exe)
	quoted := "'" + strings.ReplaceAll(exe, "'", `'\''`) + "'"
	return quoted + " _post-commit-clear 2>/dev/null || true"
}

func freshHookContent() string {
	return "#!/bin/sh\n" + hookInvocation() + "\n"
}

func appendHookBlock(content string) string {
	block := "\n" + hookBlockStart + "\n" + hookInvocation() + "\n" + hookBlockEnd + "\n"
	if strings.HasSuffix(content, "\n") {
		return content + strings.TrimPrefix(block, "\n")
	}
	return content + block
}

type hookInstallResult int

const (
	hookAlreadyPresent hookInstallResult = iota
	hookInstalledFresh
	hookAppended
)

func installHook(gitDir string) (hookInstallResult, error) {
	path := hookPath(gitDir)
	hooksDir := filepath.Dir(path)

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return hookAlreadyPresent, fmt.Errorf("create hooks directory: %w", err)
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return hookAlreadyPresent, fmt.Errorf("read post-commit hook: %w", err)
		}
		if err := os.WriteFile(path, []byte(freshHookContent()), 0755); err != nil {
			return hookAlreadyPresent, fmt.Errorf("write post-commit hook: %w", err)
		}
		return hookInstalledFresh, nil
	}

	content := string(existing)
	if hookInstalled(content) {
		updated, changed := refreshHookInvocation(content)
		if changed {
			if err := os.WriteFile(path, []byte(updated), 0755); err != nil {
				return hookAlreadyPresent, fmt.Errorf("update post-commit hook: %w", err)
			}
			return hookAppended, nil
		}
		return hookAlreadyPresent, nil
	}

	updated := appendHookBlock(content)
	if err := os.WriteFile(path, []byte(updated), 0755); err != nil {
		return hookAlreadyPresent, fmt.Errorf("update post-commit hook: %w", err)
	}
	return hookAppended, nil
}

func ensureHook(gitDir string) error {
	result, err := installHook(gitDir)
	if err != nil {
		return err
	}
	switch result {
	case hookInstalledFresh, hookAppended:
		fmt.Println("(installed jot's post-commit hook in this repo)")
	}
	return nil
}

func refreshHookInvocation(content string) (string, bool) {
	want := hookInvocation()
	lines := strings.Split(content, "\n")
	changed := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, hookMarker) || strings.Contains(trimmed, legacyHookMarker) {
			if trimmed != want {
				lines[i] = want
				changed = true
			}
		}
	}
	if !changed {
		return content, false
	}
	result := strings.Join(lines, "\n")
	if strings.HasSuffix(content, "\n") && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result, true
}

func runGitLog(args []string) int {
	gitArgs := append([]string{"log"}, args...)
	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "git log: %v\n", err)
		return 1
	}
	return 0
}

func runGitAdd() int {
	cmd := exec.Command("git", "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "git add: %v\n", err)
		return 1
	}
	return 0
}

func runGitPush(args []string) int {
	gitArgs := append([]string{"push"}, args...)
	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "git push: %v\n", err)
		return 1
	}
	return 0
}

func gitCommit(message string) int {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "git commit: %v\n", err)
		return 1
	}
	return 0
}

func initHookMessage(result hookInstallResult) string {
	switch result {
	case hookInstalledFresh:
		return "installed fresh post-commit hook"
	case hookAppended:
		return "appended jot invocation to existing post-commit hook"
	default:
		return "post-commit hook already present"
	}
}
