package common

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestSimpleLter(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	ctx := context.Background()
	lgr := DbgLogger()
	require.True(t, true)
	l10 := rate.Limit(10.0)
	l10ter := rate.NewLimiter(l10, 5)
	for i := range 30 {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(300 * time.Millisecond)
		}
		// lgr.Debug("wait...", "i", i)
		l10ter.Wait(ctx)
		lgr.Debug("done", "i", i)
	}
}

func TestLterCb(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	lgr := DbgLogger()
	l10 := rate.Limit(10.0)
	l10ter := rate.NewLimiter(l10, 5)
	for i := range 30 {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(300 * time.Millisecond)
		}
		err := RateLimit(l10ter, 1,
			func(tk int) { lgr.Debug("done", "i", i, "tk", tk) },
			func(d time.Duration) { lgr.Debug("waiting", "d", d) },
		)
		if err != nil {
			lgr.Error("rate limit", "lter", l10ter.Burst(), "err", err)
		}
	}
}

func TestLargeLter(t *testing.T) {
	ctx := context.Background()
	lgr := DbgLogger()
	l10gbps := rate.Limit(128 * 10e6)
	l10gbpser := rate.NewLimiter(l10gbps, 10e6)
	for i := range 1000 {
		l10gbpser.WaitN(ctx, 10e5)
		lgr.Debug("done", "i", i)
	}
}

func TestLargeCbLter(t *testing.T) {
	lgr := DbgLogger()
	l10gbps := rate.Limit(128 * 10e6)
	l10gbpser := rate.NewLimiter(l10gbps, 10e6)
	for i := range 1000 {
		RateLimit(l10gbpser,
			10e5,
			func(tk int) { lgr.Debug("done", "i", i, "tk", tk) },
			func(d time.Duration) { lgr.Debug("waiting", "d", d) })
	}
}

func TestSlowMotionLter(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	lgr := DbgLogger()
	lim := rate.Limit(3. / 20)
	limTer := rate.NewLimiter(lim, 3)
	done := false
	for i := 0; !done; i++ {
		err := RateLimit(limTer,
			1,
			func(tk int) { lgr.Debug("done", "i", i, "tk", tk) },
			func(d time.Duration) {
				lgr.Debug("waiting", "d", d)
				if i >= 3 {
					done = true
				}
			})
		require.NoError(t, err)
	}
}

func TestEvBdwLterEvOnly(t *testing.T) {
	lgr := DbgLogger()
	i := 0
	done := false

	ebl := NewEvBdwLter(3./20, 0.0, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
		if i >= 3 {
			done = true
		}
	})
	for ; !done; i++ {
		err := ebl.RtlThis(0)
		require.NoError(t, err)
	}
}

func TestEvBdwLterBdwOnly(t *testing.T) {
	lgr := DbgLogger()
	i := 0

	ebl := NewEvBdwLter(0.0, 128*10e6, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
	})
	st := time.Now()
	for ; i < 50; i++ {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(30 * time.Millisecond)
		}
		err := ebl.RtlThis(10e6)
		require.NoError(t, err)
	}
	et := time.Now()
	_, nBytes := ebl.Counts()
	delta := et.Sub(st)
	tbdw := float64(nBytes) / delta.Seconds()
	lgr.Debug("test bandwidth", "nbytes", nBytes, "duration", delta.Seconds(), "tbdw", tbdw)
}

func TestEvBdwLterBoth(t *testing.T) {
	lgr := DbgLogger()
	i := 0

	ebl := NewEvBdwLter(160., 128*10e6, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
	})
	st := time.Now()
	for ; i < 100; i++ {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(30 * time.Millisecond)
		}
		err := ebl.RtlThis(10e6)
		require.NoError(t, err)
	}
	et := time.Now()
	_, nBytes := ebl.Counts()
	delta := et.Sub(st)
	tbdw := float64(nBytes) / delta.Seconds()
	lgr.Debug("test bandwidth", "nbytes", nBytes, "duration", delta.Seconds(), "tbdw", tbdw)
}
