//go:build windows
// +build windows

package wsl

import (
	_ "embed"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/lima-vm/lima/pkg/driver"
	"github.com/lima-vm/lima/pkg/executil"
	"github.com/lima-vm/lima/pkg/store/filenames"
	"github.com/lima-vm/lima/pkg/textutil"
	"github.com/sirupsen/logrus"
)

// startVM calls WSL to start a VM.
// Takes argument for VM name.
func startVM(name string) error {
	_, err := executil.RunUTF16leCommand("wsl.exe", "--distribution", "lima-"+name)
	if err != nil {
		return err
	}
	return nil
}

// initVM calls WSL to import a new VM specifically for Lima.
func initVM(name, instanceDir string) error {
	logrus.Infof("Importing distro from %q to %q", path.Join(instanceDir, filenames.WslRootFsDir), path.Join(instanceDir, filenames.WslRootFs))
	_, err := executil.RunUTF16leCommand("wsl.exe", "--import", "lima-"+name, path.Join(instanceDir, filenames.WslRootFsDir), path.Join(instanceDir, filenames.WslRootFs))
	if err != nil {
		return err
	}
	return nil
}

// stopVM calls WSL to stop a running VM.
// Takes arguments for name.
func stopVM(name string) error {
	_, err := executil.RunUTF16leCommand("wsl.exe", "--terminate", "lima-"+name)
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

//go:embed lima-init.TEMPLATE.sh
var limaBoot string

func provisionVM(driver *driver.BaseDriver) error {
	ciDataPath := filepath.Join(driver.Instance.Dir, filenames.CIDataISODir)
	m := map[string]string{
		"CIDataPath": ciDataPath,
	}
	textutil.ExecuteTemplate(limaBoot, m)

	_, err := exec.Command(
		"wsl.exe",
		"-d",
		driver.Instance.DistroName,
		"bash",
		"-c",
		limaBoot,
	).CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"error running wslCommand that executes boot.sh: %w,"+
				"check /var/log/lima-init.log for more details", err)
	}
	return nil
}
