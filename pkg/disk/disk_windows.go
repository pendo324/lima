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

func CreateDisk(path, format string, size int) error {
	if _, err := os.Stat(path); err == nil || !errors.Is(err, fs.ErrNotExist) {
		// disk already exists
		return err
	}

	if format != "vhdx" {
		return fmt.Errorf("format %q is not supported on windows, try 'vhdx'", format)
	}

	// diskpart seems to use the filename to determine vhd vs vhdx
	// no extension seems to default to vhd
	pathExt := path + ".vhdx"

	// size needs to be in MiB
	script := fmt.Sprintf(`@"
create vdisk file="%s" type="expandable" maximum=%d
"@ | diskpart`, pathExt, size/1048576)

	_, err := exec.Command("powershell.exe",
		"-nologo",
		"-noprofile",
		script,
	).CombinedOutput()

	if err != nil {
		return err
	}

	return os.Rename(pathExt, path)
}
