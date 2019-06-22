package handler

import (
	"encoding/json"

	errors "golang.org/x/xerrors"
	"gopkg.in/olahol/melody.v1"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

// NowPlayingProvider provides handlers relating to music.
type NowPlayingProvider struct {
	svc      api.NowPlayingService
	streamer api.NowPlayingStreamingService
}

// NewNowPlayingProvider creates a new NowPlayingProvider.
func NewNowPlayingProvider(
	svc api.NowPlayingService,
	streamer api.NowPlayingStreamingService,
) *NowPlayingProvider {
	return &NowPlayingProvider{svc: svc, streamer: streamer}
}

// RESTHandler handles GET requests for the currently playing track on my
// Spotify account.
func (p *NowPlayingProvider) RESTHandler(log *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		cplaying, err := p.svc.NowPlaying()
		if err != nil {
			log.WithError(err).Error("Failed to get currently playing track.")
			return errors.Errorf("getting currently playing track: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, cplaying)
	}
}

// StreamingHandler handles requests for NowPlaying event streams.
func (p *NowPlayingProvider) StreamingHandler(
	log *logrus.Logger,
) echo.HandlerFunc {
	// Configure Melody.
	var (
		mel        = melody.New()
		serializer nowPlayingStateSerializer
	)

	connlog := log.WithField("stage", "connect").Logger
	mel.HandleConnect(func(s *melody.Session) {
		np, err := p.svc.NowPlaying()
		if err != nil {
			connlog.
				WithError(err).
				Error("Error getting latest NowPlaying from upstream.")
			return
		}

		data, err := json.Marshal(nowPlayingStreamMessage{
			Event:   npEventNowPlaying,
			Payload: np,
		})
		if err != nil {
			log.WithError(err).Error("Failed to marshal JSON message.")
		}

		if err := s.Write(data); err != nil {
			log.WithError(err).Error("Failed to write to socket.")
		}
	})

	go func(stream <-chan api.MaybeNowPlaying) {
		broadlog := log.WithField("stage", "broadcast")
		for maybe := range stream {
			message, err := serializer.SerializeMaybe(maybe)
			if err != nil {
				broadlog.
					WithError(err).
					Error("Error while marshalling stream response.")
				continue
			}

			if message == nil {
				continue
			}
			if err = mel.Broadcast(message); err != nil {
				broadlog.WithError(err).Error("Failed to broadcast stream object.")
				continue
			}
		}
	}(p.streamer.NowPlayingStream())

	handlelog := log.WithField("stage", "handle").Logger
	return func(c echo.Context) error {
		if err := mel.HandleRequest(c.Response().Writer, c.Request()); err != nil {
			handlelog.WithError(err).Error("Melody failed to handle request.")
		}
		return nil
	}
}

type nowPlayingStateSerializer struct{ prev api.MaybeNowPlaying }

func (serializer *nowPlayingStateSerializer) SerializeMaybe(
	maybe api.MaybeNowPlaying,
) (message []byte, err error) {
	var (
		data nowPlayingStreamMessage
		curr = maybe.NowPlaying
		prev = serializer.prev.NowPlaying
	)

	defer func() { serializer.prev = maybe }()

	// Fancy state machinery.
	switch {
	case maybe.IsError():
		data.Event = npEventError
		data.Payload = maybe.Err

	case (prev == nil) && (curr == nil):
		return nil, nil

	case (prev == nil && curr != nil) ||
		(curr == nil && prev != nil) ||
		curr.Playing != prev.Playing ||
		curr.Track.URL != prev.Track.URL:
		data.Event = npEventNowPlaying
		data.Payload = maybe.NowPlaying

	case curr.Playing:
		data.Event = npEventProgress
		data.Payload = curr.Progress

	default:
		return nil, nil
	}

	return json.Marshal(&data)
}

type nowPlayingStreamMessage struct {
	Event   string         `json:"event"`
	Payload zero.Interface `json:"payload"`
}

const (
	npEventError      = "error"
	npEventNowPlaying = "nowplaying"
	npEventProgress   = "progress"
)
