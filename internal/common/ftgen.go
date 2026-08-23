package common

import (
	"fmt"
	mrand "math/rand"
	"os"
	"sync"
)

type ftGenCtx struct {
	root                 string
	nd, nf, depth, maxFs int
	noCreate             bool
	conc                 int
	dirs                 [][]int
	dirsPerDepth         []map[string][]int
	files                [][]int
	fGenQueue            chan []int
}

func iDir2dKey(ctx *ftGenCtx, iPath []int) string {
	rs := ""
	for _, i := range iPath {
		if rs == "" {
			rs = fmt.Sprintf("%d", i)
		} else {
			rs += fmt.Sprintf("/%d", i)
		}
	}
	return rs
}

func ftgenPath(ctx *ftGenCtx, iPath []int, isFile bool) string {
	rs := ctx.root
	for x, i := range iPath {
		if isFile && x == len(iPath)-1 {
			rs += fmt.Sprintf("/f%06d", i)
		} else {
			rs += fmt.Sprintf("/d%04d", i)
		}
	}
	return rs
}

func ftGenDir(ctx *ftGenCtx, dDepth int) error {
	iDir := make([]int, dDepth+1)
	for cDepth := range dDepth + 1 {
		ddn := len(ctx.dirsPerDepth[cDepth])
		if ddn == 0 {
			ctx.dirsPerDepth[cDepth] = map[string][]int{}
			iDir[cDepth] = 0
		} else if cDepth < dDepth {
			iDir[cDepth] = mrand.Intn(ddn)
			continue
		} else {
			iDir[cDepth] = ddn
		}
		ctx.dirsPerDepth[cDepth][iDir2dKey(ctx, iDir)] = iDir
		ctx.dirs = append(ctx.dirs, iDir)
	}
	if ctx.noCreate {
		return nil
	}
	if err := os.MkdirAll(ftgenPath(ctx, iDir, false), 0750); err != nil {
		return err
	}
	return nil
}

func ftGenFile(ctx *ftGenCtx) error {
	var wg sync.WaitGroup
	var fGenQueue chan []int
	var fGenErrors chan error
	if !ctx.noCreate && ctx.conc > 1 {
		fGenQueue = make(chan []int, ctx.conc)
		fGenErrors = make(chan error, ctx.conc)
		for range ctx.conc {
			wg.Add(1)
			go func() {
				for iPath := range fGenQueue {
					t := iPath
					_ = t
					if err := MakeTestFile(ftgenPath(ctx, iPath, true), mrand.Intn(ctx.maxFs-1)+1); err != nil {
						fGenErrors <- err
						break
					}
				}
				wg.Done()
			}()
		}
	}

	for _, xf := range mrand.Perm(ctx.nf) {
		xd := mrand.Intn(len(ctx.dirs) + 1)
		ctx.files[xf] = []int{}
		if xd > 0 {
			ctx.files[xf] = ctx.dirs[xd-1]
		}
		ctx.files[xf] = append(ctx.files[xf], xf)
		if ctx.noCreate {
			continue
		}
		if ctx.conc < 2 {
			if err := MakeTestFile(ftgenPath(ctx, ctx.files[xf], true), mrand.Intn(ctx.maxFs-1)+1); err != nil {
				return err
			}
			continue
		}
		fGenQueue <- ctx.files[xf]
	}
	var gErr error
	if !ctx.noCreate && ctx.conc > 1 {
		go func() {
			for err := range fGenErrors {
				gErr = err
			}
		}()
		close(fGenQueue)
		wg.Wait()
		close(fGenErrors)
	}
	return gErr
}

func doFtGen(ctx *ftGenCtx) error {
	for xd := 0; xd < ctx.nd; xd++ {
		if err := ftGenDir(ctx, mrand.Intn(ctx.depth)); err != nil {
			return err
		}
	}
	if err := ftGenFile(ctx); err != nil {
		return err
	}
	return nil
}

func FileTreeGenerate(root string, nd, nf, depth, maxFs int, noCreate bool, conc int) error {
	ctx := &ftGenCtx{
		root: root,
		nd:   nd, nf: nf, depth: depth, maxFs: maxFs,
		noCreate:     noCreate,
		conc:         conc,
		dirs:         [][]int{},
		dirsPerDepth: make([]map[string][]int, depth),
		files:        make([][]int, nf),
	}
	return doFtGen(ctx)
}
