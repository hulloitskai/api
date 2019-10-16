# api

_My personal API, now in GraphQL!_

[![Git Tag][tag-img]][tag]
[![Drone][drone-img]][drone]
[![Go Report Card][grp-img]][grp]
[![GoDoc][godoc-img]][godoc]
[![Microbadger][microbadger-img]][microbadger]

> **Why?** Well, why... not?

## GraphQL Endpoint

Since `v2.0.0`, my API is primarily accessed over
[GraphQL](https://graphql.org/).

For example, you can get the name of the song track I'm listening to right now
with the following `curl` query:

```bash
curl \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{ "query": "{ music { current { track { name } } } }" }' \
  https://api.stevenxie.me/v2/graphql
```

_Check out [the interactive API explorer](https://api.stevenxie.me/v2/graphiql)!_

[tag]: https://github.com/stevenxie/api/releases
[tag-img]: https://img.shields.io/github/tag/stevenxie/api.svg
[drone]: https://ci.stevenxie.me/stevenxie/api
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/api/status.svg
[grp]: https://goreportcard.com/report/go.stevenxie.me/api
[grp-img]: https://goreportcard.com/badge/go.stevenxie.me/api
[godoc]: https://godoc.org/go.stevenxie.me/api
[godoc-img]: https://godoc.org/go.stevenxie.me/api?status.svg
[microbadger]: https://microbadger.com/images/stevenxie/api
[microbadger-img]: https://images.microbadger.com/badges/image/stevenxie/api.svg
