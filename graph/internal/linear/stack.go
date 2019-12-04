package linear

import "github.com/zakimal/aap.v2/graph"

type VertexStack []graph.Vertex

func (s *VertexStack) Len() int {
	return len(*s)
}

func (s *VertexStack) Pop() graph.Vertex {
	vs := *s
	vs, v := vs[:len(vs)-1], vs[len(vs)-1]
	*s = vs
	return v
}

func (s *VertexStack) Push(v graph.Vertex) {
	*s = append(*s, v)
}
