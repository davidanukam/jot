package main

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

func copyToClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return pipeToCommand("pbcopy", text)
	case "windows":
		return pipeToCommand("clip", text)
	default:
		candidates := []struct {
			name string
			args []string
		}{
			{"wl-copy", nil},
			{"xclip", []string{"-selection", "clipboard"}},
			{"xsel", []string{"--clipboard", "--input"}},
		}

		var tried []string
		for _, c := range candidates {
			tried = append(tried, c.name)
			if _, err := exec.LookPath(c.name); err != nil {
				continue
			}
			return pipeToCommand(c.name, text, c.args...)
		}
		return fmt.Errorf("no clipboard tool found (tried: %s)", joinNames(tried))
	}
}

func pipeToCommand(name, text string, args ...string) error {
	cmd := exec.Command(name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := io.WriteString(stdin, text); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return err
	}
	if err := stdin.Close(); err != nil {
		_ = cmd.Wait()
		return err
	}
	return cmd.Wait()
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	out := names[0]
	for i := 1; i < len(names); i++ {
		out += ", " + names[i]
	}
	return out
}
