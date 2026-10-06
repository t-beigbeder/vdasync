package common

import (
	"fmt"
	"regexp"
)

func compileRe(ss []string) ([]*regexp.Regexp, error) {
	rs := []*regexp.Regexp{}
	for _, s := range ss {
		r, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func ReFromSlice(ress []string, label string) ([]*regexp.Regexp, error) {
	regs, err := compileRe(ress)
	if err != nil {
		return nil, fmt.Errorf("compile regexp %s: %s", label, err)
	}
	return regs, nil
}

func ReFromFile(path_ string, label string) ([]*regexp.Regexp, error) {
	if path_ == "" {
		return nil, nil
	}
	ress, err := FileLines(path_)
	if err != nil {
		return nil, fmt.Errorf("read regexp from %s: %s", label, err)
	}
	return ReFromSlice(ress, label)
}

func isIncluded(relPath string, inclRegs []*regexp.Regexp) bool {
	if len(inclRegs) == 0 {
		return true
	}
	for _, ire := range inclRegs {
		if ire.MatchString(relPath) {
			return true
		}
	}
	return false
}

func isExcluded(relPath string, exclRegs []*regexp.Regexp) bool {
	for _, ere := range exclRegs {
		if ere.MatchString(relPath) {
			return true
		}
	}
	return false
}

// IsIncluded checks relPath is included (empty list means all) or not excluded
func IsIncluded(relPath string, inclRegs, exclRegs []*regexp.Regexp) bool {
	return isIncluded(relPath, inclRegs) && !isExcluded(relPath, exclRegs)
}
