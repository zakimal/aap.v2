package graph

type Builder interface {
	VertexAdder
	EdgeAdder
}

type DirectedBuilder interface {
	Builder
	Directed
}

type UndirectedBuilder interface {
	Builder
	Undirected
}

type WeightedBuilder interface {
	WeightedEdgeAdder
	VertexAdder
}

type DirectedWeightedBuilder interface {
	WeightedBuilder
	Directed
}

type UndirectedWeightedBuilder interface {
	Undirected
	WeightedBuilder
}
