package traverse

import "github.com/zakimal/aap.v2/graph"

var _ Graph = graph.Graph(nil)

type Graph interface {
	From(id int64) graph.Vertices
	Edge(uid, vid int64) graph.Edge
}
