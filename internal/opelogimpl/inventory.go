package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func InventoryCsvImport(oplm opelog.OpeLogManager, csvPath string, algos string) error {
	cf, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer cf.Close()
	csr := csv.NewReader(cf)
	csr.FieldsPerRecord = -1
	if algos == "" {
		algos = "sha256"
	}
	sAlgos := strings.Split(algos, ",")
	sCols := make([]int, len(sAlgos)+1)
	headFound := false
	for {
		cCols, err := csr.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !headFound {
			headFound = true
			for i, col := range cCols {
				if col == "relPath" {
					sCols[0] = i + 1
					continue
				}
				for j, algo := range sAlgos {
					if col == algo {
						sCols[j+1] = i + 1
					}
				}
			}
			for k, ix := range sCols {
				if ix > 0 {
					sCols[k]--
					continue
				}
				if k == 0 {
					return fmt.Errorf("headers: relPath missing")
				}
				return fmt.Errorf("headers: %s missing", sAlgos[k-1])
			}
			continue
		}
		relPath := ""
		if len(cCols) > sCols[0] {
			relPath = cCols[sCols[0]]
		}
		var sCss []string
		for i := range len(sCols) - 1 {
			ix := sCols[i+1]
			if len(cCols) <= ix {
				break
			}
			sCss = append(sCss, fmt.Sprintf("%s:%s", sAlgos[i], cCols[ix]))
		}
		le, err := oplm.GetLogicalEntry(relPath)
		if err != nil {
			return err
		}
		if le == nil {
			le = &opelog.LogicalEntry{}
		}
		seNum := le.AddOrShareSe(&opelog.StoredEntry{})
		tcss, err := common.Checksums2TypedChecksums(strings.Join(sCss, ","))
		if err != nil {
			return err
		}
		var tcsNums []int32
		for _, tcs := range tcss {
			tcsNums = append(tcsNums, le.AddOrShareTcs(tcs))
		}
		le.SourceEvents = append(le.SourceEvents,
			&opelog.Event{
				Kind: opelog.EVT_INV_LOADED, TimeStamp: time.Now().Unix(),
				SeNum: seNum, TcsNums: tcsNums})
		if err = oplm.PutLogicalEntry(relPath, le); err != nil {
			return err
		}
	}
}
