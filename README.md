# api

_A personal API._

[![Git Tag][tag-img]][tag]
[![Drone][drone-img]][drone]
[![Go Report Card][grp-img]][grp]
[![GoDoc][godoc-img]][godoc]
[![Microbadger][microbadger-img]][microbadger]

> **Why?** Well, why... not?

## REST Endpoints

| Path                                                                   | Description                                  |
| ---------------------------------------------------------------------- | -------------------------------------------- |
| [`/v1/`](https://api.stevenxie.me/v1/)                                 | API server metadata.                         |
| [`/v1/about`](https://api.stevenxie.me/v1/about)                       | General personal information.                |
| [`/v1/commits`](https://api.stevenxie.me/v1/commits)                   | Recent commits from GitHub.                  |
| [`/v1/nowplaying`](https://api.stevenxie.me/v1/nowplaying)             | Currently playing track from Spotify.        |
| [`/v1/productivity`](https://api.stevenxie.me/v1/productivity)         | Productivity metrics from RescueTime.        |
| [`/v1/availability`](https://api.stevenxie.me/v1/availability)         | Personal availability information from GCal. |
| [`/v1/location`](https://api.stevenxie.me/v1/location)                 | Current location data from Google Maps.      |
| [`/v1/location/history`](https://api.stevenxie.me/v1/location/history) | Recent location history from Google Maps.    |

## Websocket Endpoints

### Now Playing

**Endpoint:** `/v1/nowplaying/ws`

All messages from this endpoint are in the format:

```js
{
  "event": String,
  "payload": (String | Object | Number)
}
```

| Event        | Payload Type | Payload Description                                                           |
| ------------ | ------------ | ----------------------------------------------------------------------------- |
| `error`      | `String`     | A description of an error from the server.                                    |
| `nowplaying` | `Object`     | A full `NowPlaying` object, which describes a track that's currently playing. |
| `progress`   | `Number`     | The progress of the currently playing track, in milliseconds.                 |

Upon socket connection, a `nowplaying` event will be sent; afterwards,
`progress` events will be broadcasted until either the playback stops, or the
track changes, in which case a new `nowplaying` event will be sent.

[tag]: https://github.com/stevenxie/api/releases
[tag-img]: https://img.shields.io/github/tag/stevenxie/api.svg
[drone]: https://ci.stevenxie.me/stevenxie/api
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/api/status.svg
[grp]: https://goreportcard.com/report/github.com/stevenxie/api
[grp-img]: https://goreportcard.com/badge/github.com/stevenxie/api
[godoc]: https://godoc.org/github.com/stevenxie/api
[godoc-img]: https://godoc.org/github.com/stevenxie/api?status.svg
[microbadger]: https://microbadger.com/images/stevenxie/api
[microbadger-img]: https://images.microbadger.com/badges/image/stevenxie/api.svg
