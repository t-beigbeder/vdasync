package common

import (
	"log/slog"
	"time"
)

type BgJob struct {
	lgr      *slog.Logger
	job      func()
	stopChan chan bool
	doneChan chan bool
}

type PeriodicJob struct {
	BgJob
	periodSecs int64
}

func (pb *PeriodicJob) Start() {
	ticker := time.NewTicker(time.Duration(pb.periodSecs * int64(time.Second)))
LOOP:
	for {
		select {
		case <-pb.stopChan:
			pb.lgr.Debug("stopped")
			break LOOP
		case tick := <-ticker.C:
			pb.lgr.Debug("tick", "tick", tick)
			pb.job()
		}
	}
	pb.lgr.Debug("stopped")
	pb.job()
	close(pb.doneChan)
}

func (pb *BgJob) Stop() {
	pb.lgr.Debug("stopping")
	close(pb.stopChan)
}

func (pb *BgJob) Done() chan bool {
	return pb.doneChan
}

func NewPeriodicJob(lgr *slog.Logger, periodSecs int64, job func()) *PeriodicJob {
	return &PeriodicJob{
		BgJob: BgJob{
			lgr:      lgr,
			job:      job,
			stopChan: make(chan bool),
			doneChan: make(chan bool),
		},
		periodSecs: periodSecs,
	}
}
