//go:build windows
// +build windows

package httpclientutil

import (
	"context"
	"fmt"
	"net"
	"net/http"

	winio "github.com/Microsoft/go-winio"
	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/lima-vm/lima/pkg/windows"
)

// // NewHTTPClientWithVSockPort creates a client.
// port is the port to use for the vsock.
func NewHTTPClientWithVSockPort(instanceName string, port int) (*http.Client, error) {
	VMIDStr, err := windows.GetInstanceVMID(fmt.Sprintf("lima-%s", instanceName))
	if err != nil {
		return nil, err
	}
	VMIDGUID, err := guid.FromString(VMIDStr)
	if err != nil {
		return nil, err
	}

	serviceGUID, err := guid.FromString(fmt.Sprintf("%x%s", port, windows.MagicVSOCKSuffix))
	if err != nil {
		return nil, err
	}

	sockAddr := &winio.HvsockAddr{
		VMID:      VMIDGUID,
		ServiceID: serviceGUID,
	}

	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.Dial(ctx, sockAddr)
			},
		},
	}
	return hc, nil
}
