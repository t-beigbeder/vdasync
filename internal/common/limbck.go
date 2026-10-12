package common

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
)

const BckExtFormat = "20060102-150405"

func listBck(dss dssa.Dssa, path_ string) ([]string, error) {
	des, err := dss.List(path.Dir(path_))
	if err != nil {
		return nil, err
	}
	fn := path.Base(path_)
	pfx := fmt.Sprintf("%s.", fn)
	names := []string{}
	for _, de := range des {
		name := path.Base(de.Path)
		if len(name) != len(fn)+len(BckExtFormat)+1 || !strings.HasPrefix(name, pfx) {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}

func LimitedBackup(dss dssa.Dssa, path_ string, limit int) error {
	de, err := dss.Stat(path_)
	if err != nil {
		return err
	}
	tss := time.Unix(de.Mtime, 0).UTC().Format(BckExtFormat)
	newPath := fmt.Sprintf("%s.%s", path_, tss)
	if err = CopyEntry(dss, path_, newPath); err != nil {
		return err
	}
	if limit <= 0 {
		return nil
	}
	for {
		ns, err := listBck(dss, path_)
		if err != nil {
			return err
		}
		if len(ns) <= limit {
			return nil
		}
		rp := path.Join(path.Dir(path_), ns[len(ns)-limit-1])
		if err = dss.Rm(rp); err != nil {
			return err
		}
	}
}
