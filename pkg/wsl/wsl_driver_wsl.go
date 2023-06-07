package wsl

import (
	"context"
	"fmt"
	"strings"

	"github.com/lima-vm/lima/pkg/driver"
	"github.com/lima-vm/lima/pkg/limayaml"
	"github.com/lima-vm/lima/pkg/reflectutil"
	"github.com/lima-vm/lima/pkg/store"
	"github.com/sirupsen/logrus"
)

const Enabled = true

type LimaWslDriver struct {
	*driver.BaseDriver
}

func New(driver *driver.BaseDriver) *LimaWslDriver {
	return &LimaWslDriver{
		BaseDriver: driver,
	}
}

func (l *LimaWslDriver) Validate() error {
	// TODO: add new mount type for WSL2 since this is handled by WSL2 automatically and no other mount type should be used.
	if *l.Yaml.MountType != limayaml.WSLMount {
		return fmt.Errorf("field `mountType` must be %q for WSL2 driver, got %q", limayaml.WSLMount, *l.Yaml.MountType)
	}
	if *l.Yaml.Firmware.LegacyBIOS {
		return fmt.Errorf("`firmware.legacyBIOS` configuration is not supported for WSL2 driver")
	}
	// TODO: revise this list for WSL2
	if unknown := reflectutil.UnknownNonEmptyFields(l.Yaml, "VMType",
		"Arch",
		"Images",
		"CPUs",
		"CPUType",
		"Memory",
		"Disk",
		"Mounts",
		"MountType",
		"SSH",
		"Firmware",
		"Provision",
		"Containerd",
		"Probes",
		"PortForwards",
		"Message",
		"Networks",
		"Env",
		"DNS",
		"HostResolver",
		"PropagateProxyEnv",
		"CACertificates",
		"Rosetta",
		"AdditionalDisks",
		"Audio",
		"Video",
	); len(unknown) > 0 {
		logrus.Warnf("Ignoring: vmType %s: %+v", *l.Yaml.VMType, unknown)
	}

	if !limayaml.IsNativeArch(*l.Yaml.Arch) {
		return fmt.Errorf("unsupported arch: %q", *l.Yaml.Arch)
	}

	for k, v := range l.Yaml.CPUType {
		if v != "" {
			logrus.Warnf("Ignoring: vmType %s: cpuType[%q]: %q", *l.Yaml.VMType, k, v)
		}
	}

	for i, image := range l.Yaml.Images {
		if unknown := reflectutil.UnknownNonEmptyFields(image, "File"); len(unknown) > 0 {
			logrus.Warnf("Ignoring: vmType %s: images[%d]: %+v", *l.Yaml.VMType, i, unknown)
		}
		// TODO: real filetype checks
		if image.Arch == *l.Yaml.Arch && (!strings.HasSuffix(image.Location, "tar.")) {
			return fmt.Errorf("unsupported image type for vmType: %s, tarball root file system required: %q", *l.Yaml.VMType, image.Location)
		}
	}

	for i, mount := range l.Yaml.Mounts {
		if unknown := reflectutil.UnknownNonEmptyFields(mount); len(unknown) > 0 {
			logrus.Warnf("Ignoring: vmType %s: mounts[%d]: %+v", *l.Yaml.VMType, i, unknown)
		}
	}

	for i, network := range l.Yaml.Networks {
		if unknown := reflectutil.UnknownNonEmptyFields(network); len(unknown) > 0 {
			logrus.Warnf("Ignoring: vmType %s: networks[%d]: %+v", *l.Yaml.VMType, i, unknown)
		}
	}

	audioDevice := *l.Yaml.Audio.Device
	if audioDevice != "" {
		logrus.Warnf("Ignoring: vmType %s: `audio.device`: %+v", *l.Yaml.VMType, audioDevice)
	}

	videoDisplay := *l.Yaml.Video.Display
	if videoDisplay != "" {
		logrus.Warnf("Ignoring: vmType %s: `audio.device`: %+v", *l.Yaml.VMType, videoDisplay)
	}
	return nil
}

func (l *LimaWslDriver) CreateDisk() error {
	// TODO: rewrite EnsureDisk to work with vhdx
	// if err := EnsureDisk(l.BaseDriver); err != nil {
	// 	return err
	// }

	return nil
}

func (l *LimaWslDriver) Start(ctx context.Context) (chan error, error) {
	logrus.Infof("Starting WSL VM")
	status, err := store.GetWslStatus("lima-" + l.Instance.Name)
	if err != nil {
		return nil, err
	}

	err = nil

	if status == store.StatusStopped {
		err = startVM("lima-" + l.Instance.Name)
	} else if status == store.StatusUnititialized {
		err = EnsureFs(l.BaseDriver)
		if err != nil {
			return nil, err
		}
		err = initVM("lima-"+l.Instance.Name, l.Instance.Dir)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Requires WSLG, which requires specific version of WSL2 to be installed.
// TODO: Add check.
func (l *LimaWslDriver) CanRunGUI() bool {
	// return *l.Yaml.Video.Display == "wsl"
	return false
}

func (l *LimaWslDriver) RunGUI() error {
	if l.CanRunGUI() {
		// return l.machine.StartGraphicApplication(1920, 1200)
	}
	return fmt.Errorf("RunGUI is not support for the given driver '%s' and diplay '%s'", "wsl", *l.Yaml.Video.Display)
}

func (l *LimaWslDriver) Stop(_ context.Context) error {
	logrus.Info("Shutting down WSL2 VM")

	return stopVM("lima-" + l.Instance.Name)
}
