package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lima-vm/lima/pkg/textutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newInstallSystemdCommand() *cobra.Command {
	var installSystemdCommand = &cobra.Command{
		Use:   "install-systemd",
		Short: "install a systemd unit (user)",
		RunE:  installSystemdAction,
	}
	installSystemdCommand.Flags().Int("tcp-port", 0, "use tcp server on specified port")
	installSystemdCommand.Flags().Int("vsock-port", 0, "use vsock server on specified port")
	installSystemdCommand.MarkFlagsMutuallyExclusive("tcp-port", "vsock-port")
	return installSystemdCommand
}

func installSystemdAction(cmd *cobra.Command, args []string) error {
	tcp, err := cmd.Flags().GetInt("tcp-port")
	if err != nil {
		return err
	}
	vsock, err := cmd.Flags().GetInt("vsock-portock")
	if err != nil {
		return err
	}
	unit, err := generateSystemdUnit(tcp, vsock)
	if err != nil {
		return err
	}
	unitPath := "/etc/systemd/system/lima-guestagent.service"
	if _, err := os.Stat(unitPath); !errors.Is(err, os.ErrNotExist) {
		logrus.Infof("File %q already exists, overwriting", unitPath)
	} else {
		unitDir := filepath.Dir(unitPath)
		if err := os.MkdirAll(unitDir, 0755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(unitPath, unit, 0644); err != nil {
		return err
	}
	logrus.Infof("Written file %q", unitPath)
	argss := [][]string{
		{"daemon-reload"},
		{"enable", "--now", "lima-guestagent.service"},
	}
	for _, args := range argss {
		cmd := exec.Command("systemctl", append([]string{"--system"}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		logrus.Infof("Executing: %s", strings.Join(cmd.Args, " "))
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	logrus.Info("Done")
	return nil
}

//go:embed lima-guestagent.TEMPLATE.service
var systemdUnitTemplate string

func generateSystemdUnit(tcpPort, vsockPort int) ([]byte, error) {
	selfExeAbs, err := os.Executable()
	if err != nil {
		return nil, err
	}

	var args []string
	if tcpPort != 0 {
		args = append(args, fmt.Sprintf("--tcp-port %d", tcpPort))
	}
	if vsockPort != 0 {
		args = append(args, fmt.Sprintf("--vsock-port %d", vsockPort))
	}

	m := map[string]string{
		"Binary": selfExeAbs,
		"Args":   strings.Join(args, " "),
	}
	return textutil.ExecuteTemplate(systemdUnitTemplate, m)
}
