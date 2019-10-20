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

## Transit Shortcuts

There are several transit-related
[Siri Shortcuts](https://support.apple.com/en-ca/guide/shortcuts/welcome/ios)
that can be used with my API.

I really wanted to be able to get the next bus departures without taking my
phone out of my pocket (using Siri-enabled earbuds) in the mornings, so I
built some shortcuts that let me do that:

- [**When's the next bus?**](https://www.icloud.com/shortcuts/c1de939a8bb943a69cbca6ddc07ba7a6)

  This shortcut asks for the transit route you want to take (i.e. "The 7", "The 19B", "The GO 25"), and computes the departure times for that route
  at the stop closest to you. Uses realtime transit data when available.

- [**When's the next ION?**](https://www.icloud.com/shortcuts/51513a4a845545faa24c53aad7f418cc)

  Like the above shortcut, but hard-codes the route to the
  [GRT 301 ION](https://www.grt.ca/en/ion-light-rail.aspx).

- [**Find nearby buses**](https://www.icloud.com/shortcuts/579230ca27fb4b14bd2ffbcf26b6244f)

  Get a list of nearby transports (transit routes and direction of travel).

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
