package walker

import (
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
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
	dss := localfiles.MakeLocalFilesDssa()
	if _, err = RecChmodRO(lgr, 2, dss, path.Join(td, "dau/dRO"), "source"); err != nil {
		return
	}
	moinsUne := int64((30*365+6)*24*3600 + 12*3600)
	if _, err = RecTouch(lgr, 2, dss, path.Join(td, "dau/dMod"), "source", moinsUne); err != nil {
		return
	}
	return
}

func UpdateAugmentedTestFilesTree(td string, maxDirs, maxFiles, childrenPerDir, maxFileSize int) (sumAddedDirs int, sumAddedFiles int, err error) {
	var (
		addedDirs  int
		addedFiles int
		sde        *dssa.DataEntry
	)
	lgr := common.GetLogger()
	dss := localfiles.MakeLocalFilesDssa()
	if _, err = RecChmodRW(lgr, 2, dss, path.Join(td, "dau/dRO"), "source"); err != nil {
		return
	}

	if err = os.Mkdir(path.Join(td, "dau/dAddFiles/dNewOne"), 0750); err != nil {
		return
	}
	sumAddedDirs += 1
	if err = os.Mkdir(path.Join(td, "dau/dRO/dNewOne"), 0750); err != nil {
		return
	}
	sumAddedDirs += 1

	RemoveAll(lgr, 2, dss, path.Join(td, "dau/dAddFiles/dRemoved"), "source", false)
	if err = os.Remove(path.Join(td, "dau/dAddFiles/fRemoved.dat")); err != nil {
		return
	}

	for _, fp := range []string{
		"dau/dAddFiles/dStay/fNewOne.dat", "dau/dAddFiles/fNewOne.dat",
		"dau/dRO/fNewOne.dat", "dau/dMod/dAddFiles/fStay.dat",
	} {
		if err = common.MakeTestFile(path.Join(td, fp), 1024); err != nil {
			return
		}
		sumAddedFiles += 1
	}
	mtfp := path.Join(td, "dau/dAddFiles/fStay.dat")
	if sde, err = dss.Stat(mtfp); err != nil {
		return
	}
	if err = common.MakeTestFile(mtfp, int(sde.Size)); err != nil {
		return
	}
	if err = dss.SetStat(sde, true, false); err != nil {
		return
	}
	sde2, _ := dss.Stat(mtfp)
	_ = sde2

	for _, sd := range []string{"dau/dAddFiles/dNewOne"} {
		addedDirs, addedFiles, err = common.AugmentTestFilesTree(path.Join(td, sd))
		if err != nil {
			return
		}
		sumAddedDirs += addedDirs
		sumAddedFiles += addedFiles
	}

	if _, err = RecChmodRO(lgr, 2, dss, path.Join(td, "dau/dRO"), "source"); err != nil {
		return
	}
	plusUne := int64((30*365+7)*24*3600 + 12*3600)
	if _, err = RecTouch(lgr, 2, dss, path.Join(td, "dau/dMod"), "source", plusUne); err != nil {
		return
	}

	return
}

func SetTestDirRW(td string, da string) {
	RecChmodRW(common.GetNullLogger(), 2, localfiles.MakeLocalFilesDssa(), td, da)
}
