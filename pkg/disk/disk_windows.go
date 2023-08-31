//go:build windows
// +build windows

package disk

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/Microsoft/go-winio/vhd"
)

func CreateDisk(path, format string, size int) error {
	if _, err := os.Stat(path); err == nil || !errors.Is(err, fs.ErrNotExist) {
		// disk already exists
		return err
	}

	if format != "vhdx" {
		return fmt.Errorf("format %q is not supported on windows, try 'vhdx'", format)
	}

	return vhd.CreateVhdx(path, uint32(size), 1)
}
