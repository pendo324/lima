package ioutilx

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// ReadAtMaximum reands n at maximum.
func ReadAtMaximum(r io.Reader, n int64) ([]byte, error) {
	lr := &io.LimitedReader{
		R: r,
		N: n,
	}
	b, err := io.ReadAll(lr)
	if err != nil {
		if errors.Is(err, io.EOF) && lr.N <= 0 {
			err = fmt.Errorf("exceeded the limit (%d bytes): %w", n, err)
		}
	}
	return b, err
}

func FromUTF16le(r io.Reader) io.Reader {
	o := transform.NewReader(r, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder())
	return o
}

func FromUTF16leToString(r io.Reader) (string, error) {
	out, err := io.ReadAll(FromUTF16le(r))
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func CannonicalWindowsPath(orig string) string {
	newPath := orig
	out, err := exec.Command("cygpath", "-m", orig).CombinedOutput()
	if err != nil {
		logrus.WithError(err).Errorf("failed to convert path to mingw, maybe not using Git ssh?")
	} else {
		logrus.Infof("cygpath output (for %s): %s", orig, string(out))
		newPath = strings.TrimSpace(string(out))
	}
	return newPath
}
