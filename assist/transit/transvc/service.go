package transvc

import (
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
)

// NewService creates a new transit.Service.
func NewService(
	loc transit.LocatorService,
	rts transit.RealTimeService,
	opts ...svcutil.BasicOption,
) transit.Service {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return service{
		loc: loc,
		rts: rts,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	loc transit.LocatorService
	rts transit.RealTimeService
	log *logrus.Entry
}

var _ transit.Service = (*service)(nil)
