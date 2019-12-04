package iterator

//
//import (
//	"github.com/zakimal/aap.v2/graph"
//	"reflect"
//)
//
//type RandomEdges struct {
//	edges reflect.Value
//	iter  *reflect.MapIter
//	pos   int
//	curr  graph.Edge
//}
//
//func NewRandomEdges(edges map[int64]graph.Edge) *RandomEdges {
//	re := reflect.ValueOf(edges)
//	return &RandomEdges{
//		edges: re,
//		iter:  re.MapRange(),
//		pos:   0,
//		curr:  nil,
//	}
//}
//
//func (e *RandomEdges) Len() int {
//	return e.edges.Len() - e.pos
//}
//
//func (e *RandomEdges) Next() bool {
//	if e.pos >= e.edges.Len() {
//		return false
//	}
//	ok := e.iter.Next()
//	if ok {
//		e.pos++
//		e.curr = e.iter.Value().Interface().(graph.Edge)
//	}
//	return ok
//}
//
//func (e *RandomEdges) Edge() graph.Edge {
//	return e.curr
//}
//
//func (e *RandomEdges) Reset() {
//	e.curr = nil
//	e.pos = 0
//	e.iter = e.edges.MapRange()
//}
