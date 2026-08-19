package common

import (
	mrand "math/rand"
)

type ftGenCtx struct {
	root                 string
	nd, nf, depth, maxFs int
	create               bool
	dirs                 map[string][]int
	dirsPerDepth         []map[string][]int
}

func dirPath(ctx *ftGenCtx, iPath []int) string {
	return ""
}

func ftGenDir(ctx *ftGenCtx, dDepth int) error {
	if dDepth == 0 {
		ctx.dirs[""] = []int{}
	}
	pars := make([]int, dDepth)
	_ = pars
	for pDepth := 0; pDepth < dDepth; pDepth++ {
		pn := len(ctx.dirsPerDepth[pDepth])
		if pn == 0 {
			if err := ftGenDir(ctx, dDepth-1); err != nil {
				return err
			}
			pn = len(ctx.dirsPerDepth[pDepth])
		}
	}
	return nil
}

func ftGenFile(ctx *ftGenCtx) error {
	return nil
}

func doFtGen(ctx *ftGenCtx) error {
	for xd := 0; xd < ctx.nd; xd++ {
		if err := ftGenDir(ctx, mrand.Intn(ctx.depth+1) - 1); err != nil {
			return err
		}
	}
	for xf := 0; xf < ctx.nf; xf++ {
		if err := ftGenFile(ctx); err != nil {
			return err
		}
	}
	return nil
}

func FileTreeGenerate(root string, nd, nf, depth, maxFs int, create bool) error {
	ctx := &ftGenCtx{
		root: root,
		nd:   nd, nf: nf, depth: depth, maxFs: maxFs,
		dirs: map[string][]int{},
		dirsPerDepth: make([]map[string][]int, depth+1),
	}
	return doFtGen(ctx)
}
