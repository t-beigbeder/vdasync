package common

import (
	"fmt"
	"io"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func WriteFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

const MaxLoadFileSize = 65536

func LoadFile(path string) ([]byte, error) {
	sz, err := FileSize(path)
	if err != nil {
		return nil, err
	}
	if sz > MaxLoadFileSize {
		return nil, fmt.Errorf("file size is %d bytes > %d", sz, MaxLoadFileSize)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
