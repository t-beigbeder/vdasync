package encrypted

import (
	"fmt"
	"io"
	"log/slog"
	"path"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
)

func (msts *m2edsvc) listAll(idx map[string]*dssa.DataEntry, path_ string) error {
	des, err := msts.List(path_)
	if err != nil {
		return err
	}
	for _, de := range des {
		if de.IsDir {
			if err = msts.listAll(idx, de.Path); err != nil {
				return err
			}
		}
		idx[de.Path] = de
	}
	return nil
}

func CheckIndex(lgr *slog.Logger, underlying dssa.Dssa, rootPath string, ageIdentities []string, ageRecipients []string, dryRun bool) error {
	msts := &m2edsvc{
		M2StSvc: metasts.M2StSvc{
			Lgr: lgr,
			StSvc: &m2edsStSvc{
				dss:           underlying,
				rootPath:      rootPath,
				ageIdentities: ageIdentities,
				ageRecipients: ageRecipients,
			},
		},
	}
	if err := msts.NewSession(); err != nil {
		return fmt.Errorf("CheckIndex: cannot load metadata: %v", err)
	}
	defer msts.EndSession()
	idx := map[string]*dssa.DataEntry{}
	if err := msts.listAll(idx, "/"); err != nil {
		return fmt.Errorf("CheckIndex: metadata inconsistent %v", err)
	}
	for path_, de := range idx {
		if de.IsDir || de.IsSymLink {
			continue
		}
		ap := path.Join(rootPath, common.Id2Path(de.Id))
		lgr.Debug("CheckIndex: reading data", "path", path_, "ePath", ap)
		ede, err := underlying.Stat(ap)
		if err != nil || ede.IsDir || ede.IsSymLink {
			lgr.Error("CheckIndex: Stat error", "path", path_, "ePath", ap, "id", de.Id, "err", err)
			if !dryRun {
				msts.Del(path_)
			}
			continue
		}
		sr, err := underlying.GetReadCloser(ap)
		if err != nil {
			lgr.Error("CheckIndex: GetReadCloser error", "path", path_, "ePath", ap, "err", err)
			if !dryRun {
				underlying.Rm(ap)
				msts.Del(path_)
			}
			continue
		}
		er, err := makeEReader(sr, ageIdentities...)
		if err != nil {
			lgr.Error("CheckIndex: makeEReader error", "path", path_, "ePath", ap, "err", err)
			if !dryRun {
				underlying.Rm(ap)
				msts.Del(path_)
			}
			continue
		}
		defer er.Close()
		_, err = io.Copy(io.Discard, er)
		if err != nil {
			lgr.Error("CheckIndex: read error", "path", path_, "ePath", ap, "err", err)
			if !dryRun {
				underlying.Rm(ap)
				msts.Del(path_)
			}
			continue
		}
	}
	if err := msts.EndSession(); err != nil {
		return err
	}
	return nil
}
