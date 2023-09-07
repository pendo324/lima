// Package filenames defines the names of the files that appear under an instance dir
// or inside the config directory.
//
// See docs/internal.md .
package filenames

// Instance names starting with an underscore are reserved for lima internal usage

const (
	ConfigDir   = "_config"
	CacheDir    = "_cache"    // not yet implemented
	NetworksDir = "_networks" // network log files are stored here
	DisksDir    = "_disks"    // disks are stored here
)

// Filenames used inside the ConfigDir

const (
	UserPrivateKey = "user"
	UserPublicKey  = UserPrivateKey + ".pub"
	NetworksConfig = "networks.yaml"
	Default        = "default.yaml"
	Override       = "override.yaml"
)

// Filenames that may appear under an instance directory

const (
	LimaYAML           = "lima.yaml"
	CIDataISO          = "cidata.iso"
	CIDataISODir       = "cidata"
	BaseDisk           = "basedisk"
	DiffDisk           = "diffdisk"
	Kernel             = "kernel"
	KernelCmdline      = "kernel.cmdline"
	Initrd             = "initrd"
	QMPSock            = "qmp.sock"
	SerialLog          = "serial.log" // default serial (ttyS0, but ttyAMA0 on qemu-system-{arm,aarch64})
	SerialSock         = "serial.sock"
	SerialPCILog       = "serialp.log" // pci serial (ttyS0 on qemu-system-{arm,aarch64})
	SerialPCISock      = "serialp.sock"
	SerialVirtioLog    = "serialv.log" // virtio serial
	SerialVirtioSock   = "serialv.sock"
	SSHSock            = "ssh.sock"
	SSHConfig          = "ssh.config"
	VhostSock          = "virtiofsd-%d.sock"
	VNCDisplayFile     = "vncdisplay"
	VNCPasswordFile    = "vncpassword"
	GuestAgentSock     = "ga.sock"
	HostAgentPID       = "ha.pid"
	HostAgentSock      = "ha.sock"
	HostAgentStdoutLog = "ha.stdout.log"
	HostAgentStderrLog = "ha.stderr.log"
	VzIdentifier       = "vz-identifier"
	VzEfi              = "vz-efi"

	// SocketDir is the default location for forwarded sockets with a relative paths in HostSocket
	SocketDir = "sock"
)

// Filenames used under a disk directory

const (
	DataDisk        = "datadisk"
	DataDiskWindows = "datadisk.vhdx"
	InUseBy         = "in_use_by"
)

// LongestSock is the longest socket name.
// On macOS, the full path of the socket (excluding the NUL terminator) must be less than 104 characters.
// See unix(4).
//
// On Linux, the full path must be less than 108 characters.
//
// ssh appends 16 bytes of random characters when it first creates the socket:
// https://github.com/openssh/openssh-portable/blob/V_8_7_P1/mux.c#L1271-L1285
const LongestSock = SSHSock + ".1234567890123456"

func PIDFile(name string) string {
	return name + ".pid"
}
