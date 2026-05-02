package common

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
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

func FileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%016x", h.Sum(nil)), nil
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

func GetFileStat(path string) (os.FileInfo, [2]int, [3]dssa.Rights, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, [2]int{}, [3]dssa.Rights{}, err
	}
	ugIds, ugoRights := GetAccessRights(fi)
	return fi, ugIds, ugoRights, nil
}
