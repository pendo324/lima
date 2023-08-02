package executil

import (
	"fmt"
	"os/exec"

	"github.com/lima-vm/lima/pkg/ioutilx"
)

func RunUTF16leCommand(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	outString, err := ioutilx.FromUTF16leToString(out)
	if err != nil {
		return "", fmt.Errorf("failed to convert output from UTF16 when running wsl command wsl.exe %v, err: %w", args, err)
	}
	return outString, nil
}
