package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
)

func InventoryCsvExport(rootPath string, csvPath string, algos string) error {
	if algos == "" {
		algos = "sha256"
	}
	wrw, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer wrw.Close()
	cw := csv.NewWriter(wrw)
	if err = cw.Write(slices.Concat([]string{"relPath", "isDir", "size", "mTime", "isSymLink", "symLinkTarget"}, strings.Split(algos, ","))); err != nil {
		return err
	}

	err = filepath.Walk(rootPath, func(path_ string, info fs.FileInfo, err error) error {
		rp := common.RelPath(path_, rootPath)
		if info.IsDir() {
			if err = cw.Write([]string{rp, "1", "", dispMtime(info.ModTime())}); err != nil {
				return err
			}
			return nil
		}
		rdr, err := os.Open(path_)
		if err != nil {
			return err
		}
		css, err := common.ReaderChecksum(rdr, algos)
		if err != nil {
			return err
		}
		csvLine := make([]string, 6+len(strings.Split(css, ",")))
		csvLine[0] = rp
		csvLine[1] = "0"
		csvLine[2] = fmt.Sprintf("%d", info.Size())
		csvLine[3] = dispMtime(info.ModTime())
		csvLine[4] = "0"
		csvLine[5] = ""
		for i, cs := range strings.Split(css, ",") {
			alCs := strings.Split(cs, ":")
			if len(alCs) != 2 {
				return fmt.Errorf("invalid algo/checksum %s", algos)
			}
			csvLine[i+6] = alCs[1]
		}
		if err = cw.Write(csvLine); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	cw.Flush()
	if err = wrw.Close(); err != nil {
		return err
	}
	return nil
}

func dispMtime(ts time.Time) string {
	return ts.UTC().Format(time.RFC3339)
}
