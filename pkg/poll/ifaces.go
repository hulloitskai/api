package poll

import "go.stevenxie.me/gopkg/zero"

type (
	// An Actor can produce and receive arbitrary values.
	Actor interface {
		Producer

		// Recv is called synchronously in order to receive a value.
		// It does not have to be concurrent-safe.
		Recv(zero.Interface, error)
	}

	// A Producer can produce a value in a concurrent-safe manner.
	Producer interface {
		Prod() (zero.Interface, error)
	}

	// A ProdFunc is a function that trivially implements the Producer interface.
	ProdFunc func() (zero.Interface, error)
)

var _ Producer = (*ProdFunc)(nil)

// Prod trivially implements the Producer interface for a ProdFunc.
func (pf ProdFunc) Prod() (zero.Interface, error) { return pf() }
