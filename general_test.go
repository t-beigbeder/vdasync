package otvldtacsy

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestGeneral(t *testing.T) {
	require.False(t, common.FileExists("nono"))
}
