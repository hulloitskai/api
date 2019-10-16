package locsvc

import (
	"context"
	"time"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/poll"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"
)

// NewHistoryServicePrecacher creates a new HistoryServicePrecacher.
func NewHistoryServicePrecacher(
	svc location.HistoryService,
	interval time.Duration,
	opts ...svcutil.BasicOption,
) HistoryServicePrecacher {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	log := logutil.AddComponent(cfg.Logger, (*HistoryServicePrecacher)(nil))
	return HistoryServicePrecacher{
		HistoryService: svc,
		pc: poll.NewPrecacher(
			poll.ProdFunc(func() (zero.Interface, error) {
				return svc.RecentHistory(context.Background())
			}),
			interval,
			poll.WithPrecacherLogger(log),
		),
	}
}

// A HistoryServicePrecacher is a location.HistoryService that precaches
// recent location history at a regular interval.
type HistoryServicePrecacher struct {
	location.HistoryService
	pc *poll.Precacher
}

var _ location.HistoryService = (*HistoryServicePrecacher)(nil)

// RecentHistory implements location.HistoryService.RecentHistory.
func (hsp HistoryServicePrecacher) RecentHistory(context.Context) (
	[]location.HistorySegment, error) {
	v, err := hsp.pc.Results()
	if err != nil {
		return nil, err
	}
	if segs, ok := v.([]location.HistorySegment); ok {
		return segs, nil
	}
	return nil, nil
}

// Stop stops the cacher from requesting new values.
func (hsp HistoryServicePrecacher) Stop() { hsp.pc.Stop() }
