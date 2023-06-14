package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/docker/go-units"
	hostagentclient "github.com/lima-vm/lima/pkg/hostagent/api/client"
	"github.com/lima-vm/lima/pkg/ioutilx"
	"github.com/lima-vm/lima/pkg/limayaml"
	"github.com/lima-vm/lima/pkg/store/dirnames"
	"github.com/lima-vm/lima/pkg/store/filenames"
	"github.com/lima-vm/lima/pkg/textutil"
	"github.com/sirupsen/logrus"
)

type Status = string

const (
	StatusUnknown       Status = ""
	StatusUnititialized Status = "Unititialized"
	StatusInstalling    Status = "Installing"
	StatusBroken        Status = "Broken"
	StatusStopped       Status = "Stopped"
	StatusRunning       Status = "Running"
)

type Instance struct {
	Name            string             `json:"name"`
	Status          Status             `json:"status"`
	Dir             string             `json:"dir"`
	VMType          limayaml.VMType    `json:"vmType"`
	Arch            limayaml.Arch      `json:"arch"`
	CPUType         string             `json:"cpuType"`
	CPUs            int                `json:"cpus,omitempty"`
	Memory          int64              `json:"memory,omitempty"` // bytes
	Disk            int64              `json:"disk,omitempty"`   // bytes
	Message         string             `json:"message,omitempty"`
	AdditionalDisks []limayaml.Disk    `json:"additionalDisks,omitempty"`
	Networks        []limayaml.Network `json:"network,omitempty"`
	SSHLocalPort    int                `json:"sshLocalPort,omitempty"`
	SSHConfigFile   string             `json:"sshConfigFile,omitempty"`
	HostAgentPID    int                `json:"hostAgentPID,omitempty"`
	DriverPID       int                `json:"driverPID,omitempty"`
	Errors          []error            `json:"errors,omitempty"`
	Config          *limayaml.LimaYAML `json:"config,omitempty"`
	RootFsPath      string             `json:"rootfs,omitempty"`
	DistroName      string             `json:"distroName,omitempty"`
	SSHAddress      string             `json:"sshAddress,omitempty"`
}

func (inst *Instance) LoadYAML() (*limayaml.LimaYAML, error) {
	if inst.Dir == "" {
		return nil, errors.New("inst.Dir is empty")
	}
	yamlPath := filepath.Join(inst.Dir, filenames.LimaYAML)
	return LoadYAMLByFilePath(yamlPath)
}

// Inspect returns err only when the instance does not exist (os.ErrNotExist).
// Other errors are returned as *Instance.Errors
func Inspect(instName string) (*Instance, error) {
	inst := &Instance{
		Name:   instName,
		Status: StatusUnknown,
	}
	// InstanceDir validates the instName but does not check whether the instance exists
	instDir, err := InstanceDir(instName)
	if err != nil {
		return nil, err
	}
	// Make sure inst.Dir is set, even when YAML validation fails
	inst.Dir = instDir
	yamlPath := filepath.Join(instDir, filenames.LimaYAML)
	y, err := LoadYAMLByFilePath(yamlPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		inst.Errors = append(inst.Errors, err)
		return inst, nil
	}
	inst.Config = y
	inst.Arch = *y.Arch
	inst.VMType = *y.VMType
	inst.CPUType = y.CPUType[*y.Arch]
	inst.SSHAddress = "127.0.0.1"
	inst.SSHLocalPort = *y.SSH.LocalPort // maybe 0
	inst.SSHConfigFile = filepath.Join(instDir, filenames.SSHConfig)
	inst.HostAgentPID, err = ReadPIDFile(filepath.Join(instDir, filenames.HostAgentPID))
	if err != nil {
		inst.Status = StatusBroken
		inst.Errors = append(inst.Errors, err)
	}

	if inst.HostAgentPID != 0 {
		haSock := filepath.Join(instDir, filenames.HostAgentSock)
		haClient, err := hostagentclient.NewHostAgentClient(haSock)
		if err != nil {
			inst.Status = StatusBroken
			inst.Errors = append(inst.Errors, fmt.Errorf("failed to connect to %q: %w", haSock, err))
		} else {
			ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
			defer cancel()
			info, err := haClient.Info(ctx)
			if err != nil {
				inst.Status = StatusBroken
				inst.Errors = append(inst.Errors, fmt.Errorf("failed to get Info from %q: %w", haSock, err))
			} else {
				inst.SSHLocalPort = info.SSHLocalPort
			}
		}
	}

	if inst.VMType == limayaml.WSL {
		inst.DistroName = fmt.Sprintf("%s-%s", "lima", inst.Name)
		status, err := GetWslStatus(instName, inst.DistroName)
		if err != nil {
			inst.Status = StatusBroken
			inst.Errors = append(inst.Errors, err)
		} else {
			inst.Status = status
		}

		if inst.Status == StatusStopped || inst.Status == StatusRunning {
			sshAddr, err := GetSSHAddress(instName, inst.DistroName)
			if err == nil {
				inst.SSHAddress = sshAddr
			} else {
				inst.Errors = append(inst.Errors, err)
			}
		}
	} else {
		inst.CPUs = *y.CPUs
		memory, err := units.RAMInBytes(*y.Memory)
		if err == nil {
			inst.Memory = memory
		}
		disk, err := units.RAMInBytes(*y.Disk)
		if err == nil {
			inst.Disk = disk
		}
		inst.AdditionalDisks = y.AdditionalDisks
		inst.Networks = y.Networks

		inst.DriverPID, err = ReadPIDFile(filepath.Join(instDir, filenames.PIDFile(*y.VMType)))
		if err != nil {
			inst.Status = StatusBroken
			inst.Errors = append(inst.Errors, err)
		}

		if inst.Status == StatusUnknown {
			if inst.HostAgentPID > 0 && inst.DriverPID > 0 {
				inst.Status = StatusRunning
			} else if inst.HostAgentPID == 0 && inst.DriverPID == 0 {
				inst.Status = StatusStopped
			} else if inst.HostAgentPID > 0 && inst.DriverPID == 0 {
				inst.Errors = append(inst.Errors, errors.New("host agent is running but driver is not"))
				inst.Status = StatusBroken
			} else {
				inst.Errors = append(inst.Errors, fmt.Errorf("%s driver is running but host agent is not", inst.VMType))
				inst.Status = StatusBroken
			}
		}
	}
	tmpl, err := template.New("format").Parse(y.Message)
	if err != nil {
		inst.Errors = append(inst.Errors, fmt.Errorf("message %q is not a valid template: %w", y.Message, err))
		inst.Status = StatusBroken
	} else {
		data, err := AddGlobalFields(inst)
		if err != nil {
			inst.Errors = append(inst.Errors, fmt.Errorf("cannot add global fields to instance data: %w", err))
			inst.Status = StatusBroken
		} else {
			var message strings.Builder
			err = tmpl.Execute(&message, data)
			if err != nil {
				inst.Errors = append(inst.Errors, fmt.Errorf("cannot execute template %q: %w", y.Message, err))
				inst.Status = StatusBroken
			} else {
				inst.Message = message.String()
			}
		}
	}
	return inst, nil
}

