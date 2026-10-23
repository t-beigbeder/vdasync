package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	mx           sync.Mutex
	hasPondEvs   bool
	evLter       *rate.Limiter
	pondEvLter   *rate.Limiter
	bdwLter      *rate.Limiter
	delayCb      func(kw string, d time.Duration)
	evCount      int64
	pondEvCount  int64
	bytesCount   int64
	sT           time.Time
	sRt          time.Time
	dEvCount     int64
	pondDEvCount int64
	dBytesCount  int64
}

func NewEvBdwLter(evFreq float64, bdw float64, delayCb func(string, time.Duration)) *EvBdwLter {
	ebl := &EvBdwLter{delayCb: delayCb}
	if evFreq != 0.0 {
		ebl.evLter = rate.NewLimiter(rate.Limit(evFreq), 1)
		ebl.pondEvLter = rate.NewLimiter(rate.Limit(evFreq*100), 100)
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
		if ebl.hasPondEvs {
			evr = float64(ebl.pondEvCount) / du.Seconds() / 100
		} else {
			evr = float64(ebl.evCount) / du.Seconds()
		}
		bdw = float64(ebl.bytesCount) / du.Seconds()
	}
	if ebl.hasPondEvs {
		evCount = int64(ebl.pondEvCount / 100)
	} else {
		evCount = ebl.evCount
	}
	bytesCount = ebl.bytesCount
	return
}

func (ebl *EvBdwLter) Record() time.Time {
	ebl.mx.Lock()
	defer ebl.mx.Unlock()
	now := time.Now()
	if ebl.sT.IsZero() {
		ebl.sT = now
		ebl.evCount, ebl.pondEvCount, ebl.bytesCount = 0, 0, 0
	}
	ebl.sRt = now
	ebl.dEvCount, ebl.pondDEvCount, ebl.dBytesCount = 0, 0, 0
	return ebl.sRt
}

func (ebl *EvBdwLter) Stats() (time.Duration, float64, float64) {
	ebl.mx.Lock()
	defer ebl.mx.Unlock()
	du := time.Since(ebl.sRt)
	var curEvr float64
	if ebl.hasPondEvs {
		curEvr = float64(ebl.dEvCount) / du.Seconds()
	} else {
		curEvr = float64(ebl.pondDEvCount) / du.Seconds() / 100
	}
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

func (ebl *EvBdwLter) PondeRtlThis(weight int, nBytes int) error {
	ebl.hasPondEvs = true
	if ebl.pondEvLter != nil {
		now := time.Now()
		rsv := ebl.pondEvLter.ReserveN(now, weight)
		if !rsv.OK() {
			return fmt.Errorf("cannot rate limit %d weight (max %d)", weight, ebl.pondEvLter.Burst())
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
	ebl.pondEvCount += int64(weight)
	ebl.pondDEvCount += int64(weight)
	ebl.bytesCount += int64(nBytes)
	ebl.dBytesCount += int64(nBytes)
	ebl.mx.Unlock()
	return nil
}

// ParseEvBdwbRates provides the events rate per second and bandwidth from string such as 10000:128e6
func ParseEvBdwbRates(ebrs string) (evRate, bandwidth float64, err error) {
	if ebrs == "" {
		return
	}
	ebrss := strings.Split(ebrs, ":")
	if ebrss[0] != "" {
		if evRate, err = strconv.ParseFloat(ebrss[0], 64); err != nil {
			return
		}
	}
	if len(ebrss) > 1 && ebrss[1] != "" {
		bandwidth, err = strconv.ParseFloat(ebrss[1], 64)
	}
	return
}

type CancelEvLter struct {
	evLter       *rate.Limiter
	cancelableCb func(doing chan any, d time.Duration)
	doing        chan any
	done         chan any
}

func NewCancelEvLter(evFreq float64, cancelableCb func(doing chan any, d time.Duration)) (*CancelEvLter, func()) {
	cel := &CancelEvLter{cancelableCb: cancelableCb}
	cel.evLter = rate.NewLimiter(rate.Limit(evFreq), 1)
	cel.doing = make(chan any)
	cel.done = make(chan any)
	return cel, func() { close(cel.doing) }
}

func (cel *CancelEvLter) RtlThis() error {
	var (
		isCanceled bool
		d          time.Duration
		tmer       *time.Timer
	)
	select {
	case <-cel.doing:
		isCanceled = true
	default:
		now := time.Now()
		rsv := cel.evLter.ReserveN(now, 1)
		if !rsv.OK() {
			return errors.New("cannot reserve a single token for events")
		}
		d = rsv.DelayFrom(now)
	}
	if isCanceled {
		return errors.New("event limiter canceled")
	}
	if d == 0 {
		return nil
	}
	if d > 0 && cel.cancelableCb != nil {
		cel.cancelableCb(cel.doing, d)
	}
	tmer = time.NewTimer(d)
	select {
	case <-tmer.C:
		tmer.Stop()
	case <-cel.doing:
		tmer.Stop()
	}
	return nil
}

func (cel *CancelEvLter) Stop() {
	close(cel.done)
}

func (cel *CancelEvLter) Wait() {
	<-cel.done
}
