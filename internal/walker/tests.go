package walker

import (
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
)

func PrepareAugmentedTestFilesTree(td string, maxDirs, maxFiles, childrenPerDir, maxFileSize int) (sumAddedDirs int, sumAddedFiles int, err error) {
	/*
		/path/to/td/dau
		/path/to/td/dau/dAddFiles
		/path/to/td/dau/dAddFiles/dRemoved
		/path/to/td/dau/dAddFiles/dStay
		/path/to/td/dau/dAddFiles/fRemoved.dat
		/path/to/td/dau/dAddFiles/fStay.dat
		/path/to/td/dau/dLinks
		/path/to/td/dau/dLinks/dotDot.lnk
		/path/to/td/dau/dLinks/fRef.dat
		/path/to/td/dau/dLinks/fRef.lnk
		/path/to/td/dau/dLinks/notYet.lnk
		/path/to/td/dau/dMod
		/path/to/td/dau/dRO
	*/
	sumAddedDirs, sumAddedFiles, err = common.MakeAugmentedTestFilesTree(td, maxDirs, maxFiles, childrenPerDir, maxFileSize)
	lgr := common.GetLogger()
	if _, err = RecChmodRO(lgr, 2, localfiles.MakeLocalFilesDssa(), path.Join(td, "dau/dRO"), "source"); err != nil {
		return
	}
	moinsUne := int64((30*365+7)*24*3600 - 1)
	if _, err = RecTouch(lgr, 2, localfiles.MakeLocalFilesDssa(), path.Join(td, "dau/dMod"), "source", moinsUne); err != nil {
		return
	}
	return
}

func SetTestDirRW(td string, da string) {
	RecChmodRW(common.GetLogger(), 2, localfiles.MakeLocalFilesDssa(), path.Join(td), da)
}
