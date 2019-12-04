package simple

import (
	"github.com/zakimal/aap.v2/graph"
)

type Edge struct {
	F, T graph.Vertex
}

func (e Edge) From() graph.Vertex {
	return e.F
}

func (e Edge) To() graph.Vertex {
	return e.T
}

func (e Edge) ReversedEdge() graph.Edge {
	return Edge{
		F: e.T,
		T: e.F,
	}
}
