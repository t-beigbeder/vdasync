package common

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

func GenId() (string, error) {
	uuid_, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%064x", sha256.Sum256([]byte(uuid_.String()))), nil
}

func Id2Path(id string) []string {
	if len(id) != 64 {
		return nil
	}
	return []string{id[0:3], id[3:6], id[6:]}
}

func Path2Id(path_ []string) string {
	if len(path_) < 3 {
		return ""
	}
	lp := len(path_)
	return path_[lp-3] + path_[lp-2] + path_[lp-1]
}
