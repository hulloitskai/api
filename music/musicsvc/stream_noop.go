package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/pkg/basic"
)

// NewNoopCurrentStreamer creates a no-op music.CurrentStreamer.
func NewNoopCurrentStreamer(opts ...basic.Option) music.CurrentStreamer {
	opt := basic.BuildOptions(opts...)
	return noopCurrentStreamer{
		log: logutil.WithComponent(opt.Logger, (*noopCurrentStreamer)(nil)),
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
