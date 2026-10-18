package common

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

func RateLimit(lter *rate.Limiter, tokens int, doer func(int), delayCb func(time.Duration)) error {
	rsv := lter.ReserveN(time.Now(), tokens)
	if !rsv.OK() {
		return fmt.Errorf("cannot reserve %d tokens (burst %d)", tokens, lter.Burst())
	}
	d := rsv.Delay()
	if d > 0 && delayCb != nil {
		delayCb(d)
	}
	time.Sleep(d)
	if doer != nil {
		doer(tokens)
	}
	return nil
}

type EvBdwLter struct {
	evLter     *rate.Limiter
	bdwLter    *rate.Limiter
	delayCb    func(kw string, d time.Duration)
	evCount    int64
	bytesCount int64
}

func NewEvBdwLter(evFreq float64, bdw float64, delayCb func(string, time.Duration)) *EvBdwLter {
	ebl := &EvBdwLter{delayCb: delayCb}
	if evFreq != 0.0 {
		ebl.evLter = rate.NewLimiter(rate.Limit(evFreq), 1)
	}
	if bdw != 0.0 {
		ebl.bdwLter = rate.NewLimiter(rate.Limit(bdw/1.3), int(bdw/30)+1)
	}
	return ebl
}

func (ebl *EvBdwLter) Counts() (int64, int64) {
	return ebl.evCount, ebl.bytesCount
}

func (ebl *EvBdwLter) RtlThis(nBytes int) error {
	if ebl.evLter != nil {
		rsv := ebl.evLter.Reserve()
		if !rsv.OK() {
			return errors.New("cannot reserve a single token for events")
		}
		d := rsv.Delay()
		if d > 0 && ebl.delayCb != nil {
			ebl.delayCb("e", d)
		}
		time.Sleep(d)
		ebl.evCount++
	}
	if ebl.bdwLter != nil && nBytes > 0 {
		rsv := ebl.bdwLter.ReserveN(time.Now(), nBytes)
		if !rsv.OK() {
			return fmt.Errorf("cannot rate limit %d bytes (max %d)", nBytes, ebl.bdwLter.Burst())
		}
		d := rsv.Delay()
		if d > time.Millisecond && ebl.delayCb != nil {
			ebl.delayCb("b", d)
		}
		time.Sleep(d)
		ebl.bytesCount += int64(nBytes)
	}
	return nil
}
