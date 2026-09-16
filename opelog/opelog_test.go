package opelog

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestEnum(t *testing.T) {
	require.Equal(t, "EVT_ABS", EVT_ABS.String())
	require.Equal(t, "ORI_LIST", ORI_LIST.String())
	common.DbgLogger().Debug("TestEnum", "evt", EVT_ATTS_CHG, "ori", ORI_STAT)
}
