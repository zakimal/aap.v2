package simple

import (
	"fmt"
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/uid"
	"github.com/zakimal/aap.v2/graph/iterator"
)

var (
	dg *DirectedGraph
	_  graph.Graph         = dg
	_  graph.Directed      = dg
	_  graph.VertexAdder   = dg
	_  graph.VertexRemover = dg
	_  graph.EdgeAdder     = dg
	_  graph.EdgeRemover   = dg
)

// TODO: node adder, edge setter

type DirectedGraph struct {
	vertices map[int64]graph.Vertex
	from     map[int64]map[int64]graph.Edge
	to       map[int64]map[int64]graph.Edge

	vertexIDs uid.IDSet
}

func NewDirectedGraph() *DirectedGraph {
	return &DirectedGraph{
		vertices:  make(map[int64]graph.Vertex),
		from:      make(map[int64]map[int64]graph.Edge),
		to:        make(map[int64]map[int64]graph.Edge),
		vertexIDs: uid.NewIDSet(),
	}
}

func (g *DirectedGraph) Vertex(id int64) graph.Vertex {
	return g.vertices[id]
}

func (g *DirectedGraph) Vertices() graph.Vertices {
	if len(g.vertices) == 0 {
		return graph.Empty
	}
	vertices := make([]graph.Vertex, len(g.vertices))
	i := 0
	for _, v := range g.vertices {
		vertices[i] = v
		i++
	}
	return iterator.NewOrderedVertices(vertices)
}

func (g *DirectedGraph) Edge(uid, vid int64) graph.Edge {
	edge, ok := g.from[uid][vid]
	if !ok {
		return nil
	}
	return edge
}

func (g *DirectedGraph) Edges() graph.Edges {
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

func (g *DirectedGraph) From(id int64) graph.Vertices {
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

func (g *DirectedGraph) To(id int64) graph.Vertices {
	if _, ok := g.to[id]; !ok {
		return graph.Empty
	}
	to := make([]graph.Vertex, len(g.to[id]))
	i := 0
	for uid := range g.to[id] {
		to[i] = g.vertices[uid]
		i++
	}
	if len(to) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedVertices(to)
}

func (g *DirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	if _, ok := g.from[xid][yid]; ok {
		return true
	}
	_, ok := g.from[yid][xid]
	return ok
}

func (g *DirectedGraph) HasEdgeFromTo(uid, vid int64) bool {
	if _, ok := g.from[uid][vid]; !ok {
		return false
	}
	return true
}

func (g *DirectedGraph) NewVertex() graph.Vertex {
	if len(g.vertices) == 0 {
		return Vertex(0)
	}
	if int64(len(g.vertices)) == uid.MAX {
		panic("simple: cannot allocate vertex: no slot")
	}
	return Vertex(g.vertexIDs.NewID())
}

func (g *DirectedGraph) AddVertex(v graph.Vertex) {
	if _, exists := g.vertices[v.ID()]; exists {
		panic(fmt.Sprintf("simple: vertex ID collision: %d", v.ID()))
	}
	g.vertices[v.ID()] = v
	g.vertexIDs.Use(v.ID())
}

func (g *DirectedGraph) RemoveVertex(id int64) {
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

func (g *DirectedGraph) NewEdge(from, to graph.Vertex) graph.Edge {
	return Edge{
		F: from,
		T: to,
	}
}

func (g *DirectedGraph) AddEdge(edge graph.Edge) {
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
		g.from[fid] = map[int64]graph.Edge{tid: edge}
	}

	if tm, ok := g.to[tid]; ok {
		tm[fid] = edge
	} else {
		g.to[tid] = map[int64]graph.Edge{fid: edge}
	}
}

func (g *DirectedGraph) RemoveEdge(fid, tid int64) {
	if _, ok := g.vertices[fid]; !ok {
		return
	}

	if _, ok := g.vertices[tid]; !ok {
		return
	}

	delete(g.from[fid], tid)
	delete(g.to[tid], fid)
}
