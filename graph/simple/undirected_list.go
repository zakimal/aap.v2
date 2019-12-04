package simple

import (
	"fmt"
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/uid"
	"github.com/zakimal/aap.v2/graph/iterator"
)

var (
	ug *UndirectedGraph

	_ graph.Graph         = ug
	_ graph.Undirected    = ug
	_ graph.VertexAdder   = ug
	_ graph.VertexRemover = ug
	_ graph.EdgeAdder     = ug
	_ graph.EdgeRemover   = ug
)

type UndirectedGraph struct {
	vertices map[int64]graph.Vertex
	edges    map[int64]map[int64]graph.Edge

	vertexIDs uid.IDSet
}

func NewUndirectedGraph() *UndirectedGraph {
	return &UndirectedGraph{
		vertices:  make(map[int64]graph.Vertex),
		edges:     make(map[int64]map[int64]graph.Edge),
		vertexIDs: uid.NewIDSet(),
	}
}

func (g *UndirectedGraph) Vertex(id int64) graph.Vertex {
	return g.vertices[id]
}

func (g *UndirectedGraph) Vertices() graph.Vertices {
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

func (g *UndirectedGraph) Edge(uid, vid int64) graph.Edge {
	return g.EdgeBetween(uid, vid)
}

func (g *UndirectedGraph) EdgeBetween(xid, yid int64) graph.Edge {
	edge, ok := g.edges[xid][yid]
	if !ok {
		return nil
	}
	if edge.From().ID() == xid {
		return edge
	}
	return edge.ReversedEdge()
}

func (g *UndirectedGraph) Edges() graph.Edges {
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

func (g *UndirectedGraph) From(id int64) graph.Vertices {
	if _, ok := g.vertices[id]; !ok {
		return graph.Empty
	}

	nodes := make([]graph.Vertex, len(g.edges[id]))
	i := 0
	for from := range g.edges[id] {
		nodes[i] = g.vertices[from]
		i++
	}
	if len(nodes) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedVertices(nodes)
}

func (g *UndirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	_, ok := g.edges[xid][yid]
	return ok
}

func (g *UndirectedGraph) NewVertex() graph.Vertex {
	if len(g.vertices) == 0 {
		return Vertex(0)
	}
	if int64(len(g.vertices)) == uid.MAX {
		panic("simple: cannot allocate vertex: no slot")
	}
	return Vertex(g.vertexIDs.NewID())
}

func (g *UndirectedGraph) AddVertex(v graph.Vertex) {
	if _, exists := g.vertices[v.ID()]; exists {
		panic(fmt.Sprintf("simple: vertex ID collision: %d", v.ID()))
	}
	g.vertices[v.ID()] = v
	g.vertexIDs.Use(v.ID())
}

func (g *UndirectedGraph) RemoveVertex(id int64) {
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

func (g *UndirectedGraph) NewEdge(from, to graph.Vertex) graph.Edge {
	return Edge{
		F: from,
		T: to,
	}
}

func (g *UndirectedGraph) AddEdge(edge graph.Edge) {
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
		g.edges[fid] = map[int64]graph.Edge{tid: edge}
	}
	if tm, ok := g.edges[tid]; ok {
		tm[fid] = edge
	} else {
		g.edges[tid] = map[int64]graph.Edge{fid: edge}
	}
}

func (g *UndirectedGraph) RemoveEdge(fid, tid int64) {
	if _, ok := g.vertices[fid]; !ok {
		return
	}
	if _, ok := g.vertices[tid]; !ok {
		return
	}

	delete(g.edges[fid], tid)
	delete(g.edges[tid], fid)
}
