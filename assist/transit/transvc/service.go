package transvc

import (
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/pkg/basic"
	"go.stevenxie.me/gopkg/logutil"
)

// NewService creates a new transit.Service.
func NewService(
	loc transit.LocatorService,
	rts transit.RealtimeSource,
	opts ...basic.Option,
) transit.Service {
	cfg := basic.BuildConfig(opts...)
	return service{
		loc: loc,
		rts: rts,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	loc transit.LocatorService
	rts transit.RealtimeSource
	log *logrus.Entry
}

var _ transit.Service = (*service)(nil)
