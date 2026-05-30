package common

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/t-beigbeder/vdasync/dssa"
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

func ReaderSha256(rdr io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, rdr); err != nil {
		return "", err
	}

	return fmt.Sprintf("%064x", h.Sum(nil)), nil
}

func FileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	return ReaderSha256(f)
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

func Rights2Mod(ugoRights [3]dssa.Rights) (mode os.FileMode) {
	ur, gr, or := ugoRights[0], ugoRights[1], ugoRights[2]
	if ur.Read {
		mode |= 1 << 8
	}
	if ur.Write {
		mode |= 1 << 7
	}
	if ur.Execute {
		mode |= 1 << 6
	}
	if gr.Read {
		mode |= 1 << 5
	}
	if gr.Write {
		mode |= 1 << 4
	}
	if gr.Execute {
		mode |= 1 << 3
	}
	if or.Read {
		mode |= 1 << 2
	}
	if or.Write {
		mode |= 1 << 1
	}
	if or.Execute {
		mode |= 1
	}
	return
}

func Mod2Rights(perm os.FileMode) [3]dssa.Rights {
	return [3]dssa.Rights{
		{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0},
		{Read: perm&(1<<5) != 0, Write: perm&(1<<4) != 0, Execute: perm&(1<<3) != 0},
		{Read: perm&(1<<2) != 0, Write: perm&(1<<1) != 0, Execute: perm&(1) != 0},
	}
}
