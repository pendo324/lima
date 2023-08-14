//go:build windows
// +build windows

package windows

import (
	"encoding/json"
	"fmt"

	"github.com/lima-vm/lima/pkg/executil"
)

type CommandLineJSON []struct {
	CommandLine string
}

// GetProcessCommandLine returns a slice of string containing all commandlines for a given process name.
func GetProcessCommandLine(name string) ([]string, error) {
	out, err := executil.RunUTF16leCommand([]string{
		"powershell.exe",
		"-nologo",
		"-noprofile",
		fmt.Sprintf(
			`Get-CimInstance Win32_Process -Filter "name = '%s'" | Select CommandLine | ConvertTo-Json`,
			name,
		),
	})

	if err != nil {
		return nil, err
	}

	var outJSON CommandLineJSON
	json.Unmarshal([]byte(out), &outJSON)

	var ret []string
	for _, s := range outJSON {
		ret = append(ret, s.CommandLine)
	}

	return ret, nil
}
