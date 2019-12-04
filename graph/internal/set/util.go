package set

import (
	"unsafe"
)

func vertexSetSame(a, b VertexSet) bool {
	return *(*uintptr)(unsafe.Pointer(&a)) == *(*uintptr)(unsafe.Pointer(&b))
}

func intSetSame(a, b IntSet) bool {
	return *(*uintptr)(unsafe.Pointer(&a)) == *(*uintptr)(unsafe.Pointer(&b))
}

func int64SetSame(a, b Int64Set) bool {
	return *(*uintptr)(unsafe.Pointer(&a)) == *(*uintptr)(unsafe.Pointer(&b))
}
