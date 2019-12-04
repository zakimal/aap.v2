package graph

type Edges interface {
	Iterator
	Edge() Edge
}

type EdgeSlicer interface {
	EdgeSlice() []Edge
}

func EdgesOf(iter Edges) []Edge {
	if iter == nil {
		return nil
	}
	length := iter.Len()
	switch {
	case length == 0:
		return nil
	case length < 0:
		panic("graph: called EdgesOf on indeterminate iterator")
	}
	switch iter := iter.(type) {
	case EdgeSlicer:
		return iter.EdgeSlice()
	}
	es := make([]Edge, 0, length)
	for iter.Next() {
		es = append(es, iter.Edge())
	}
	return es
}

type WeightedEdges interface {
	Iterator
	WeightedEdge() WeightedEdge
}

type WeightedEdgeSlicer interface {
	WeightedEdgeSlice() []WeightedEdge
}

func WeightedEdgesOf(it WeightedEdges) []WeightedEdge {
	if it == nil {
		return nil
	}
	length := it.Len()
	switch {
	case length == 0:
		return nil
	case length < 0:
		panic("graph: called WeightedEdgesOf on indeterminate iterator")
	}
	switch it := it.(type) {
	case WeightedEdgeSlicer:
		return it.WeightedEdgeSlice()
	}
	e := make([]WeightedEdge, 0, length)
	for it.Next() {
		e = append(e, it.WeightedEdge())
	}
	return e
}
