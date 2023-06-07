package wsl

import (
	"fmt"
	"os/exec"
	"path"
	"strconv"

	"github.com/lima-vm/lima/pkg/store/filenames"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func wslCommand(args ...string) (string, error) {
	cmd := exec.Command("wsl.exe", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	decoded, _, err := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder(), out)
	if err != nil {
		return "", fmt.Errorf("failed to convert output from UTF16 when running wsl command wsl.exe %v, err: %w", args, err)
	}
	return string(decoded), nil
}

// startVM calls WSL to start a VM.
// Takes argument for VM name.
func startVM(name string) error {
	_, err := wslCommand("--distribution", "lima-"+name)
	if err != nil {
		return err
	}
	return nil
}

// initVM calls WSL to import a new VM specifically for Lima.
func initVM(name, instanceDir string) error {
	logrus.Infof("Importing distro from %q to %q", path.Join(instanceDir, filenames.WslRootFsDir), path.Join(instanceDir, filenames.WslRootFs))
	_, err := wslCommand("--import", "lima-"+name, path.Join(instanceDir, filenames.WslRootFsDir), path.Join(instanceDir, filenames.WslRootFs))
	if err != nil {
		return err
	}
	return nil
}

// stopVM calls WSL to stop a running VM.
// Takes arguments for name.
func stopVM(name string) error {
	_, err := wslCommand("--terminate", "lima-"+name)
	if err != nil {
		return err
	}
	return nil
}

func supportsWsl2() error {
	cmd := exec.Command("powershell.exe", "[System.Environment]::OSVersion.Version.Major")
	osMajorVerOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get OS major version: %w", err)
	}
	osMajorVer, err := strconv.Atoi(string(osMajorVerOut))
	if err != nil {
		return fmt.Errorf("failed to convert OS major version to int: %w", err)
	}
	if osMajorVer > 10 {
		return nil
	}
	if osMajorVer < 10 {
		return fmt.Errorf("wsl2 only supported on Windows versions 10 (build 19041) or 11")
	}
	if osMajorVer == 10 {
		cmd = exec.Command("powershell.exe", "[System.Environment]::OSVersion.Version.Build")
		osBuildOut, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to get OS build: %w", err)
		}
		osBuild, err := strconv.Atoi(string(osBuildOut))
		if err != nil {
			return fmt.Errorf("failed to convert OS build to int: %w", err)
		}
		if osBuild < 19041 {
			return fmt.Errorf("wsl2 only supported on Windows versions 10 (build 19041) or 11")
		}
		return nil
	}
	return nil
}
