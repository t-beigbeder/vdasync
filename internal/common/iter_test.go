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

func TestSlicesEqual(t *testing.T) {
	require.True(t, SlicesEqual(nil, nil))
	require.True(t, SlicesEqual([]string{}, nil))
	require.True(t, SlicesEqual(nil, []string{}))
	require.True(t, SlicesEqual([]string{}, []string{}))
	require.True(t, SlicesEqual([]string{}, []string{}))
	require.False(t, SlicesEqual(nil, []string{"a"}))
	require.False(t, SlicesEqual([]string{"a"}, nil))
	require.False(t, SlicesEqual([]string{"a"}, []string{"b"}))
	require.False(t, SlicesEqual([]string{"a"}, []string{"a", "b"}))
	require.True(t, SlicesEqual([]string{"a", "a", "b"}, []string{"a", "b", "b"}))
	require.False(t, SlicesEqual([]string{"a", "c"}, []string{"b", "a"}))
	require.True(t, SlicesEqual([]string{"a", "b"}, []string{"b", "a"}))
}
