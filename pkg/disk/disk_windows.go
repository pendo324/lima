//go:build windows
// +build windows

package disk

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
)

// CreateDisk creates a disk with size at path. Currently only supports vhdx on Windows.
//
// winio.CreateVhdx only works when Hyper-V or Hyper-V PowerShell cmdlets is installed
// Since these features are not available on all editions of Windows, use diskpart instead
func CreateDisk(path, format string, size int) error {
	if _, err := os.Stat(path); err == nil || !errors.Is(err, fs.ErrNotExist) {
		// disk already exists
		return err
	}

	if format != "vhdx" {
		return fmt.Errorf("format %q is not supported on windows, try 'vhdx'", format)
	}

	// size must be in MiB
	sizeMiB := size / 1048576

	if sizeMiB < 3 {
		return fmt.Errorf("vhdx disks must be >= 3MiB on windows, got '%d'", sizeMiB)
	}

	// diskpart seems to use the filename to determine vhd vs vhdx
	// no extension seems to default to vhd
	pathExt := path + ".vhdx"

	script := fmt.Sprintf(`@"
create vdisk file="%s" type="expandable" maximum=%d
"@ | diskpart`, pathExt, sizeMiB)

	_, err := exec.Command("powershell.exe",
		"-nologo",
		"-noprofile",
		script,
	).CombinedOutput()

	if err != nil {
		return err
	}

	return nil
}
