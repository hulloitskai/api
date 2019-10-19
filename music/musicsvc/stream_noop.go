package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
)

// NewNoopCurrentStreamer creates a no-op music.CurrentStreamer.
func NewNoopCurrentStreamer(opts ...svcutil.BasicOption) music.CurrentStreamer {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return noopCurrentStreamer{
		log: logutil.AddComponent(cfg.Logger, (*noopCurrentStreamer)(nil)),
	}
}

type noopCurrentStreamer struct {
	log *logrus.Entry
}

func (stream noopCurrentStreamer) StreamCurrent(
	ctx context.Context,
	_ chan<- music.CurrentlyPlayingResult,
) error {
	stream.log.WithContext(ctx).Info("Currently-playing stream was requested.")
	return nil
}
