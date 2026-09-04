package opelogimpl

import (
	"errors"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opelog"
)

type OplWalker interface {
	Run() error
}

type oplWalkerImpl struct {
	lgr  *slog.Logger
	oplq opelog.Queue
	oplm opelog.OpeLogManager
	sds  dssa.Dssa
	tds  dssa.Dssa
}

func (ow *oplWalkerImpl) Run() error {
	return errors.ErrUnsupported
}

func NewOplWalker() OplWalker {
	return &oplWalkerImpl{}
}
