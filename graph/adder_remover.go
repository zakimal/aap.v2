package graph

type VertexAdder interface {
	NewVertex() Vertex
	AddVertex(vertex Vertex)
}

type VertexRemover interface {
	RemoveVertex(id int64)
}

type EdgeAdder interface {
	NewEdge(from, to Vertex) Edge
	AddEdge(edge Edge)
}

type EdgeRemover interface {
	RemoveEdge(fid, tid int64)
}

type WeightedEdgeAdder interface {
	NewWeightedEdge(from, to Vertex, weight float64) WeightedEdge
	AddWeightedEdge(weightedEdge WeightedEdge)
}
