package common

import (
	"context"
	"os"
	"sync"
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
	lgr := GetLogger()
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
	lgr := GetLogger()
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
	lgr := GetLogger()
	l10gbps := rate.Limit(128 * 10e6)
	l10gbpser := rate.NewLimiter(l10gbps, 10e6)
	for i := range 1000 {
		l10gbpser.WaitN(ctx, 10e5)
		lgr.Debug("done", "i", i)
	}
}

func TestLargeCbLter(t *testing.T) {
	lgr := GetLogger()
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
	lgr := GetLogger()
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

func TestEvBdwLterCountsOnly(t *testing.T) {
	lgr := GetLogger()
	i := 0
	done := false

	ebl := NewEvBdwLter(0.0, 0.0, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
		if i >= 1000 {
			done = true
		}
	})
	ebl.Record()
	for ; !done; i++ {
		err := ebl.RtlThis(10e6)
		require.NoError(t, err)
	}
	nEv, nBytes, du, evr, bdw := ebl.Counts()
	lgr.Debug("test bandwidth", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
}

func TestEvBdwLterEvOnly(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	lgr := GetLogger()
	i := 0
	done := false

	ebl := NewEvBdwLter(3./20, 0.0, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
		if i >= 3 {
			done = true
		}
	})
	ebl.Record()
	for ; !done; i++ {
		err := ebl.RtlThis(10e6)
		require.NoError(t, err)
	}
	nEv, nBytes, du, evr, bdw := ebl.Counts()
	lgr.Debug("test bandwidth", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
}

func TestEvBdwLterBdwOnly(t *testing.T) {
	lgr := GetLogger()
	i := 0

	ebl := NewEvBdwLter(0.0, 1e8, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
	})
	ebl.Record()
	for ; i < 200; i++ {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(1 * time.Millisecond)
		}
		err := ebl.RtlThis(20000)
		require.NoError(t, err)
	}
	nEv, nBytes, du, evr, bdw := ebl.Counts()
	lgr.Debug("test bandwidth", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
}

func TestEvBdwLterBoth(t *testing.T) {
	lgr := GetLogger()
	i := 0

	ebl := NewEvBdwLter(1400, 1e8, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
	})
	ebl.Record()
	for ; i < 400; i++ {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(100 * time.Microsecond)
			du, evr, bdw := ebl.Stats()
			lgr.Debug("cur bandwidth", "duration", du.Seconds(), "evr", evr, "bdw", bdw)
		}
		err := ebl.RtlThis(80000)
		require.NoError(t, err)
	}
	nEv, nBytes, du, evr, bdw := ebl.Counts()
	lgr.Debug("test bandwidth", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
}

func TestEvBdwLterPond(t *testing.T) {
	lgr := GetLogger()
	i := 0

	ebl := NewEvBdwLter(1400, 1e8, func(kw string, d time.Duration) {
		lgr.Debug("waiting", "kw", kw, "d", d, "i", i)
	})
	ebl.Record()
	for ; i < 400; i++ {
		if (i+1)%10 == 0 {
			lgr.Debug("sleep...", "i", i)
			time.Sleep(100 * time.Microsecond)
			du, evr, bdw := ebl.Stats()
			lgr.Debug("cur bandwidth", "duration", du.Seconds(), "evr", evr, "bdw", bdw)
		}
		err := ebl.PondeRtlThis(99, 80000)
		require.NoError(t, err)
	}
	nEv, nBytes, du, evr, bdw := ebl.Counts()
	lgr.Debug("test bandwidth", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
}

func TestParseEvBdwbRates(t *testing.T) {
	e, b, err := ParseEvBdwbRates("")
	require.NoError(t, err)
	require.Equal(t, float64(0), e)
	require.Equal(t, float64(0), b)
	e, b, err = ParseEvBdwbRates("1e5")
	require.NoError(t, err)
	require.Equal(t, float64(100000), e)
	require.Equal(t, float64(0), b)
	e, b, err = ParseEvBdwbRates("1e5:")
	require.NoError(t, err)
	require.Equal(t, float64(100000), e)
	require.Equal(t, float64(0), b)
	e, b, err = ParseEvBdwbRates("1e5:128e6")
	require.NoError(t, err)
	require.Equal(t, float64(100000), e)
	require.Equal(t, float64(128000000), b)
	e, b, err = ParseEvBdwbRates(":128e6")
	require.NoError(t, err)
	require.Equal(t, float64(0), e)
	require.Equal(t, float64(128000000), b)
}

func TestCancelEvLter(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	lgr := GetLogger()
	cel, cf := NewCancelEvLter(5, func(doing chan any, d time.Duration) {
		lgr.Debug("cb", "d", d)
		tmer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-doing:
			tmer.Stop()
			lgr.Debug("callback received doing")
		case <-tmer.C:
			lgr.Debug("callback received tmer expiration")
		}
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		lgr.Debug("producing...")
		for i := range 20 {
			lgr.Debug("RtlThis", "i", i)
			if err := cel.RtlThis(); err != nil {
				lgr.Debug("RtlThis returned", "err", err)
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
		lgr.Debug("producing stopping cel")
		cel.Stop()
		time.Sleep(100 * time.Millisecond)
		lgr.Debug("producing done")
		wg.Done()
	}()
	time.Sleep(2 * time.Second)
	lgr.Debug("test canceling cel")
	cf()
	lgr.Debug("test waiting done...")
	cel.Wait()
	lgr.Debug("test waiting sync...")
	wg.Wait()
	lgr.Debug("test done")
}
