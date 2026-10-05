package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

const stdColNames = "relPath,isDir,size,mTime,isSymLink,symLinkTarget"

var colIdx = map[string]int{}
var idxCol = map[int]string{}
var halgoNames []string

func init() {
	for hc := range maps.Values(opeloggrpc.HalgoCode_value) {
		colIdx[common.AlgoCode(hc).String()] = -1
	}
	halgoNames = slices.Collect(maps.Keys(colIdx))
	for _, cn := range strings.Split(stdColNames, ",") {
		colIdx[cn] = -1
	}
}

func setupColIdx(row []string, algos string) error {
	colNames := slices.Collect(maps.Keys(colIdx))
	requestedAlgoNames := strings.Split(algos, ",")
	for ix, col := range row {
		if !slices.Contains(colNames, col) {
			return fmt.Errorf("setupColIdx: column name %s (%d) is not standard", col, ix+1)
		}
		colIdx[col] = ix
		idxCol[ix] = col
	}
	for _, ran := range requestedAlgoNames {
		idx, ok := colIdx[ran]
		if !ok {
			return fmt.Errorf("setupColIdx: requested hash algo %s is not standard", ran)
		}
		if idx == -1 {
			return fmt.Errorf("setupColIdx: requested hash algo %s is not a column", ran)
		}
	}
	for _, stdCn := range strings.Split(stdColNames, ",") {
		if colIdx[stdCn] == -1 {
			return fmt.Errorf("setupColIdx: standard column %s is not a column", stdCn)
		}
	}
	return nil
}

func dcBool(v string) (bool, error) {
	if v == "1" {
		return true, nil
	}
	if v == "0" {
		return false, nil
	}
	return false, fmt.Errorf("bad bool format %s", v)
}

func dcTime(v string) (int64, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func InventoryCsvImport(oplm opelog.OpeLogManager, inventTs int64, csvPath string, algos string) error {
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
	headFound := false
	for {
		row, err := csr.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !headFound {
			headFound = true
			if err := setupColIdx(row, algos); err != nil {
				return err
			}
			continue
		}
		namedValues := map[string]string{}
		for ix, col := range row {
			namedValues[idxCol[ix]] = col
		}
		var sCss []string
		for _, algo := range strings.Split(algos, ",") {
			col := namedValues[algo]
			if col == "" {
				continue
			}
			sCss = append(sCss, fmt.Sprintf("%s:%s", algo, col))
		}
		relPath := namedValues["relPath"]
		le, err := oplm.GetLogicalEntry(relPath)
		if err != nil {
			return err
		}
		if le == nil {
			le = opelog.NewLogicalEntry()
		}
		tcss, err := common.Checksums2TypedChecksums(strings.Join(sCss, ","))
		if err != nil {
			return err
		}
		isDir, err := dcBool(namedValues["isDir"])
		if err != nil {
			return err
		}
		size, err := strconv.ParseInt(namedValues["size"], 10, 64)
		if err != nil && namedValues["size"] != "" {
			return err
		}
		mTime, err := dcTime(namedValues["mTime"])
		if err != nil {
			return err
		}
		isSymLink, err := dcBool(namedValues["isSymLink"])
		if err != nil {
			return err
		}
		se := &opelog.StoredEntry{
			IsDir:         isDir,
			Size:          size,
			Mtime:         mTime,
			IsSymLink:     isSymLink,
			SymLinkTarget: namedValues["symLinkTarget"],
		}
		le.SetState(0, inventTs, false, opelog.STC_UNSPECIFIED, "", se, tcss, 0)
		if err = oplm.PutLogicalEntry(relPath, le); err != nil {
			return err
		}
	}
}
