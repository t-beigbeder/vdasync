package common

import (
	"crypto/sha256"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
)

func GenId() (string, error) {
	uuid_, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%064x", sha256.Sum256([]byte(uuid_.String()))), nil
}

func Id2Path(id string) string {
	if len(id) != 64 {
		return ""
	}
	return path.Join([]string{id[0:2], id[2:]}...)
}

func Path2Id(path_ string) string {
	pe := strings.Split(path_, "/")
	if len(pe) < 2 {
		return ""
	}
	lp := len(pe)
	return pe[lp-2] + pe[lp-1]
}
