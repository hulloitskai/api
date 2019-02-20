package mongo

import errors "golang.org/x/xerrors"

// ErrInvalidID is returned when an hex ID could not be converted into an
// ObjectID.
var ErrInvalidID = errors.New("mongo: id is not a valid ObjectID")
