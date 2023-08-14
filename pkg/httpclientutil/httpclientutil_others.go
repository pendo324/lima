//go:build !windows
// +build !windows

package httpclientutil

import (
	"net"
	"net/http"

	"github.com/mdlayher/vsock"
)

// // NewHTTPClientWithVSockPort creates a client.
// port is the port to use for the vsock.
func NewHTTPClientWithVSockPort(port int) *http.Client {
	hc := &http.Client{
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return vsock.Dial(2, uint32(port), nil)
			},
		},
	}
	return hc
}
