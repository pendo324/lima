//go:build windows
// +build windows

package wsl

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lima-vm/lima/pkg/driver"
	"github.com/lima-vm/lima/pkg/executil"
	"github.com/lima-vm/lima/pkg/store"
	"github.com/lima-vm/lima/pkg/store/filenames"
	"github.com/lima-vm/lima/pkg/textutil"
	"github.com/sirupsen/logrus"
)

// startVM calls WSL to start a VM.
// Takes argument for VM name.
func startVM(ctx context.Context, name string) error {
	_, err := executil.RunUTF16leCommand([]string{
		"wsl.exe",
		"--distribution",
		"lima-" + name,
	}, executil.WithContext(&ctx))
	if err != nil {
		return err
	}
	return nil
}

// initVM calls WSL to import a new VM specifically for Lima.
func initVM(ctx context.Context, name, instanceDir string) error {
	logrus.Infof("Importing distro from %q to %q", path.Join(instanceDir, filenames.WslRootFsDir), path.Join(instanceDir, filenames.WslRootFs))
	_, err := executil.RunUTF16leCommand([]string{
		"wsl.exe",
		"--import",
		"lima-" + name,
		path.Join(instanceDir, filenames.WslRootFsDir),
		path.Join(instanceDir, filenames.WslRootFs),
	}, executil.WithContext(&ctx))
	if err != nil {
		return err
	}
	return nil
}

// stopVM calls WSL to stop a running VM.
// Takes arguments for name.
func stopVM(name string) error {
	_, err := executil.RunUTF16leCommand([]string{
		"wsl.exe",
		"--terminate",
		"lima-" + name,
	})
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

func provisionVM(ctx context.Context, driver *driver.BaseDriver) (chan error, error) {
	ciDataPath := filepath.Join(driver.Instance.Dir, filenames.CIDataISODir)
	m := map[string]string{
		"CIDataPath": ciDataPath,
	}
	out, err := textutil.ExecuteTemplate(limaBoot, m)
	if err != nil {
		return nil, fmt.Errorf("failed to construct wsl boot.sh script: %w", err)
	}
	outString := strings.Replace(string(out), `\r\n`, `\n`, -1)

	errCh := make(chan error)

	go func() {
		cmd := exec.CommandContext(
			ctx,
			"wsl.exe",
			"-d",
			driver.Instance.DistroName,
			"bash",
			"-c",
			outString,
		)
		if _, err := cmd.CombinedOutput(); err != nil {
			errCh <- fmt.Errorf(
				"error running wslCommand that executes boot.sh: %w, "+
					"check /var/log/lima-init.log for more details", err)
		}

		for {
			select {
			case <-ctx.Done():
				logrus.Info("Context closed, stopping vm")
				if status, err := store.GetWslStatus(driver.Instance.Name, driver.Instance.DistroName); err == nil &&
					status == store.StatusRunning {
					stopVM(driver.Instance.Name)
				}
			}
		}
	}()

	return errCh, err
}
