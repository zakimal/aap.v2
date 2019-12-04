package iterator

import "github.com/zakimal/aap.v2/graph"

type OrderedWeightedEdges struct {
	idx   int
	edges []graph.WeightedEdge
}

func NewOrderedWeightedEdges(edges []graph.WeightedEdge) *OrderedWeightedEdges {
	return &OrderedWeightedEdges{
		idx:   -1,
		edges: edges,
	}
}

func (e *OrderedWeightedEdges) Len() int {
	if e.idx >= len(e.edges) {
		return 0
	}
	if e.idx <= 0 {
		return len(e.edges)
	}
	return len(e.edges[e.idx:])
}

func (e *OrderedWeightedEdges) Next() bool {
	if uint(e.idx)+1 < uint(len(e.edges)) {
		e.idx++
		return true
	}
	e.idx = len(e.edges)
	return false
}

func (e *OrderedWeightedEdges) Reset() {
	e.idx = -1
}

func (e *OrderedWeightedEdges) WeightedEdge() graph.WeightedEdge {
	if e.idx >= len(e.edges) || e.idx < 0 {
		return nil
	}
	return e.edges[e.idx]
}

func (e *OrderedWeightedEdges) WeightedEdgeSlice() []graph.WeightedEdge {
	if e.idx >= len(e.edges) {
		return nil
	}
	idx := e.idx
	if idx == -1 {
		idx = 0
	}
	e.idx = len(e.edges)
	return e.edges[idx:]
}
