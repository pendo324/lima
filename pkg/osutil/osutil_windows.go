package osutil

import (
	"fmt"
	"io/fs"
	"syscall"
)

// UnixPathMax is the value of UNIX_PATH_MAX.
const UnixPathMax = 108

// Stat is a selection of syscall.Stat_t
type Stat struct {
	Uid uint32 //nolint:revive
	Gid uint32 //nolint:revive
}

func SysStat(fi fs.FileInfo) (Stat, bool) {
	return Stat{Uid: 0, Gid: 0}, false
}

// SigInt is the value of SIGINT.
const SigInt = Signal(2)

// SigKill is the value of SIGKILL.
const SigKill = Signal(9)

type Signal int

func SysKill(pid int, sig Signal) error {
	return windows.GenerateConsoleCtrlEvent(syscall.CTRL_BREAK_EVENT, uint32(p.Pid))
}

func Ftruncate(fd int, length int64) (err error) {
	return fmt.Errorf("unimplemented")
}
