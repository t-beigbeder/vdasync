package common

import (
	"os"
	"path"
)

func XdgConfigDir() string {
	xcd := os.Getenv("XDG_CONFIG_HOME")
	if xcd != "" {
		return xcd
	}
	home := os.Getenv("HOME")
	if home != "" {
		return path.Join(home, ".config")
	}
	return ""
}
