# api

_My personal API, now in GraphQL!_

[![Git Tag][tag-img]][tag]
[![Drone][drone-img]][drone]
[![Go Report Card][grp-img]][grp]
[![GoDoc][godoc-img]][godoc]
[![License][license-img]][license]

> **Why?** Well, why... not?

## GraphQL Endpoint

Since `v2.0.0`, my API is primarily accessed over
[GraphQL](https://graphql.org/).

For example, you can get the name of the song I'm listening to right now with
the following `curl` query:

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

- [**When's the next bus?**](https://www.icloud.com/shortcuts/6e7f65dc26f1439c9411cd0dd9b2b633)

  This shortcut asks for the transit route you want to take (i.e. "the GRT 7",
  "the 19B", "the GO 25"), and computes the departure times for that route
  at the stop closest to you. Uses realtime transit data when available.

- [**When's the next ION?**](https://www.icloud.com/shortcuts/d54537c3bb2e4ebe82ebaf79be58fcf7)

  Like the above shortcut, but hard-codes the route to the
  [GRT 301 ION](https://www.grt.ca/en/ion-light-rail.aspx).

- [**Find nearby buses**](https://www.icloud.com/shortcuts/2c132eaf2761424289a797234beb833d)

  Get a list of nearby transports (transit routes and direction of travel).

[tag]: https://github.com/stevenxie/api/releases
[tag-img]: https://img.shields.io/github/tag/stevenxie/api.svg
[drone]: https://ci.stevenxie.me/stevenxie/api
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/api/status.svg
[grp]: https://goreportcard.com/report/go.stevenxie.me/api
[grp-img]: https://goreportcard.com/badge/go.stevenxie.me/api
[godoc]: https://godoc.org/go.stevenxie.me/api
[godoc-img]: https://godoc.org/go.stevenxie.me/api?status.svg
[license]: https://creativecommons.org/licenses/by/4.0/
[license-img]: https://img.shields.io/github/license/stevenxie/api
