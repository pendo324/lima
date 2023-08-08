package downloader

func Decompressor(ext string) ([]string, bool) {
	var program string
	switch ext {
	case ".gz":
		program = "7z"
	case ".bz2":
		program = "7z"
	case ".xz":
		program = "7z"
	case ".zst":
		program = "7z"
	default:
		return nil, false
	}

	return []string{program, "x"}, true
}
