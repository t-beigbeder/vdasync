package common

import "github.com/stretchr/testify/require"
import "testing"

func TestFreePort(t *testing.T) {
	p, err := GetFreePort()
	require.Nil(t, err)
	require.NotZero(t, p)
}
