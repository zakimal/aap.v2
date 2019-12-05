package path

import (
	"container/heap"
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/traverse"
)

func DijkstraFrom(u graph.Vertex, g traverse.Graph) Shortest {
	var shortest Shortest
	if h, ok := g.(graph.Graph); ok {
		if h.Vertex(u.ID()) == nil {
			return Shortest{from: u}
		}
		shortest = NewShortestPathFrom(u, graph.VerticesOf(h.Vertices()))
	} else {
		if g.From(u.ID()) == nil {
			return Shortest{from: u}
		}
		shortest = NewShortestPathFrom(u, []graph.Vertex{u})
	}

	var weight Weighting
	if wg, ok := g.(graph.Weighted); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	Q := priorityQueue{{vertex: u, dist: 0}}
	for Q.Len() != 0 {
		mid := heap.Pop(&Q).(distanceVertex)
		k := shortest.indexOf[mid.vertex.ID()]
		if mid.dist > shortest.dist[k] {
			continue // do not update to larger dist
		}
		mnid := mid.vertex.ID()
		for _, v := range graph.VerticesOf(g.From(mnid)) {
			vid := v.ID()
			j, ok := shortest.indexOf[vid]
			if !ok {
				j = shortest.add(v) // encounter newly vertex for shortest path
			}
			w, ok := weight(mnid, vid)
			if !ok {
				panic("dijkstra: invalid weight")
			}
			if w < 0 {
				panic("dijkstra: found negative edge weight")
			}
			joint := shortest.dist[k] + w
			if joint < shortest.dist[j] {
				heap.Push(&Q, distanceVertex{
					vertex: v,
					dist:   joint,
				})
				shortest.set(j, joint, k)
			}
		}
	}
	return shortest
}

func PEvalDijkstraFrom(u graph.Vertex, g graph.Graph) (shortest Shortest, update map[int64]float64) {
	if h, ok := g.(graph.Graph); ok {
		if u == nil {
			return NewShortestPathFrom(u, graph.VerticesOf(h.Vertices())), nil
		}
		if h.Vertex(u.ID()) == nil {
			return NewShortestPathFrom(u, graph.VerticesOf(h.Vertices())), nil
		}
		shortest = NewShortestPathFrom(u, graph.VerticesOf(h.Vertices()))
	} else {
		if g.From(u.ID()) == nil {
			return Shortest{from: u}, nil
		}
		shortest = NewShortestPathFrom(u, []graph.Vertex{u})
	}

	var weight Weighting
	if wg, ok := g.(graph.Weighted); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	update = make(map[int64]float64)

	Q := priorityQueue{{vertex: u, dist: 0}}
	for Q.Len() != 0 {
		mid := heap.Pop(&Q).(distanceVertex)
		k := shortest.indexOf[mid.vertex.ID()]
		if mid.dist > shortest.dist[k] {
			continue // do not update to larger dist
		}
		mnid := mid.vertex.ID()
		for _, v := range graph.VerticesOf(g.From(mnid)) {
			vid := v.ID()
			j, ok := shortest.indexOf[vid]
			if !ok {
				j = shortest.add(v) // encounter newly vertex for shortest path
			}
			w, ok := weight(mnid, vid)
			if !ok {
				panic("dijkstra: invalid weight")
			}
			if w < 0 {
				panic("dijkstra: found negative edge weight")
			}
			joint := shortest.dist[k] + w
			if joint < shortest.dist[j] {
				heap.Push(&Q, distanceVertex{
					vertex: v,
					dist:   joint,
				})
				shortest.set(j, joint, k)
				update[vid] = joint
			}
		}
	}
	return shortest, update
}

func IncEvalDijkstraFrom(updates map[int64]float64, shortest *Shortest, u graph.Vertex, g graph.Graph) (update map[int64]float64) {
	update = make(map[int64]float64)

	shortest.from = u

	var weight Weighting
	if wg, ok := g.(graph.Weighted); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	Q := priorityQueue{}
	for vid, dist := range updates {
		heap.Push(&Q, distanceVertex{
			vertex: g.Vertex(vid),
			dist:   dist,
		})
	}
	for Q.Len() != 0 {
		mid := heap.Pop(&Q).(distanceVertex)
		k := shortest.indexOf[mid.vertex.ID()]
		if mid.dist > shortest.dist[k] {
			continue // do not update to larger dist
		}
		shortest.dist[k] = mid.dist
		mnid := mid.vertex.ID()
		for _, v := range graph.VerticesOf(g.From(mnid)) {
			vid := v.ID()
			j, ok := shortest.indexOf[vid]
			if !ok {
				j = shortest.add(v) // encounter newly vertex for shortest path
			}
			w, ok := weight(mnid, vid)
			if !ok {
				panic("dijkstra: invalid weight")
			}
			if w < 0 {
				panic("dijkstra: found negative edge weight")
			}
			joint := shortest.dist[k] + w
			if joint < shortest.dist[j] {
				heap.Push(&Q, distanceVertex{
					vertex: v,
					dist:   joint,
				})
				update[vid] = joint
			}
		}
	}
	return update
}

type distanceVertex struct {
	vertex graph.Vertex
	dist   float64
}

type priorityQueue []distanceVertex

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].dist < pq[j].dist
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(dv interface{}) {
	*pq = append(*pq, dv.(distanceVertex))
}

func (pq *priorityQueue) Pop() interface{} {
	t := *pq
	var dv interface{}
	dv, *pq = t[len(t)-1], t[:len(t)-1]
	return dv
}
