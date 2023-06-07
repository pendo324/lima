package wsl

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/lima-vm/lima/pkg/driver"
	"github.com/lima-vm/lima/pkg/fileutils"
	"github.com/lima-vm/lima/pkg/store/filenames"
)

// EnsureFs downloads the root fs.
func EnsureFs(driver *driver.BaseDriver) error {
	rootFsArchive := filepath.Join(driver.Instance.Dir, filenames.WslRootFs)
	if _, err := os.Stat(rootFsArchive); errors.Is(err, os.ErrNotExist) {
		var ensuredBaseDisk bool
		errs := make([]error, len(driver.Yaml.Images))
		for i, f := range driver.Yaml.Images {
			if _, err := fileutils.DownloadFile(rootFsArchive, f.File, true, "the image", *driver.Yaml.Arch); err != nil {
				errs[i] = err
				continue
			}
			ensuredBaseDisk = true
			break
		}
		if !ensuredBaseDisk {
			return fileutils.Errors(errs)
		}
	}

	return nil
}
