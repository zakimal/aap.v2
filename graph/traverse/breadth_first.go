package traverse

import (
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/linear"
	"github.com/zakimal/aap.v2/graph/internal/set"
)

type BreadthFirst struct {
	Visit    func(vertex graph.Vertex)
	Traverse func(edge graph.Edge) bool
	queue    linear.VertexQueue
	visited  set.Int64Set
}

func (b *BreadthFirst) Walk(g Graph, from graph.Vertex, until func(v graph.Vertex, depth int) bool) graph.Vertex {
	if b.visited == nil {
		b.visited = make(set.Int64Set)
	}
	b.queue.Enqueue(from)
	if b.Visit != nil && !b.visited.Has(from.ID()) {
		b.Visit(from)
	}
	b.visited.Add(from.ID())

	var (
		depth     int
		children  int
		untilNext = 1
	)

	for b.queue.Len() > 0 {
		t := b.queue.Dequeue()
		if until != nil && until(t, depth) {
			return t
		}
		tid := t.ID()
		to := g.From(tid)
		for to.Next() {
			v := to.Vertex()
			vid := v.ID()
			if b.Traverse != nil && !b.Traverse(g.Edge(tid, vid)) {
				continue
			}
			if b.visited.Has(vid) {
				continue
			}
			if b.Visit != nil {
				b.Visit(v)
			}
			b.visited.Add(vid)
			children++
			b.queue.Enqueue(v)
		}
		if untilNext--; untilNext == 0 {
			depth++
			untilNext = children
			children = 0
		}
	}
	return nil
}

func (b *BreadthFirst) WalkAll(g graph.Undirected, before, after func(), during func(vertex graph.Vertex)) {
	b.Reset()
	nodes := g.Vertices()
	for nodes.Next() {
		from := nodes.Vertex()
		if b.Visited(from) {
			continue
		}
		if before != nil {
			before()
		}
		b.Walk(g, from, func(n graph.Vertex, _ int) bool {
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

func (b *BreadthFirst) Visited(n graph.Vertex) bool {
	return b.visited.Has(n.ID())
}

func (b *BreadthFirst) Reset() {
	b.queue.Reset()
	b.visited = nil
}
