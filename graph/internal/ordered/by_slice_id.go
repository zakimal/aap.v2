package ordered

import "github.com/zakimal/aap.v2/graph"

type BySliceIDs [][]graph.Vertex

func (c BySliceIDs) Len() int { return len(c) }
func (c BySliceIDs) Less(i, j int) bool {
	a, b := c[i], c[j]
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for k, v := range a[:l] {
		if v.ID() < b[k].ID() {
			return true
		}
		if v.ID() > b[k].ID() {
			return false
		}
	}
	return len(a) < len(b)
}
func (c BySliceIDs) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
