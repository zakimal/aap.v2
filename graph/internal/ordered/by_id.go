package ordered

import "github.com/zakimal/aap.v2/graph"

type ByID []graph.Vertex

func (v ByID) Len() int           { return len(v) }
func (v ByID) Less(i, j int) bool { return v[i].ID() < v[j].ID() }
func (v ByID) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

// Reverse reverses the order of nodes.
func Reverse(nodes []graph.Vertex) {
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
}
