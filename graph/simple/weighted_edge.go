package simple

import "github.com/zakimal/aap.v2/graph"

type WeightedEdge struct {
	F, T graph.Vertex
	W    float64
}

func (e WeightedEdge) From() graph.Vertex {
	return e.F
}

func (e WeightedEdge) To() graph.Vertex {
	return e.T
}

func (e WeightedEdge) ReversedEdge() graph.Edge {
	return WeightedEdge{
		F: e.T,
		T: e.F,
		W: e.W,
	}
}

func (e WeightedEdge) Weight() float64 {
	return e.W
}
