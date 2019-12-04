package path

import (
	"github.com/zakimal/aap.v2/graph/traverse"
	"math"
)

type Weighted interface {
	Weight(xid, yid int64) (w float64, ok bool)
}

type Weighting func(xid, yid int64) (w float64, ok bool)

func UniformCost(g traverse.Graph) Weighting {
	return func(xid, yid int64) (w float64, ok bool) {
		if xid == yid {
			return 0, true
		}
		if e := g.Edge(xid, yid); e != nil {
			return 1, true
		}
		return math.Inf(1), false
	}
}
