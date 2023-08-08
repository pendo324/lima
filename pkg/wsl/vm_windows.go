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
func startVM(ctx context.Context, driver *driver.BaseDriver) error {
	_, err := executil.RunUTF16leCommand([]string{
		"wsl.exe",
		"--distribution",
		driver.Instance.DistroName,
	}, executil.WithContext(&ctx))
	if err != nil {
		return err
	}
	return nil
}

// initVM calls WSL to import a new VM specifically for Lima.
func initVM(ctx context.Context, driver *driver.BaseDriver) error {
	rootFSPath := path.Join(driver.Instance.Dir, filenames.WslRootFsDir)
	rootFSDir := path.Join(driver.Instance.Dir, filenames.WslRootFs)
	logrus.Infof("Importing distro from %q to %q", rootFSPath, rootFSDir)
	_, err := executil.RunUTF16leCommand([]string{
		"wsl.exe",
		"--import",
		driver.Instance.DistroName,
		rootFSPath,
		rootFSDir,
	}, executil.WithContext(&ctx))
	if err != nil {
		return err
	}
	return nil
}

// stopVM calls WSL to stop a running VM.
// Takes arguments for name.
func stopVM(ctx context.Context, driver *driver.BaseDriver) error {
	_, err := executil.RunUTF16leCommand([]string{
		"wsl.exe",
		"--terminate",
		driver.Instance.DistroName,
	}, executil.WithContext(&ctx))
	if err != nil {
		return err
	}
	return nil
}

//go:embed lima-init.TEMPLATE.sh
var limaBoot string

func provisionVM(ctx context.Context, driver *driver.BaseDriver, errCh *chan error) error {
	ciDataPath := filepath.Join(driver.Instance.Dir, filenames.CIDataISODir)
	m := map[string]string{
		"CIDataPath": ciDataPath,
	}
	out, err := textutil.ExecuteTemplate(limaBoot, m)
	if err != nil {
		return fmt.Errorf("failed to construct wsl boot.sh script: %w", err)
	}
	outString := strings.Replace(string(out), `\r\n`, `\n`, -1)

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
			*errCh <- fmt.Errorf(
				"error running wslCommand that executes boot.sh: %w, "+
					"check /var/log/lima-init.log for more details", err)
		}

		for {
			select {
			case <-ctx.Done():
				logrus.Info("Context closed, stopping vm")
				if status, err := store.GetWslStatus(driver.Instance.Name, driver.Instance.DistroName); err == nil &&
					status == store.StatusRunning {
					stopVM(ctx, driver)
				}
			}
		}
	}()

	return err
}

func keepAlive(ctx context.Context, driver *driver.BaseDriver, errCh *chan error) {
	keepAliveCmd := exec.CommandContext(
		ctx,
		"wsl.exe",
		"-d",
		driver.Instance.DistroName,
		"bash",
		"-c",
		"nohup sleep 2147483647d >/dev/null 2>&1",
	)

	go func() {
		if err := keepAliveCmd.Run(); err != nil {
			*errCh <- fmt.Errorf(
				"error running wsl keepAlive command: %w", err)
		}
	}()
}
