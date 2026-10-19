package common

import (
	"errors"
	"fmt"
	"sync"
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
	mx          sync.Mutex
	evLter      *rate.Limiter
	bdwLter     *rate.Limiter
	delayCb     func(kw string, d time.Duration)
	evCount     int64
	bytesCount  int64
	sT          time.Time
	sRt         time.Time
	dEvCount    int64
	dBytesCount int64
}

func NewEvBdwLter(evFreq float64, bdw float64, delayCb func(string, time.Duration)) *EvBdwLter {
	ebl := &EvBdwLter{delayCb: delayCb}
	if evFreq != 0.0 {
		ebl.evLter = rate.NewLimiter(rate.Limit(evFreq), 1)
	}
	if bdw != 0.0 {
		burst := int(bdw*0.01) + 1
		if burst < 131072 {
			burst = 131072
		}
		ebl.bdwLter = rate.NewLimiter(rate.Limit(bdw*0.8), burst)
	}
	return ebl
}

func (ebl *EvBdwLter) Counts() (evCount int64, bytesCount int64, du time.Duration, evr float64, bdw float64) {
	if !ebl.sT.IsZero() {
		du = time.Since(ebl.sT)
		evr = float64(ebl.evCount) / du.Seconds()
		bdw = float64(ebl.bytesCount) / du.Seconds()
	}
	evCount, bytesCount = ebl.evCount, ebl.bytesCount
	return
}

func (ebl *EvBdwLter) Record() time.Time {
	ebl.mx.Lock()
	defer ebl.mx.Unlock()
	now := time.Now()
	if ebl.sT.IsZero() {
		ebl.sT = now
		ebl.evCount, ebl.bytesCount = 0, 0
	}
	ebl.sRt = now
	ebl.dEvCount, ebl.dBytesCount = 0, 0
	return ebl.sRt
}

func (ebl *EvBdwLter) Stats() (time.Duration, float64, float64) {
	ebl.mx.Lock()
	defer ebl.mx.Unlock()
	du := time.Since(ebl.sRt)
	curEvr := float64(ebl.dEvCount) / du.Seconds()
	curBdw := float64(ebl.dBytesCount) / du.Seconds()
	return du, curEvr, curBdw
}

func (ebl *EvBdwLter) RtlThis(nBytes int) error {
	if ebl.evLter != nil {
		now := time.Now()
		rsv := ebl.evLter.ReserveN(now, 1)
		if !rsv.OK() {
			return errors.New("cannot reserve a single token for events")
		}
		d := rsv.DelayFrom(now)
		if d > 0 && ebl.delayCb != nil {
			ebl.delayCb("e", d)
		}
		time.Sleep(d)
	}
	if ebl.bdwLter != nil && nBytes > 0 {
		now := time.Now()
		rsv := ebl.bdwLter.ReserveN(now, nBytes)
		if !rsv.OK() {
			return fmt.Errorf("cannot rate limit %d bytes (max %d)", nBytes, ebl.bdwLter.Burst())
		}
		d := rsv.DelayFrom(now)
		if d > 100*time.Microsecond && ebl.delayCb != nil {
			ebl.delayCb("b", d)
		}
		time.Sleep(d)
	}
	if ebl.evLter == nil && ebl.bdwLter == nil && ebl.delayCb != nil {
		ebl.delayCb("", 0)
	}
	ebl.mx.Lock()
	ebl.evCount++
	ebl.dEvCount++
	ebl.bytesCount += int64(nBytes)
	ebl.dBytesCount += int64(nBytes)
	ebl.mx.Unlock()
	return nil
}
