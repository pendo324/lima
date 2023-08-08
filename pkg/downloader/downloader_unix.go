//go:build !windows
// +build !windows

package downloader

func Decompressor(ext string) ([]string, bool) {
	var program string
	switch ext {
	case ".gz":
		program = "gzip"
	case ".bz2":
		program = "bzip2"
	case ".xz":
		program = "xz"
	case ".zst":
		program = "zstd"
	default:
		return nil, false
	}
	// -d --decompress
	return []string{program, "-d"}, true
}
