package common

import (
	"iter"
	"slices"
)

// borrowed from https://pkg.go.dev/github.com/jub0bs/iterutil#Concat
func Concat[E any](seqs ...iter.Seq[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, seq := range seqs {
			if seq == nil {
				continue
			}
			for e := range seq {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// compare unordered slices
func SlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	if len(s1) == 0 && len(s2) == 0 {
		return true
	}
	if len(s1) == 0 || len(s2) == 0 {
		return false
	}
	for _, e1 := range s1 {
		if !slices.Contains(s2, e1) {
			return false
		}
	}
	return true
}
