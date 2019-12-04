package set

import "github.com/zakimal/aap.v2/graph"

type VertexSet map[int64]graph.Vertex

func NewVertexSet() VertexSet {
	return make(VertexSet)
}

func NewVertexSetWithSize(size int) VertexSet {
	return make(VertexSet, size)
}

func (s VertexSet) Add(v graph.Vertex) {
	s[v.ID()] = v
}

func (s VertexSet) Has(v graph.Vertex) bool {
	_, ok := s[v.ID()]
	return ok
}

func (s VertexSet) Remove(v graph.Vertex) {
	delete(s, v.ID())
}

func (s VertexSet) Count() int {
	return len(s)
}

func VertexSetEqual(a, b VertexSet) bool {
	if vertexSetSame(a, b) {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for e := range a {
		if _, ok := b[e]; !ok {
			return false
		}
	}
	return true
}

func Clone(src VertexSet) VertexSet {
	dst := make(VertexSet, len(src))
	for e, v := range src {
		dst[e] = v
	}
	return dst
}

func Union(a, b VertexSet) VertexSet {
	if vertexSetSame(a, b) {
		return Clone(a)
	}
	dst := make(VertexSet)
	for e, v := range a {
		dst[e] = v
	}
	for e, v := range b {
		dst[e] = v
	}
	return dst
}

func Intersection(a, b VertexSet) VertexSet {
	if vertexSetSame(a, b) {
		return Clone(a)
	}
	dst := make(VertexSet)
	if len(a) > len(b) {
		a, b = b, a
	}
	for e, v := range a {
		if _, ok := b[e]; ok {
			dst[e] = v
		}
	}
	return dst
}
