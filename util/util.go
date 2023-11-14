package util

import (
	"os"
	"os/exec"
	"strings"
)

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func ExecuteCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	rawoutput, err := cmd.CombinedOutput()
	output := strings.ReplaceAll(string(rawoutput), "\n", "")
	return output, err
}