// ReadPIDFile returns 0 if the PID file does not exist or the process has already terminated
// (in which case the PID file will be removed).
func ReadPIDFile(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			_ = os.Remove(path)
			return 0, nil
		}
		// We may not have permission to send the signal (e.g. to network daemon running as root).
		// But if we get a permissions error, it means the process is still running.
		if !errors.Is(err, os.ErrPermission) {
			return 0, err
		}
	}
	return pid, nil
}

type FormatData struct {
	Instance
	HostOS       string
	HostArch     string
	LimaHome     string
	IdentityFile string
}

var FormatHelp = "\n" +
	"These functions are available to go templates:\n\n" +
	textutil.IndentString(2,
		strings.Join(textutil.FuncHelp, "\n")+"\n")

func AddGlobalFields(inst *Instance) (FormatData, error) {
	var data FormatData
	data.Instance = *inst
	// Add HostOS
	data.HostOS = runtime.GOOS
	// Add HostArch
	data.HostArch = limayaml.NewArch(runtime.GOARCH)
	// Add IdentityFile
	configDir, err := dirnames.LimaConfigDir()
	if err != nil {
		return FormatData{}, err
	}
	data.IdentityFile = filepath.Join(configDir, filenames.UserPrivateKey)
	// Add LimaHome
	data.LimaHome, err = dirnames.LimaDir()
	if err != nil {
		return FormatData{}, err
	}
	return data, nil
}

type PrintOptions struct {
	AllFields     bool
	TerminalWidth int
}

