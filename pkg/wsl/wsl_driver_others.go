// TODO: Uncomment this file once major development is complete
// //go:build !windows || no_wsl
// // +build !windows no_wsl

package wsl

// import (
// 	"context"
// 	"errors"

// 	"github.com/lima-vm/lima/pkg/driver"
// )

// var ErrUnsupported = errors.New("vm driver 'wsl' requires Windows 10 build 19041 or later (Hint: try recompiling Lima if you are seeing this error on Windows 10+)")

// const Enabled = false

// type LimaWslDriver struct {
// 	*driver.BaseDriver
// }

// func New(driver *driver.BaseDriver) *LimaWslDriver {
// 	return &LimaWslDriver{
// 		BaseDriver: driver,
// 	}
// }

// func (l *LimaWslDriver) Validate() error {
// 	return ErrUnsupported
// }

// func (l *LimaWslDriver) CreateDisk() error {
// 	return ErrUnsupported
// }

// func (l *LimaWslDriver) Start(ctx context.Context) (chan error, error) {
// 	return nil, ErrUnsupported
// }

// func (l *LimaWslDriver) Stop(_ context.Context) error {
// 	return ErrUnsupported
// }
