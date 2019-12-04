package simple

import "github.com/zakimal/aap.v2/graph"

type Vertex int64

func (v Vertex) ID() int64 {
	return int64(v)
}

func newVertex(id int) graph.Vertex {
	return Vertex(id)
}
