package graph

// Undirect converts a directed graph.
type Undirect struct {
	G Directed
}

var _ Undirected = Undirect{}

func (g Undirect) Vertex(id int64) Vertex {
	return g.G.Vertex(id)
}

func (g Undirect) Vertices() Vertices {
	return g.G.Vertices()
}

func (g Undirect) From(uid int64) Vertices {
	if g.G.Vertex(uid) == nil {
		return Empty
	}
	return newVertexIteratorPair(g.G.From(uid), g.G.To(uid))
}

func (g Undirect) HasEdgeBetween(xid, yid int64) bool {
	return g.G.HasEdgeBetween(xid, yid)
}

func (g Undirect) Edge(uid, vid int64) Edge {
	return g.EdgeBetween(uid, vid)
}

func (g Undirect) EdgeBetween(xid, yid int64) Edge {
	fe := g.G.Edge(xid, yid)
	re := g.G.Edge(yid, xid)
	if fe == nil && re == nil {
		return nil
	}
	return EdgePair{fe, re}
}

type UndirectWeighted struct {
	G      WeightedDirected
	Absent float64
	Merge  func(x, y float64, xe, ye Edge) float64
}

var (
	_ Undirected         = UndirectWeighted{}
	_ WeightedUndirected = UndirectWeighted{}
)

func (g UndirectWeighted) Vertex(id int64) Vertex {
	return g.G.Vertex(id)
}

func (g UndirectWeighted) Vertices() Vertices {
	return g.G.Vertices()
}

func (g UndirectWeighted) From(uid int64) Vertices {
	if g.G.Vertex(uid) == nil {
		return Empty
	}
	return newVertexIteratorPair(g.G.From(uid), g.G.To(uid))
}

func (g UndirectWeighted) HasEdgeBetween(xid, yid int64) bool {
	return g.G.HasEdgeBetween(xid, yid)
}

func (g UndirectWeighted) Edge(uid, vid int64) Edge {
	return g.WeightedEdgeBetween(uid, vid)
}

func (g UndirectWeighted) WeightedEdge(uid, vid int64) WeightedEdge {
	return g.WeightedEdgeBetween(uid, vid)
}

func (g UndirectWeighted) EdgeBetween(xid, yid int64) Edge {
	return g.WeightedEdgeBetween(xid, yid)
}

func (g UndirectWeighted) WeightedEdgeBetween(xid, yid int64) WeightedEdge {
	fe := g.G.Edge(xid, yid)
	re := g.G.Edge(yid, xid)
	if fe == nil && re == nil {
		return nil
	}

	f, ok := g.G.Weight(xid, yid)
	if !ok {
		f = g.Absent
	}
	r, ok := g.G.Weight(yid, xid)
	if !ok {
		r = g.Absent
	}

	var w float64
	if g.Merge == nil {
		w = (f + r) / 2
	} else {
		w = g.Merge(f, r, fe, re)
	}
	return WeightedEdgePair{EdgePair: [2]Edge{fe, re}, W: w}
}

func (g UndirectWeighted) Weight(xid, yid int64) (w float64, ok bool) {
	fe := g.G.Edge(xid, yid)
	re := g.G.Edge(yid, xid)

	f, fOk := g.G.Weight(xid, yid)
	if !fOk {
		f = g.Absent
	}
	r, rOK := g.G.Weight(yid, xid)
	if !rOK {
		r = g.Absent
	}
	ok = fOk || rOK

	if g.Merge == nil {
		return (f + r) / 2, ok
	}
	return g.Merge(f, r, fe, re), ok
}

// EdgePair is an opposed pair of directed edges.
type EdgePair [2]Edge

func (e EdgePair) From() Vertex {
	if e[0] != nil {
		return e[0].From()
	} else if e[1] != nil {
		return e[1].From()
	}
	return nil
}

func (e EdgePair) To() Vertex {
	if e[0] != nil {
		return e[0].To()
	} else if e[1] != nil {
		return e[1].To()
	}
	return nil
}

func (e EdgePair) ReversedEdge() Edge {
	if e[0] != nil {
		e[0] = e[0].ReversedEdge()
	}
	if e[1] != nil {
		e[1] = e[1].ReversedEdge()
	}
	return e
}

type WeightedEdgePair struct {
	EdgePair
	W float64
}

func (e WeightedEdgePair) ReversedEdge() Edge {
	e.EdgePair = e.EdgePair.ReversedEdge().(EdgePair)
	return e
}

func (e WeightedEdgePair) Weight() float64 {
	return e.W
}

// vertexIteratorPair combines two Vertices to produce a single stream of unique vertices
type vertexIteratorPair struct {
	a, b      Vertices
	curr, cnt int

	// unique indicate the vertex in b with the key ID is unique.
	unique map[int64]bool
}

func newVertexIteratorPair(a, b Vertices) *vertexIteratorPair {
	vs := vertexIteratorPair{
		a:      a,
		b:      b,
		curr:   0,
		cnt:    0,
		unique: make(map[int64]bool),
	}
	vs.b.Reset()
	for vs.a.Next() {
		if _, ok := vs.unique[vs.a.Vertex().ID()]; !ok {
			vs.cnt++
		}
		vs.unique[vs.a.Vertex().ID()] = false
	}
	vs.a.Reset()
	return &vs
}

func (v *vertexIteratorPair) Len() int {
	return v.cnt - v.curr
}

func (v *vertexIteratorPair) Next() bool {
	if v.a.Next() {
		v.curr++
		return true
	}
	for v.b.Next() {
		if v.unique[v.b.Vertex().ID()] {
			v.curr++
			return true
		}
	}
	return false
}

func (v *vertexIteratorPair) Vertex() Vertex {
	if v.a.Len() != 0 {
		return v.a.Vertex()
	}
	return v.b.Vertex()
}

func (v *vertexIteratorPair) Reset() {
	v.curr = 0
	v.a.Reset()
	v.b.Reset()
}
