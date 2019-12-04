package traverse

import (
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/linear"
	"github.com/zakimal/aap.v2/graph/internal/set"
)

type DepthFirst struct {
	Visit    func(vertex graph.Vertex)
	Traverse func(edge graph.Edge) bool
	stack    linear.VertexStack
	visited  set.Int64Set
}

func (d *DepthFirst) Walk(g Graph, from graph.Vertex, until func(vertex graph.Vertex) bool) graph.Vertex {
	if d.visited == nil {
		d.visited = make(set.Int64Set)
	}
	d.stack.Push(from)
	if d.Visit != nil && !d.visited.Has(from.ID()) {
		d.Visit(from)
	}
	d.visited.Add(from.ID())

	for d.stack.Len() > 0 {
		t := d.stack.Pop()
		if until != nil && until(t) {
			return t
		}
		tid := t.ID()
		to := g.From(tid)
		for to.Next() {
			n := to.Vertex()
			nid := n.ID()
			if d.Traverse != nil && !d.Traverse(g.Edge(tid, nid)) {
				continue
			}
			if d.visited.Has(nid) {
				continue
			}
			if d.Visit != nil {
				d.Visit(n)
			}
			d.visited.Add(nid)
			d.stack.Push(n)
		}
	}
	return nil
}

func (d *DepthFirst) WalkAll(g graph.Undirected, before, after func(), during func(vertex graph.Vertex)) {
	d.Reset()
	nodes := g.Vertices()
	for nodes.Next() {
		from := nodes.Vertex()
		if d.Visited(from) {
			continue
		}
		if before != nil {
			before()
		}
		d.Walk(g, from, func(n graph.Vertex) bool {
			if during != nil {
				during(n)
			}
			return false
		})
		if after != nil {
			after()
		}
	}
}

func (d *DepthFirst) Visited(n graph.Vertex) bool {
	return d.visited.Has(n.ID())
}

func (d *DepthFirst) Reset() {
	d.stack = d.stack[:0]
	d.visited = nil
}
