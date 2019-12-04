package iterator

import "github.com/zakimal/aap.v2/graph"

type OrderedEdges struct {
	idx   int
	edges []graph.Edge
}

func NewOrderedEdges(edges []graph.Edge) *OrderedEdges {
	return &OrderedEdges{
		idx:   -1,
		edges: edges,
	}
}

func (e *OrderedEdges) Len() int {
	if e.idx >= len(e.edges) {
		return 0
	}
	if e.idx <= 0 {
		return len(e.edges)
	}
	return len(e.edges[e.idx:])
}

func (e *OrderedEdges) Next() bool {
	if uint(e.idx)+1 < uint(len(e.edges)) {
		e.idx++
		return true
	}
	e.idx = len(e.edges)
	return false
}

func (e *OrderedEdges) Edge() graph.Edge {
	if e.idx >= len(e.edges) || e.idx < 0 {
		return nil
	}
	return e.edges[e.idx]
}

func (e *OrderedEdges) Reset() {
	e.idx = -1
}

func (e *OrderedEdges) EdgeSlice() []graph.Edge {
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
