package iterator

//
//import (
//	"github.com/zakimal/aap.v2/graph"
//	"reflect"
//)
//
//type RandomVertices struct {
//	vertices reflect.Value
//	iter     *reflect.MapIter
//	pos      int
//	curr     graph.Vertex
//}
//
//func NewRandomVertices(vertices map[int64]graph.Vertex) *RandomVertices {
//	rv := reflect.ValueOf(vertices)
//	return &RandomVertices{
//		vertices: rv,
//		iter:     rv.MapRange(),
//		pos:      0,
//		curr:     nil,
//	}
//}
//
//func (v *RandomVertices) Len() int {
//	return v.vertices.Len() - v.pos
//}
//
//func (v *RandomVertices) Next() bool {
//	if v.pos >= v.vertices.Len() {
//		return false
//	}
//	ok := v.iter.Next()
//	if ok {
//		v.pos++
//		v.curr = v.iter.Value().Interface().(graph.Vertex)
//	}
//	return ok
//}
//
//func (v *RandomVertices) Vertex() graph.Vertex {
//	return v.curr
//}
//
//func (v *RandomVertices) Reset() {
//	v.curr = nil
//	v.pos = 0
//	v.iter = v.vertices.MapRange()
//}
