package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/basic"
)

// NewNoopCurrentStreamer creates a no-op music.CurrentStreamer.
func NewNoopCurrentStreamer(opts ...basic.Option) music.CurrentStreamer {
	cfg := basic.BuildConfig(opts...)
	return noopCurrentStreamer{
		log: logutil.WithComponent(cfg.Logger, (*noopCurrentStreamer)(nil)),
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
