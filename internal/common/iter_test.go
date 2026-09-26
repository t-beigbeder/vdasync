package common

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcat(t *testing.T) {
	a := slices.Values([]string{"a"})
	b := slices.Values([]string{"b"})
	ab := slices.Values([]string{"a", "b"})
	es := slices.Values([]string{})
	require.Equal(t, []string{"a", "b"}, slices.Collect(Concat(a, b)))
	require.Equal(t, []string{"a", "b"}, slices.Collect(Concat(ab)))
	require.Equal(t, []string{"a", "b"}, slices.Collect(Concat(ab, nil)))
	require.Equal(t, []string{"a", "b"}, slices.Collect(Concat(nil, ab, nil)))
	require.Equal(t, []string{"a", "b"}, slices.Collect(Concat(ab, es)))
	require.Equal(t, []string{"a", "b"}, slices.Collect(Concat(es, ab, es)))
}
