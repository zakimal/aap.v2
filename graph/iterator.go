package graph

type Iterator interface {
	Next() bool
	Len() int
	Reset()
}
