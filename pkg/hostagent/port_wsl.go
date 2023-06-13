package hostagent

import (
	"context"
	"fmt"
	"net"
	"os/exec"

	"github.com/lima-vm/lima/pkg/ioutilx"
	"github.com/lima-vm/sshocker/pkg/ssh"
)

func runNetshWithCtx(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "netsh")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			if ee, ok := err.(*exec.ExitError); ok {
				if ee.Exited() && ee.ExitCode() != 0 {
					return "", fmt.Errorf("error calling netsh: %w", err)
				}
			}
		}

		return "", fmt.Errorf("error calling netsh: %w", err)
	}

	outString, err := ioutilx.FromUTF16leToString(out)
	if err != nil {
		return "", fmt.Errorf("failed to convert output from UTF16 when running wsl command netsh.exe %v, err: %w", args, err)
	}
	return outString, nil
}

func forwardTCPWsl(ctx context.Context, _ *ssh.SSHConfig, _ int, local, remote string, verb string) error {
	commonOpts := []string{"interface", "portproxy"}

	listenAddress, listenPort, err := net.SplitHostPort(local)
	if err != nil {
		return err
	}
	connectaddress, connectport, err := net.SplitHostPort(local)
	if err != nil {
		return err
	}

	switch verb {
	case verbCancel:
		{
			cancelOpts := append(commonOpts, []string{"delete", "v4tov4", "listenport=", listenPort, "listenaddress=", listenAddress}...)
			_, err := runNetshWithCtx(ctx, cancelOpts...)
			return err
		}
	case verbForward:
		{
			forwardOpts := append(commonOpts, []string{"add", "v4tov4", "listenport=", listenPort, "connectaddress=", connectaddress, "connectport=", connectport, "listenaddress=", listenAddress}...)
			_, err := runNetshWithCtx(ctx, forwardOpts...)
			return err
		}
	}

	return fmt.Errorf("unhandled forwarding type")
}
