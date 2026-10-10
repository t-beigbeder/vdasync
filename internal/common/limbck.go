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

func listBck(path_ string) ([]string, error) {
	des, err := os.ReadDir(path.Dir(path_))
	fn := path.Base(path_)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, de := range des {
		name := de.Name()
		if len(name) != len(fn)+16 || !strings.HasPrefix(name, fmt.Sprintf("%s.", fn)) {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}

func LimitedBackup(path_ string, limit int) error {
	fi, err := os.Stat(path_)
	if err != nil {
		return err
	}
	tss := time.Unix(fi.ModTime().Unix(), 0).Format("20060102-150405")
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
		a := true
		_ = a
		if err = os.Remove(rp); err != nil {
			return err
		}
	}
}
