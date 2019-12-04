package simple

import (
	"fmt"
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/uid"
	"github.com/zakimal/aap.v2/graph/iterator"
)

var (
	wdg *WeightedDirectedGraph

	_ graph.Graph             = wdg
	_ graph.Weighted          = wdg
	_ graph.Directed          = wdg
	_ graph.WeightedDirected  = wdg
	_ graph.VertexAdder       = wdg
	_ graph.VertexRemover     = wdg
	_ graph.WeightedEdgeAdder = wdg
	_ graph.EdgeRemover       = wdg
)

type WeightedDirectedGraph struct {
	vertices map[int64]graph.Vertex
	from     map[int64]map[int64]graph.WeightedEdge
	to       map[int64]map[int64]graph.WeightedEdge

	self, absent float64

	vertexIDs uid.IDSet
}

func NewWeightedDirectedGraph(self, absent float64) *WeightedDirectedGraph {
	return &WeightedDirectedGraph{
		vertices:  make(map[int64]graph.Vertex),
		from:      make(map[int64]map[int64]graph.WeightedEdge),
		to:        make(map[int64]map[int64]graph.WeightedEdge),
		self:      self,
		absent:    absent,
		vertexIDs: uid.NewIDSet(),
	}
}

func (g *WeightedDirectedGraph) Vertex(id int64) graph.Vertex {
	return g.vertices[id]
}

func (g *WeightedDirectedGraph) Vertices() graph.Vertices {
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

func (g *WeightedDirectedGraph) Edge(uid, vid int64) graph.Edge {
	return g.WeightedEdge(uid, vid)
}

func (g *WeightedDirectedGraph) Edges() graph.Edges {
	var edges []graph.Edge
	for _, u := range g.vertices {
		for _, e := range g.from[u.ID()] {
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedEdges(edges)
}

func (g *WeightedDirectedGraph) WeightedEdge(uid, vid int64) graph.WeightedEdge {
	edge, ok := g.from[uid][vid]
	if !ok {
		return nil
	}
	return edge
}

func (g *WeightedDirectedGraph) WeightedEdges() graph.WeightedEdges {
	var edges []graph.WeightedEdge
	for _, u := range g.vertices {
		for _, e := range g.from[u.ID()] {
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedWeightedEdges(edges)
}

func (g *WeightedDirectedGraph) From(id int64) graph.Vertices {
	if _, ok := g.from[id]; !ok {
		return graph.Empty
	}
	from := make([]graph.Vertex, len(g.from[id]))
	i := 0
	for vid := range g.from[id] {
		from[i] = g.vertices[vid]
		i++
	}
	if len(from) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedVertices(from)
}

func (g *WeightedDirectedGraph) To(id int64) graph.Vertices {
	if _, ok := g.to[id]; !ok {
		return graph.Empty
	}
	to := make([]graph.Vertex, len(g.to[id]))
	i := 0
	for vid := range g.to[id] {
		to[i] = g.vertices[vid]
		i++
	}
	if len(to) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedVertices(to)
}

func (g *WeightedDirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	if _, ok := g.from[xid][yid]; ok {
		return true
	}
	_, ok := g.from[yid][xid]
	return ok
}

func (g *WeightedDirectedGraph) HasEdgeFromTo(uid, vid int64) bool {
	if _, ok := g.from[uid][vid]; !ok {
		return false
	}
	return true
}

func (g *WeightedDirectedGraph) Weight(xid, yid int64) (w float64, ok bool) {
	if xid == yid {
		return g.self, true
	}
	if to, ok := g.from[xid]; ok {
		if e, ok := to[yid]; ok {
			return e.Weight(), true
		}
	}
	return g.absent, false
}

func (g *WeightedDirectedGraph) NewVertex() graph.Vertex {
	if len(g.vertices) == 0 {
		return Vertex(0)
	}
	if int64(len(g.vertices)) == uid.MAX {
		panic("simple: cannot allocate vertex: no slot")
	}
	return Vertex(g.vertexIDs.NewID())
}

func (g *WeightedDirectedGraph) AddVertex(v graph.Vertex) {
	if _, exists := g.vertices[v.ID()]; exists {
		panic(fmt.Sprintf("simple: vertex ID collision: %d", v.ID()))
	}
	g.vertices[v.ID()] = v
	g.vertexIDs.Use(v.ID())
}

func (g *WeightedDirectedGraph) RemoveVertex(id int64) {
	if _, ok := g.vertices[id]; !ok {
		return
	}
	delete(g.vertices, id)

	for from := range g.from[id] {
		delete(g.to[from], id)
	}
	delete(g.from, id)

	for to := range g.to[id] {
		delete(g.from[to], id)
	}
	delete(g.to, id)

	g.vertexIDs.Release(id)
}

func (g *WeightedDirectedGraph) NewWeightedEdge(from, to graph.Vertex, weight float64) graph.WeightedEdge {
	return WeightedEdge{F: from, T: to, W: weight}
}

func (g *WeightedDirectedGraph) AddWeightedEdge(edge graph.WeightedEdge) {
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

	if fm, ok := g.from[fid]; ok {
		fm[tid] = edge
	} else {
		g.from[fid] = map[int64]graph.WeightedEdge{tid: edge}
	}
	if tm, ok := g.to[tid]; ok {
		tm[fid] = edge
	} else {
		g.to[tid] = map[int64]graph.WeightedEdge{fid: edge}
	}
}

func (g *WeightedDirectedGraph) RemoveEdge(fid, tid int64) {
	if _, ok := g.vertices[fid]; !ok {
		return
	}
	if _, ok := g.vertices[tid]; !ok {
		return
	}

	delete(g.from[fid], tid)
	delete(g.to[tid], fid)
}
