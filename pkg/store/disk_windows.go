//go:build windows
// +build windows

package store

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/docker/go-units"
)

// inspectDiskSize parses the output of diskpart to get size of disk fName.
//
// Disks must have a file extension for diskpart to work.
// TODO: Add proper parsing support to https://github.com/lima-vm/go-qcow2reader/blob/master/image/vhdx/vhdx.go
func inspectDiskSize(fName string) (int64, error) {
	script := fmt.Sprintf(`@"
select vdisk file=%s
detail vdisk
"@ | diskpart`, fName)

	out, err := exec.Command("powershell.exe",
		"-nologo",
		"-noprofile",
		script,
	).CombinedOutput()
	if err != nil {
		return 0, err
	}

	var sizeNum string
	var sizeUnit string
	// Example output (whitespace preserved)
	// Device type ID: 3 (Unknown)
	// Vendor ID: {EC984AEC-A0F9-47E9-901F-71415A66345B} (Microsoft Corporation)
	// State: Added
	// Virtual size:   10 MB
	// Physical size: 4096 KB
	// Filename: C:\path\to\test.vhdx
	// Is Child: No
	// Parent Filename:
	// Associated disk#: Not found.
	re := regexp.MustCompile(`Virtual size:\s*(?P<num>\d+)\s(?P<unit>[A-Za-z]+)`)
	if matches := re.FindStringSubmatch(string(out)); matches != nil {
		sizeNum = matches[re.SubexpIndex("num")]
		sizeUnit = matches[re.SubexpIndex("unit")]
	} else {
		return 0, fmt.Errorf("failed to parse size from diskpart output for disk at %q", fName)
	}

	size := sizeNum + sizeUnit
	sizeB, err := units.RAMInBytes(size)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size %q into bytes: %w", size, err)
	}

	return sizeB, nil
}
