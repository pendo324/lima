package wsl

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/lima-vm/lima/pkg/driver"
	"github.com/lima-vm/lima/pkg/ioutilx"
	"github.com/lima-vm/lima/pkg/store/filenames"
	"github.com/sirupsen/logrus"
)

func wslCommand(args ...string) (string, error) {
	cmd := exec.Command("wsl.exe", args...)
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

func powershellCommand(args ...string) (string, error) {
	cmd := exec.Command("powershell.exe", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	outString, err := ioutilx.FromUTF16leToString(out)
	if err != nil {
		return "", fmt.Errorf("failed to convert output from UTF16 when running powershell command powershell.exe %v, err: %w", args, err)
	}
	return outString, nil
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

func attachDisks(driver *driver.BaseDriver) error {
	ciDataPath := filepath.Join(driver.Instance.Dir, filenames.CIDataISO)

	logrus.Infof("Attaching cidata...")
	logrus.Infof("creating cidata dir in distro %s...", driver.Instance.DistroName)
	_, err := wslCommand("-d", driver.Instance.DistroName, "mkdir", "/mnt/lima-cidata")
	if err != nil {
		return fmt.Errorf("failed to create mount path in VM %s: %w", driver.Instance.Name, err)
	}
	logrus.Infof("mounting cidata in distro %s...", driver.Instance.DistroName)
	_, err = wslCommand("-d", driver.Instance.DistroName, "mount", "-t", "iso9660", ciDataPath, "/mnt/lima-cidata")
	if err != nil {
		return fmt.Errorf("failed to create mount path in VM %s: %w", driver.Instance.Name, err)
	}
	logrus.Infof("cidata mounted!")
	return nil
}
