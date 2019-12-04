package graph

type Vertices interface {
	Iterator
	Vertex() Vertex
}

type VertexSlicer interface {
	VertexSlice() []Vertex
}

func VerticesOf(iter Vertices) []Vertex {
	if iter == nil {
		return nil
	}
	length := iter.Len()
	switch {
	case length == 0:
		return nil
	case length < 0:
		panic("graph: called VerticesOf on indeterminate iterator")
	}
	switch iter := iter.(type) {
	case VertexSlicer:
		return iter.VertexSlice()
	}
	vs := make([]Vertex, 0, length)
	for iter.Next() {
		vs = append(vs, iter.Vertex())
	}
	return vs
}
