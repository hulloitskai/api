package locgql

import (
	"go.stevenxie.me/api/location"
)

func convertHistorySegments(segs []location.HistorySegment) []HistorySegment {
	segments := make([]HistorySegment, len(segs))
	for i := range segs {
		var (
			seg = &segs[i]
			s   = HistorySegment{HistorySegment: &segs[i]}
		)
		if a := seg.Address; a != "" {
			s.Address = &a
		}
		if d := seg.Distance; d != 0 {
			s.Distance = &d
		}
		segments[i] = s
	}
	return segments
}

func convertAddress(a *location.Address) Address {
	addr := Address{Address: a}
	if d := a.District; d != "" {
		addr.District = &d
	}
	if s := a.Street; s != "" {
		addr.Street = &s
	}
	if n := a.Number; n != "" {
		addr.Number = &n
	}
	return addr
}

func convertPlace(p *location.Place) Place {
	return Place{
		Place:   p,
		Address: convertAddress(&p.Address),
	}
}
