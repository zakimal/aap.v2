package iterator

import "github.com/zakimal/aap.v2/graph"

type OrderedVertices struct {
	idx      int
	vertices []graph.Vertex
}

func NewOrderedVertices(vertices []graph.Vertex) *OrderedVertices {
	return &OrderedVertices{
		idx:      -1,
		vertices: vertices,
	}
}

func (v *OrderedVertices) Len() int {
	if v.idx >= len(v.vertices) {
		return 0
	}
	if v.idx <= 0 {
		return len(v.vertices)
	}
	return len(v.vertices[v.idx:])
}

func (v *OrderedVertices) Next() bool {
	if uint(v.idx)+1 < uint(len(v.vertices)) {
		v.idx++
		return true
	}
	v.idx = len(v.vertices)
	return false
}

func (v *OrderedVertices) Vertex() graph.Vertex {
	if v.idx >= len(v.vertices) || v.idx < 0 {
		return nil
	}
	return v.vertices[v.idx]
}

func (v *OrderedVertices) VertexSlice() []graph.Vertex {
	if v.idx >= len(v.vertices) {
		return nil
	}
	idx := v.idx
	if idx == -1 {
		idx = 0
	}
	v.idx = len(v.vertices)
	return v.vertices[idx:]
}

func (v *OrderedVertices) Reset() {
	v.idx = -1
}
