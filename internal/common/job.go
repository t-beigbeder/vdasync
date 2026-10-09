package common

import (
	"log/slog"
	"time"
)

type PeriodicJob struct {
	lgr        *slog.Logger
	job        func()
	periodSecs int64
	stopChan   chan bool
	doneChan   chan bool
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

func (pb *PeriodicJob) Stop() {
	pb.lgr.Debug("stopping")
	close(pb.stopChan)
}

func (pb *PeriodicJob) Done() chan bool {
	return pb.doneChan
}

func NewPeriodicJob(lgr        *slog.Logger, periodSecs int64, job func()) *PeriodicJob{
	return &PeriodicJob{
		lgr: lgr,
		periodSecs: periodSecs,
		job: job,
		stopChan: make(chan bool),
		doneChan: make(chan bool),
	}
}