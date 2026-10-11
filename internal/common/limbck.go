package common

import (
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

const bckExtFormat = "20060102-150405"

func listBck(path_ string) ([]string, error) {
	des, err := os.ReadDir(path.Dir(path_))
	if err != nil {
		return nil, err
	}
	fn := path.Base(path_)
	pfx := fmt.Sprintf("%s.", fn)
	names := []string{}
	for _, de := range des {
		name := de.Name()
		if len(name) != len(fn)+len(bckExtFormat)+1 || !strings.HasPrefix(name, pfx) {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}

// TODO: implement with DSS and test in localfiles/common_test
func LimitedBackup(path_ string, limit int) error {
	fi, err := os.Stat(path_)
	if err != nil {
		return err
	}
	tss := time.Unix(fi.ModTime().Unix(), 0).Format(bckExtFormat)
	rc, err := os.Open(path_)
	if err != nil {
		return err
	}
	defer rc.Close()
	newPath := fmt.Sprintf("%s.%s", path_, tss)
	wc, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer wc.Close()
	if _, err = io.Copy(wc, rc); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	if limit <= 0 {
		return nil
	}
	for {
		ns, err := listBck(path_)
		if err != nil {
			return err
		}
		if len(ns) <= limit {
			return nil
		}
		rp := path.Join(path.Dir(path_), ns[len(ns)-limit-1])
		if err = os.Remove(rp); err != nil {
			return err
		}
	}
}
