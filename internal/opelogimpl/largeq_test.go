package opelogimpl

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestLargeqSimple(t *testing.T) {
	// t.Skip("won't work")
	const conc = 4
	lgr := common.DbgLogger()
	td := t.TempDir()
	lq, err := NewLargeQ(lgr, td, 10000)
	require.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		lgr.Debug("push started")
		for i := range 100000 {
			require.NoError(t, lq.Put(fmt.Sprintf("%d", i)))
		}
		require.NoError(t, lq.Put("EOF"))
		lgr.Debug("push done")
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)
	subTotals := [conc]int{}
	for cons := range conc {
		wg.Add(1)
		go func(cons int) {
			lgr.Debug("pull started", "cons", cons)
			for {
				si, err := lq.Get()
				if err != nil {
					lgr.Error("pull err", "cons", cons, "err", err)
					break
				}
				if string(si) == "EOF" {
					if err := lq.Close(); err != nil {
						lgr.Error("pull close err", "cons", cons, "err", err)
						break
					}
				} else {
					subTotals[cons]++
				}
			}
			lgr.Debug("pull done", "cons", cons)
			wg.Done()
		}(cons)
	}
	wg.Wait()
	total := 0
	for cons := range conc {
		total += subTotals[cons]
	}
	require.Equal(t, 100000, total)
}

func TestCond(t *testing.T) {
	var pokemonList = []string{"Pikachu", "Charmander", "Squirtle", "Bulbasaur", "Jigglypuff"}
	var cond = sync.NewCond(&sync.Mutex{})
	var pokemon = ""
	var wg sync.WaitGroup
	lgr := common.DbgLogger()
	consDone := false

	wg.Add(1)
	go func() {
		lgr.Debug("consumer starts")
		cond.L.Lock()
		defer cond.L.Unlock()

		// waits until Pikachu appears
		for pokemon != "Pikachu" {
			cond.Wait()
			lgr.Debug("consumer caught", "pokemon", pokemon)
		}
		lgr.Debug("consumer caught Pikachu")
		consDone = true
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// Every 1ms, a random Pokémon appears
		for i := 0; i < 100 && !consDone; i++ {
			time.Sleep(time.Millisecond)
			cond.L.Lock()
			pokemon = pokemonList[rand.Intn(len(pokemonList))]
			lgr.Debug("producer signals", "pokemon", pokemon)
			cond.L.Unlock()
			cond.Signal()
		}
		wg.Done()
	}()
	wg.Wait()
}
