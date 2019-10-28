package transutil

import (
	"github.com/segmentio/fasthash/fnv1a"
	"go.stevenxie.me/api/v2/assist/transit"
)

// HashTransport returns a unique hash number that identifies the
// given transit.Transport.
func HashTransport(t *transit.Transport) uint32 {
	return HashTransportComponents(t.Route, t.Direction, t.Operator.Code)
}

// HashTransportComponents is like HashTransport, but accepts the individual
// components used to compute the hash.
func HashTransportComponents(route, dir, opCode string) uint32 {
	return fnv1a.HashString32(route + dir + opCode)
}
