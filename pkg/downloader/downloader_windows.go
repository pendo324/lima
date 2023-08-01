//go:build windows
// +build windows

package downloader

func Decompressor(ext string) ([]string, bool) {
	return []string{"7z", "x"}, true
}
