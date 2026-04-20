package common

import (
	"testing"
)

func TestLog(t *testing.T) {
	GetLogger().Info("a message", "with", "that")
}
