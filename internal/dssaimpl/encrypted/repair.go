package encrypted

import (
	"fmt"
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

func CheckIndex(lgr *slog.Logger, underlying dssa.Dssa, rootPath string, ageIdentities []string, dryRun bool) error {
	msts := &m2edsvc{
		M2StSvc: metasts.M2StSvc{
			Lgr: lgr,
			StSvc: &m2edsStSvc{
				dss:           underlying,
				rootPath:      rootPath,
				ageIdentities: ageIdentities,
				ageRecipients: nil,
			},
		},
	}
	if err := msts.NewSession(); err != nil {
		return fmt.Errorf("CheckIndex: cannot load metadata: %v", err)
	}
	idx := map[string]*dssa.DataEntry{}
	if err := msts.listAll(idx, "/"); err != nil {
		return fmt.Errorf("CheckIndex: metadata inconsistent %v", err)
	}
	for path_, de := range(idx) {
		lgr.Debug("ah", "path", path_, "de", de)
		ap := path.Join(rootPath, common.Id2Path(de.Id))
		ede, err := underlying.Stat(ap)
		if err != nil {
			lgr.Error("CheckIndex: Stat error", "path", path_, "ePath", ap, "err", err)
			if !dryRun {
				msts.Del(path_)
			}
		}
		_ = ede
	}
	return fmt.Errorf("CheckIndex: not yet fully implemented")
}
