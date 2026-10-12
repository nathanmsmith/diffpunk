package server

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	program := "xdg-open"
	if runtime.GOOS == "darwin" {
		program = "open"
	}
	cmd := exec.Command(program, url)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, "diff punk: browser launcher failed; open the printed URL:", err)
		}
	}()
	return nil
}
