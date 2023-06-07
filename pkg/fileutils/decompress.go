package fileutils

import (
	"bufio"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
)

func DecompressGz(compressedPath, outPath string) error {
	f, err := os.Open(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	gzipStream := bufio.NewReader(f)
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}
	outF, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return decompressToFile(uncompressedStream, outF)
}

func DecompressZlib(compressedPath, outPath string) error {
	f, err := os.Open(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	gzipStream := bufio.NewReader(f)
	uncompressedStream, err := zlib.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}
	outF, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return decompressToFile(uncompressedStream, outF)
}

func decompressToFile(src io.ReadCloser, dest io.WriteCloser) error {
	_, err := io.Copy(dest, src)
	if err != nil {
		retErr := fmt.Errorf("failed to read compressed response: %w", err)
		closeErr := src.Close()
		if closeErr != nil {
			retErr = fmt.Errorf("%w. failed to close src: %w", retErr, closeErr)
		}
		return retErr
	}

	if err = src.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	if err = dest.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
