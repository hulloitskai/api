package handler

import (
	"encoding/json"

	melody "gopkg.in/olahol/melody.v1"

	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/pkg/zero"
	"go.stevenxie.me/api/service/music"
)

// NowPlayingHandler handles requests for the currently playing track on my
// Spotify account.
func NowPlayingHandler(
	svc music.NowPlayingService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		cplaying, err := svc.NowPlaying()
		if err != nil {
			log.WithError(err).Error("Failed to get currently playing track.")
			return errors.Wrap(err, "getting currently playing track")
		}

		// Send info as JSON.
		return jsonPretty(c, cplaying)
	}
}

// NowPlayingStreamingHandler handles requests for current playing track
// streams.
func NowPlayingStreamingHandler(
	svc music.NowPlayingStreamingService,
	log *logrus.Logger,
) echo.HandlerFunc {
	// Configure Melody.
	var (
		mel        = melody.New()
		serializer nowPlayingStreamSerializer
	)

	connEntry := log.WithField("stage", "connect")
	mel.HandleConnect(func(s *melody.Session) {
		np, err := svc.NowPlaying()
		if err != nil {
			connEntry.
				WithError(err).
				Error("Error getting latest NowPlaying from upstream.")
			return
		}

		message, err := json.Marshal(nowPlayingStreamMessage{
			Event:   npEventNowPlaying,
			Payload: np,
		})
		if err != nil {
			connEntry.
				WithError(err).
				Error("Failed to marshal JSON message.")
			return
		}

		if err = s.Write(message); err != nil {
			connEntry.
				WithError(err).
				Error("Failed to write to socket.")
		}
	})

	broadEntry := log.WithField("stage", "broadcast")
	go func(stream <-chan struct {
		NowPlaying *music.NowPlaying
		Err        error
	}) {
		for value := range stream {
			message, err := serializer.Serialize(value.NowPlaying, value.Err)
			if err != nil {
				broadEntry.
					WithError(err).
					Error("Error while marshalling stream value.")
				continue
			}
			if message == nil {
				continue
			}
			if err = mel.Broadcast(message); err != nil {
				broadEntry.
					WithError(err).
					Error("Failed to broadcast stream object.")
			}
		}
	}(svc.NowPlayingStream())

	handleEntry := log.WithField("stage", "handle")
	return func(c echo.Context) error {
		if err := mel.HandleRequest(c.Response().Writer, c.Request()); err != nil {
			handleEntry.
				WithError(err).
				Error("Melody failed to handle request.")
			return err
		}
		return nil
	}
}

type nowPlayingStreamSerializer struct {
	prevNP  *music.NowPlaying
	prevErr error
}

func (serializer *nowPlayingStreamSerializer) Serialize(
	currNP *music.NowPlaying,
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
