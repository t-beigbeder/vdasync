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

	err = filepath.Walk(rootPath, func(path_ string, _ fs.FileInfo, err error) error {
		rp := common.RelPath(path_, rootPath)
		isSymlink := "0"
		linkTarget := ""
		css := ""
		cssMap := map[string]string{}
		info, err := os.Lstat(path_)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err = cw.Write([]string{rp, "1", "", expDispMtime(info.ModTime()), "0", ""}); err != nil {
				return err
			}
			return nil
		}
		if info.Mode().Type()&fs.ModeSymlink != 0 {
			linkTarget, err = os.Readlink(path_)
			if err != nil {
				return err
			}
			isSymlink = "1"
		} else {
			rdr, err := os.Open(path_)
			if err != nil {
				return err
			}
			css, err = common.ReaderChecksum(rdr, algos)
			if err != nil {
				return err
			}
			cssMap = common.Css2Map(css)
		}
		csvLine := make([]string, 6+len(strings.Split(css, ",")))
		csvLine[0] = rp
		csvLine[1] = "0"
		csvLine[2] = fmt.Sprintf("%d", info.Size())
		csvLine[3] = expDispMtime(info.ModTime())
		csvLine[4] = isSymlink
		csvLine[5] = linkTarget
		if css != "" {
			for i, algo := range strings.Split(algos, ",") {
				cs4al, ok := cssMap[algo]
				if !ok {
					return fmt.Errorf("invalid algo/checksum %s/%s", algo, css)
				}
				csvLine[i+6] = cs4al
			}
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

func expDispMtime(ts time.Time) string {
	return ts.UTC().Format(time.RFC3339)
}
