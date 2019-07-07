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

// NowPlayingHandler handles requests for the currently playing track on my
// Spotify account.
func NowPlayingHandler(svc api.MusicService, log *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		cplaying, err := svc.NowPlaying()
		if err != nil {
			log.WithError(err).Error("Failed to get currently playing track.")
			return errors.Errorf("getting currently playing track: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, cplaying)
	}
}

// NowPlayingStreamingHandler handles requests for NowPlaying event streams.
func NowPlayingStreamingHandler(
	svc api.MusicStreamingService,
	log *logrus.Logger,
) echo.HandlerFunc {
	// Configure Melody.
	var (
		mel        = melody.New()
		serializer nowPlayingStateSerializer
	)

	connlog := log.WithField("stage", "connect").Logger
	mel.HandleConnect(func(s *melody.Session) {
		np, err := svc.NowPlaying()
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

	go func(stream <-chan struct {
		NowPlaying *api.NowPlaying
		Err        error
	}) {
		broadlog := log.WithField("stage", "broadcast")
		for maybe := range stream {
			message, err := serializer.Serialize(maybe.NowPlaying, maybe.Err)
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
	}(svc.NowPlayingStream())

	handlelog := log.WithField("stage", "handle").Logger
	return func(c echo.Context) error {
		if err := mel.HandleRequest(c.Response().Writer, c.Request()); err != nil {
			handlelog.WithError(err).Error("Melody failed to handle request.")
		}
		return nil
	}
}

type nowPlayingStateSerializer struct {
	prevNP  *api.NowPlaying
	prevErr error
}

func (serializer *nowPlayingStateSerializer) Serialize(
	currNP *api.NowPlaying,
	currErr error,
) (message []byte, err error) {
	var data nowPlayingStreamMessage
	defer func() {
		serializer.prevNP = currNP
		serializer.prevErr = currErr
	}()

	// Fancy state machinery.
	prevNP := serializer.prevNP
	switch {
	case err != nil:
		data.Event = npEventError
		data.Payload = err

	case (prevNP == nil) && (currNP == nil):
		return nil, nil

	case (prevNP == nil && currNP != nil) ||
		(currNP == nil && prevNP != nil) ||
		currNP.Playing != prevNP.Playing ||
		currNP.Track.URL != prevNP.Track.URL:
		data.Event = npEventNowPlaying
		data.Payload = currNP

	case currNP.Playing:
		data.Event = npEventProgress
		data.Payload = currNP.Progress

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
