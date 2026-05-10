package walker

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestRecChmodRORWMtime(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	_, _, err := common.MakeAugmentedTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	_, err = RecChmodRO(lgr, 2, localfiles.MakeLocalFilesDssa(), td1, "lds4tests")
	require.Nil(t, err)
	_, err = RecChmodRW(lgr, 2, localfiles.MakeLocalFilesDssa(), td1, "lds4tests")
	require.Nil(t, err)
	_, err = RecTouch(lgr, 2, localfiles.MakeLocalFilesDssa(), td1, "lds4tests", int64(30*365*24*3600))
	require.Nil(t, err)
}
