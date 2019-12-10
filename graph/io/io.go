package io

import (
	"encoding/csv"
	"fmt"
	"github.com/zakimal/aap.v2/graph"
	"github.com/zakimal/aap.v2/graph/internal/set"
	"github.com/zakimal/aap.v2/graph/simple"
	"io"
	"math"
	"os"
	"strconv"
)

func BuildEdgeCutWeightedDirectedGraph(id uint64) (graph.Graph, map[graph.Vertex]uint64, map[graph.Vertex]uint64, map[graph.Vertex]uint64, map[graph.Vertex]uint64) {
	g := simple.NewWeightedDirectedGraph(0.0, math.Inf(1))
	var (
		fit = make(map[graph.Vertex]uint64)
		fif = make(map[graph.Vertex]uint64)
		fot = make(map[graph.Vertex]uint64)
		fof = make(map[graph.Vertex]uint64)
	)
	possessionMap := make(map[int64]uint64)
	possession, err := os.Open("/Users/zak/go/src/github.com/zakimal/aap.v2/data/weighted_directed/edge_cut/possession.csv")
	if err != nil {
		panic(err)
	}
	defer possession.Close()

	possessionReader := csv.NewReader(possession)
	possessionReader.Read() // skip header
	for {
		record, err := possessionReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		vid, _ := strconv.ParseInt(record[0], 10, 64)
		wid, _ := strconv.ParseInt(record[1], 10, 64)

		possessionMap[vid] = uint64(wid)
	}

	// read edges
	edges, err := os.Open(fmt.Sprintf("/Users/zak/go/src/github.com/zakimal/aap.v2/data/weighted_directed/edge_cut/edges/%d.csv", id))
	if err != nil {
		panic(err)
	}
	defer edges.Close()

	edgeReader := csv.NewReader(edges)
	edgeReader.Read() // skip header
	for {
		record, err := edgeReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		src, _ := strconv.ParseInt(record[0], 10, 64)
		dst, _ := strconv.ParseInt(record[1], 10, 64)
		weight, _ := strconv.ParseFloat(record[2], 64)
		we := simple.WeightedEdge{
			F: simple.Vertex(src),
			T: simple.Vertex(dst),
			W: weight,
		}
		g.AddWeightedEdge(we)
	}

	vertices, err := os.Open(fmt.Sprintf("/Users/zak/go/src/github.com/zakimal/aap.v2/data/weighted_directed/edge_cut/vertices/%d.csv", id))
	if err != nil {
		panic(err)
	}
	defer vertices.Close()

	verticesReader := csv.NewReader(vertices)
	verticesReader.Read() // skip header
	inners := set.NewVertexSet() // in-partition
	for {
		svid, err := verticesReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		vid, err := strconv.ParseInt(svid[0], 10, 64)
		if err != nil {
			panic(err)
		}
		inners.Add(g.Vertex(vid))
	}

	all := g.Vertices()
	alls := set.NewVertexSet()
	for all.Next() {
		alls.Add(all.Vertex())
	}

	outers := set.Clone(alls) // out-partition
	for _, v := range alls {
		if inners.Has(v) {
			outers.Remove(v)
		}
	}

	for _, o := range outers {
		vs := g.From(o.ID())
		for vs.Next() {
			if inners.Has(vs.Vertex()) {
				fif[o] = possessionMap[o.ID()]
			}
		}
	}

	for _, i := range inners {
		vs := g.To(i.ID())
		for vs.Next() {
			if outers.Has(vs.Vertex()) {
				fit[i] = possessionMap[i.ID()]
			}
		}
	}

	for _, o := range outers {
		vs := g.To(o.ID())
		for vs.Next() {
			if inners.Has(vs.Vertex()) {
				fot[o] = possessionMap[o.ID()]
			}
		}
	}

	for _, i := range inners {
		vs := g.From(i.ID())
		for vs.Next() {
			if outers.Has(vs.Vertex()) {
				fof[i] = possessionMap[i.ID()]
			}
		}
	}

	return g, fit, fif, fot, fof
}
