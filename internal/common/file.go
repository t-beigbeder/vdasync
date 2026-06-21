package common

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/t-beigbeder/vdasync/dssa"
)

func FileExists(path_ string) bool {
	_, err := os.Stat(path_)
	return err == nil
}

func FileSize(path_ string) (int64, error) {
	fi, err := os.Stat(path_)
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

func FileSha256(path_ string) (string, error) {
	f, err := os.Open(path_)
	if err != nil {
		return "", err
	}
	return ReaderSha256(f)
}

func WriteFile(path_ string, data []byte) error {
	file, err := os.OpenFile(path_, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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

func LoadFile(path_ string) ([]byte, error) {
	sz, err := FileSize(path_)
	if err != nil {
		return nil, err
	}
	if sz > MaxLoadFileSize {
		return nil, fmt.Errorf("file size is %d bytes > %d", sz, MaxLoadFileSize)
	}
	file, err := os.Open(path_)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func FileLines(path_ string) ([]string, error) {
	file, err := os.Open(path_)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func Lines2file(lines []string, path_ string) error {
	return WriteFile(path_, []byte(strings.Join(lines, "\n")))
}

func GetFileStat(path_ string) (os.FileInfo, [2]int, [3]dssa.Rights, error) {
	fi, err := os.Lstat(path_)
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

func Perm2Rights(perm os.FileMode) [3]dssa.Rights {
	return [3]dssa.Rights{
		{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0},
		{Read: perm&(1<<5) != 0, Write: perm&(1<<4) != 0, Execute: perm&(1<<3) != 0},
		{Read: perm&(1<<2) != 0, Write: perm&(1<<1) != 0, Execute: perm&(1) != 0},
	}
}
