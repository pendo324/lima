//go:build !windows
// +build !windows

package store

import (
	"os"

	"github.com/lima-vm/go-qcow2reader"
	"github.com/lima-vm/lima/pkg/qemu/imgutil"
)

// inspectDiskSize attempts to inspect the disk size by itself,
// and falls back to inspectDiskSizeWithQemuImg on an error.
func inspectDiskSize(fName string) (int64, error) {
	f, err := os.Open(fName)
	if err != nil {
		return inspectDiskSizeWithQemuImg(fName)
	}
	defer f.Close()
	img, err := qcow2reader.Open(f)
	if err != nil {
		return inspectDiskSizeWithQemuImg(fName)
	}
	sz := img.Size()
	if sz < 0 {
		return inspectDiskSizeWithQemuImg(fName)
	}
	return sz, nil
}

// inspectDiskSizeWithQemuImg invokes `qemu-img` binary to inspect the disk size.
func inspectDiskSizeWithQemuImg(fName string) (int64, error) {
	info, err := imgutil.GetInfo(fName)
	if err != nil {
		return -1, err
	}
	return info.VSize, nil
}
