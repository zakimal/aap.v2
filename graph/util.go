package graph

func Copy(dst Builder, src Graph) {
	vertices := src.Vertices()
	for vertices.Next() {
		dst.AddVertex(vertices.Vertex())
	}
	vertices.Reset()
	for vertices.Next() {
		u := vertices.Vertex()
		uid := u.ID()
		to := src.From(uid)
		for to.Next() {
			v := to.Vertex()
			dst.AddEdge(src.Edge(uid, v.ID()))
		}
	}
}

func CopyWeighted(dst WeightedBuilder, src Weighted) {
	vertices := src.Vertices()
	for vertices.Next() {
		dst.AddVertex(vertices.Vertex())
	}
	vertices.Reset()
	for vertices.Next() {
		u := vertices.Vertex()
		uid := u.ID()
		to := src.From(uid)
		for to.Next() {
			v := to.Vertex()
			dst.AddWeightedEdge(src.WeightedEdge(uid, v.ID()))
		}
	}
}
