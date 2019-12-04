package graph

type Vertex interface {
	ID() int64
}

type Edge interface {
	From() Vertex
	To() Vertex
	ReversedEdge() Edge
}

type WeightedEdge interface {
	Edge
	Weight() float64
}

type Graph interface {
	Vertex(id int64) Vertex
	Vertices() Vertices
	From(id int64) Vertices
	HasEdgeBetween(xid, yid int64) bool
	Edge(uid, vid int64) Edge
}

type Weighted interface {
	Graph
	WeightedEdge(uid, vid int64) WeightedEdge
	Weight(xid, yid int64) (w float64, ok bool)
}

type Undirected interface {
	Graph
	EdgeBetween(xid, yid int64) Edge
}

type WeightedUndirected interface {
	Weighted
	WeightedEdgeBetween(xid, yid int64) WeightedEdge
}

type Directed interface {
	Graph
	HasEdgeFromTo(uid, vid int64) bool
	To(id int64) Vertices
}

type WeightedDirected interface {
	Weighted
	HasEdgeFromTo(uid, vid int64) bool
	To(id int64) Vertices
}
