# api

_A personal API._

[![Git Tag][tag-img]][tag]
[![Drone][drone-img]][drone]
[![Go Report Card][grp-img]][grp]
[![GoDoc][godoc-img]][godoc]
[![Microbadger][microbadger-img]][microbadger]

> **Why?** Well, why... not?

## REST Endpoints

| Path                                                           | Description                                  |
| -------------------------------------------------------------- | -------------------------------------------- |
| [`/v1/`](https://api.stevenxie.me/v1/)                         | API server metadata.                         |
| [`/v1/about`](https://api.stevenxie.me/v1/about)               | General personal information.                |
| [`/v1/commits`](https://api.stevenxie.me/v1/commits)           | Recent commits from GitHub.                  |
| [`/v1/nowplaying`](https://api.stevenxie.me/v1/nowplaying)     | Currently playing track from Spotify.        |
| [`/v1/productivity`](https://api.stevenxie.me/v1/productivity) | Productivity metrics from RescueTime.        |
| [`/v1/availability`](https://api.stevenxie.me/v1/availability) | Personal availability information from GCal. |

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
