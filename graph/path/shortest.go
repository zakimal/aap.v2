package path

import (
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/ordered"
	"github.com/zakimal/aap.v2/graph/internal/set"
	"math"
)

type Shortest struct {
	from             graph.Vertex
	vertices         []graph.Vertex
	indexOf          map[int64]int
	invIndexOf       map[int]int64
	dist             []float64
	next             []int
	hasNegativeCycle bool
}

func NewShortestPathFrom(u graph.Vertex, vertices []graph.Vertex) Shortest {
	indexOf := make(map[int64]int, len(vertices))
	invIndexOf := make(map[int]int64, len(vertices))
	var uid int64
	if u != nil {
		uid = u.ID()
	} else {
		uid = -1
	}
	for i, v := range vertices {
		indexOf[v.ID()] = i
		invIndexOf[i] = v.ID()
		if v.ID() == uid {
			u = v
		}
	}
	shortest := Shortest{
		from:       u,
		vertices:   vertices,
		indexOf:    indexOf,
		invIndexOf: invIndexOf,
		dist:       make([]float64, len(vertices)),
		next:       make([]int, len(vertices)),
	}

	for i := range vertices {
		shortest.dist[i] = math.Inf(1)
		shortest.next[i] = -1
	}
	if uid == 0 {
		shortest.dist[indexOf[uid]] = 0
	}

	return shortest
}

func (s *Shortest) add(u graph.Vertex) int {
	uid := u.ID()
	if _, exists := s.indexOf[uid]; exists {
		panic("shortest: adding existing node")
	}
	idx := len(s.vertices)
	s.indexOf[uid] = idx
	s.invIndexOf[idx] = uid
	s.vertices = append(s.vertices, u)
	s.dist = append(s.dist, math.Inf(1))
	s.next = append(s.next, -1)
	return idx
}

func (s Shortest) set(to int, weight float64, mid int) {
	s.dist[to] = weight
	s.next[to] = mid
}

func (s Shortest) From() graph.Vertex {
	return s.from
}

func (s Shortest) WeightTo(vid int64) float64 {
	to, ok := s.indexOf[vid]
	if !ok {
		return math.Inf(1)
	}
	return s.dist[to]
}

func (s Shortest) WeightToAllVertices() map[int64]float64 {
	result := make(map[int64]float64, len(s.vertices))
	for _, to := range s.vertices {
		result[to.ID()] = s.WeightTo(to.ID())
	}
	return result
}

func (s Shortest) To(vid int64) (path []graph.Vertex, weight float64) {
	to, ok := s.indexOf[vid]
	if !ok || math.IsInf(s.dist[to], 1) {
		return nil, math.Inf(1)
	}
	from := s.indexOf[s.from.ID()]
	path = []graph.Vertex{s.vertices[to]}
	weight = math.Inf(1)
	if s.hasNegativeCycle {
		seen := make(set.IntSet)
		seen.Add(from)
		for to != from {
			if seen.Has(to) {
				weight = math.Inf(-1)
				break
			}
			seen.Add(to)
			path = append(path, s.vertices[s.next[to]])
			to = s.next[to]
		}
	} else {
		n := len(s.vertices)
		for to != from {
			path = append(path, s.vertices[s.next[to]])
			to = s.next[to]
			if n < 0 {
				panic("path: unexpected negative cycle")
			}
			n--
		}
	}
	ordered.Reverse(path)
	return path, math.Min(weight, s.dist[s.indexOf[vid]])
}
