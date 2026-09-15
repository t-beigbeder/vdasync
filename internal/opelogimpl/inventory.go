package opelogimpl

import (
	"errors"
	"io"

	"github.com/t-beigbeder/vdasync/opelog"
)

func InventoryCsvImport(olm opelog.OpeLogManager, rdr io.Reader) error {
	return errors.ErrUnsupported
}