// PrintInstances prints instances in a requested format to a given io.Writer.
// Supported formats are "json", "yaml", "table", or a go template
func PrintInstances(w io.Writer, instances []*Instance, format string, options *PrintOptions) error {
	switch format {
	case "json":
		format = "{{json .}}"
	case "yaml":
		format = "{{yaml .}}"
	case "table":
		types := map[string]int{}
		archs := map[string]int{}
		for _, instance := range instances {
			types[instance.VMType]++
			archs[instance.Arch]++
		}
		all := options != nil && options.AllFields
		width := 0
		if options != nil {
			width = options.TerminalWidth
		}
		columnWidth := 8
		hideType := false
		hideArch := false
		hideDir := false

		columns := 1 // NAME
		columns += 2 // STATUS
		columns += 2 // SSH
		// can we still fit the remaining columns (7)
		if width == 0 || (columns+7)*columnWidth > width && !all {
			hideType = len(types) == 1
		}
		if !hideType {
			columns++ // VMTYPE
		}
		// only hide arch if it is the same as the host arch
		goarch := limayaml.NewArch(runtime.GOARCH)
		// can we still fit the remaining columns (6)
		if width == 0 || (columns+6)*columnWidth > width && !all {
			hideArch = len(archs) == 1 && instances[0].Arch == goarch
		}
		if !hideArch {
			columns++ // ARCH
		}
		columns++ // CPUS
		columns++ // MEMORY
		columns++ // DISK
		// can we still fit the remaining columns (2)
		if width != 0 && (columns+2)*columnWidth > width && !all {
			hideDir = true
		}
		if !hideDir {
			columns += 2 // DIR
		}
		_ = columns

		w := tabwriter.NewWriter(w, 4, 8, 4, ' ', 0)
		fmt.Fprint(w, "NAME\tSTATUS\tSSH")
		if !hideType {
			fmt.Fprint(w, "\tVMTYPE")
		}
		if !hideArch {
			fmt.Fprint(w, "\tARCH")
		}
		fmt.Fprint(w, "\tCPUS\tMEMORY\tDISK")
		if !hideDir {
			fmt.Fprint(w, "\tDIR")
		}
		fmt.Fprintln(w)

		u, err := user.Current()
		if err != nil {
			return err
		}
		homeDir := u.HomeDir

		for _, instance := range instances {
			dir := instance.Dir
			if strings.HasPrefix(dir, homeDir) {
				dir = strings.Replace(dir, homeDir, "~", 1)
			}
			fmt.Fprintf(w, "%s\t%s\t%s",
				instance.Name,
				instance.Status,
				fmt.Sprintf("%s:%d", "127.0.0.1", instance.SSHLocalPort),
			)
			if !hideType {
				fmt.Fprintf(w, "\t%s",
					instance.VMType,
				)
			}
			if !hideArch {
				fmt.Fprintf(w, "\t%s",
					instance.Arch,
				)
			}
			fmt.Fprintf(w, "\t%d\t%s\t%s",
				instance.CPUs,
				units.BytesSize(float64(instance.Memory)),
				units.BytesSize(float64(instance.Disk)),
			)
			if !hideDir {
				fmt.Fprintf(w, "\t%s",
					dir,
				)
			}
			fmt.Fprint(w, "\n")

		}
		return w.Flush()
	default:
		// NOP
	}
	tmpl, err := template.New("format").Funcs(textutil.TemplateFuncMap).Parse(format)
	if err != nil {
		return fmt.Errorf("invalid go template: %w", err)
	}
	for _, instance := range instances {
		data, err := AddGlobalFields(instance)
		if err != nil {
			return err
		}
		data.Message = strings.TrimSuffix(instance.Message, "\n")
		err = tmpl.Execute(w, data)
		if err != nil {
			return err
		}
		fmt.Fprintln(w)
	}
	return nil
}

func GetWslStatus(instName, distroName string) (string, error) {
	// Expected output (whitespace preserved):
	// PS > wsl --list --verbose
	//   NAME      STATE           VERSION
	// * Ubuntu    Stopped         2
	cmd := exec.Command("wsl.exe", "--list", "--verbose")
	out, err := cmd.CombinedOutput()
	outString, outUTFErr := ioutilx.FromUTF16leToString(bytes.NewReader(out))
	if err != nil {
		logrus.Debugf("outString: %s", outString)
		return "", fmt.Errorf("failed to read instance state for instance %s, try running `wsl --list --verbose` to debug, err: %w", instName, err)
	}

	if outUTFErr != nil {
		return "", fmt.Errorf("failed to convert output from UTF16 for instance state for instance %s, err: %w", instName, err)
	}

	if len(outString) == 0 {
		return StatusBroken, fmt.Errorf("failed to read instance state for instance %s, try running `wsl --list --verbose` to debug, err: %w", instName, err)
	}

	var instState string
	// wsl --list --verbose may have differernt headers depending on localization, just split by line
	// Windows uses little endian by default, use unicode.UseBOM policy to retrieve BOM from the text,
	// and unicode.LittleEndian as a fallback
	for _, rows := range strings.Split(strings.ReplaceAll(string(outString), "\r\n", "\n"), "\n") {
		cols := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(rows), -1)
		nameIdx := 0
		// '*' indicates default instance
		if cols[0] == "*" {
			nameIdx = 1
		}
		if cols[nameIdx] == distroName {
			instState = cols[nameIdx+1]
			break
		}
	}

	if instState == "" {
		return StatusUnititialized, nil
	}

	return instState, nil
}

func GetSSHAddress(instName, distroName string) (string, error) {
	// Expected output (whitespace preserved, [] for optional):
	// PS > wsl -d <distroName> bash -c hostname -I | cut -d' ' -f1
	// 168.1.1.1 [10.0.0.1]
	cmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", `hostname -I | cut -d ' ' -f1`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Debugf("outString: %s", out)
		return "", fmt.Errorf("failed to get hostname for instance %s, err: %w", instName, err)
	}

	return strings.TrimSpace(string(out)), nil
}
