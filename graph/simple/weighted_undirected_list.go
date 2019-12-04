package simple

import (
	"fmt"
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/uid"
	"github.com/zakimal/aap.v2/graph/iterator"
)

var (
	wug *WeightedUndirectedGraph

	_ graph.Graph              = wug
	_ graph.Weighted           = wug
	_ graph.Undirected         = wug
	_ graph.WeightedUndirected = wug
	_ graph.VertexAdder        = wug
	_ graph.VertexRemover      = wug
	_ graph.WeightedEdgeAdder  = wug
	_ graph.EdgeRemover        = wug
)

type WeightedUndirectedGraph struct {
	vertices map[int64]graph.Vertex
	edges    map[int64]map[int64]graph.WeightedEdge

	self, absent float64

	vertexIDs uid.IDSet
}

func NewWeightedUndirectedGraph(self, absent float64) *WeightedUndirectedGraph {
	return &WeightedUndirectedGraph{
		vertices:  make(map[int64]graph.Vertex),
		edges:     make(map[int64]map[int64]graph.WeightedEdge),
		self:      self,
		absent:    absent,
		vertexIDs: uid.NewIDSet(),
	}
}

func (g *WeightedUndirectedGraph) Vertex(id int64) graph.Vertex {
	return g.vertices[id]
}

func (g *WeightedUndirectedGraph) Vertices() graph.Vertices {
	if len(g.vertices) == 0 {
		return graph.Empty
	}
	vertices := make([]graph.Vertex, len(g.vertices))
	i := 0
	for _, n := range g.vertices {
		vertices[i] = n
		i++
	}
	return iterator.NewOrderedVertices(vertices)
}

func (g *WeightedUndirectedGraph) Edge(uid, vid int64) graph.Edge {
	return g.WeightedEdgeBetween(uid, vid)
}

func (g *WeightedUndirectedGraph) Edges() graph.Edges {
	if len(g.edges) == 0 {
		return graph.Empty
	}
	var edges []graph.Edge
	seen := make(map[[2]int64]struct{})
	for _, u := range g.edges {
		for _, e := range u {
			uid := e.From().ID()
			vid := e.To().ID()
			if _, ok := seen[[2]int64{uid, vid}]; ok {
				continue
			}
			seen[[2]int64{uid, vid}] = struct{}{}
			seen[[2]int64{vid, uid}] = struct{}{}
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedEdges(edges)
}

func (g *WeightedUndirectedGraph) WeightedEdge(uid, vid int64) graph.WeightedEdge {
	return g.WeightedEdgeBetween(uid, vid)
}

func (g *WeightedUndirectedGraph) WeightedEdges() graph.WeightedEdges {
	var edges []graph.WeightedEdge
	seen := make(map[[2]int64]struct{})
	for _, u := range g.edges {
		for _, e := range u {
			uid := e.From().ID()
			vid := e.To().ID()
			if _, ok := seen[[2]int64{uid, vid}]; ok {
				continue
			}
			seen[[2]int64{uid, vid}] = struct{}{}
			seen[[2]int64{vid, uid}] = struct{}{}
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedWeightedEdges(edges)
}

func (g *WeightedUndirectedGraph) EdgeBetween(xid, yid int64) graph.Edge {
	return g.WeightedEdgeBetween(xid, yid)
}

func (g *WeightedUndirectedGraph) WeightedEdgeBetween(xid, yid int64) graph.WeightedEdge {
	edge, ok := g.edges[xid][yid]
	if !ok {
		return nil
	}
	if edge.From().ID() == xid {
		return edge
	}
	return edge.ReversedEdge().(graph.WeightedEdge)
}

func (g *WeightedUndirectedGraph) From(id int64) graph.Vertices {
	if _, ok := g.vertices[id]; !ok {
		return graph.Empty
	}
	vertices := make([]graph.Vertex, len(g.vertices))
	i := 0
	for from := range g.vertices {
		vertices[i] = g.vertices[from]
		i++
	}
	if len(vertices) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedVertices(vertices)
}

func (g *WeightedUndirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	_, ok := g.edges[xid][yid]
	return ok
}

func (g *WeightedUndirectedGraph) Weight(xid, yid int64) (w float64, ok bool) {
	if xid == yid {
		return g.self, true
	}
	if n, ok := g.edges[xid]; ok {
		if e, ok := n[yid]; ok {
			return e.Weight(), true
		}
	}
	return g.absent, false
}

func (g *WeightedUndirectedGraph) NewVertex() graph.Vertex {
	if len(g.vertices) == 0 {
		return Vertex(0)
	}
	if int64(len(g.vertices)) == uid.MAX {
		panic("simple: cannot allocate vertex: no slot")
	}
	return Vertex(g.vertexIDs.NewID())
}

func (g *WeightedUndirectedGraph) AddVertex(v graph.Vertex) {
	if _, exists := g.vertices[v.ID()]; exists {
		panic(fmt.Sprintf("simple: node ID collision: %d", v.ID()))
	}
	g.vertices[v.ID()] = v
	g.vertexIDs.Use(v.ID())
}

func (g *WeightedUndirectedGraph) RemoveVertex(id int64) {
	if _, ok := g.vertices[id]; !ok {
		return
	}
	delete(g.vertices, id)

	for from := range g.edges[id] {
		delete(g.edges[from], id)
	}
	delete(g.edges, id)

	g.vertexIDs.Release(id)
}

func (g *WeightedUndirectedGraph) NewWeightedEdge(from, to graph.Vertex, weight float64) graph.WeightedEdge {
	return WeightedEdge{F: from, T: to, W: weight}
}

func (g *WeightedUndirectedGraph) AddWeightedEdge(edge graph.WeightedEdge) {
	var (
		from = edge.From()
		fid  = from.ID()
		to   = edge.To()
		tid  = to.ID()
	)

	if fid == tid {
		panic("simple: adding self edge")
	}

	if _, ok := g.vertices[fid]; !ok {
		g.AddVertex(from)
	} else {
		g.vertices[fid] = from
	}
	if _, ok := g.vertices[tid]; !ok {
		g.AddVertex(to)
	} else {
		g.vertices[tid] = to
	}

	if fm, ok := g.edges[fid]; ok {
		fm[tid] = edge
	} else {
		g.edges[fid] = map[int64]graph.WeightedEdge{tid: edge}
	}
	if tm, ok := g.edges[tid]; ok {
		tm[fid] = edge
	} else {
		g.edges[tid] = map[int64]graph.WeightedEdge{fid: edge}
	}
}

func (g *WeightedUndirectedGraph) RemoveEdge(fid, tid int64) {
	if _, ok := g.vertices[fid]; !ok {
		return
	}
	if _, ok := g.vertices[tid]; !ok {
		return
	}

	delete(g.edges[fid], tid)
	delete(g.edges[tid], fid)
}
