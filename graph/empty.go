package graph

const Empty = nothing

var (
	_ Iterator           = Empty
	_ Vertices           = Empty
	_ VertexSlicer       = Empty
	_ Edges              = Empty
	_ EdgeSlicer         = Empty
	_ WeightedEdges      = Empty
	_ WeightedEdgeSlicer = Empty
)

const nothing = empty(true)

type empty bool

func (empty) Next() bool                        { return false }
func (empty) Len() int                          { return 0 }
func (empty) Reset()                            {}
func (empty) Vertex() Vertex                    { return nil }
func (empty) VertexSlice() []Vertex             { return nil }
func (empty) Edge() Edge                        { return nil }
func (empty) EdgeSlice() []Edge                 { return nil }
func (empty) WeightedEdge() WeightedEdge        { return nil }
func (empty) WeightedEdgeSlice() []WeightedEdge { return nil }
