package linear

import "github.com/zakimal/aap.v2/graph"

type VertexQueue struct {
	head int
	data []graph.Vertex
}

func (q *VertexQueue) Len() int {
	return len(q.data) - q.head
}

func (q *VertexQueue) Enqueue(v graph.Vertex) {
	if len(q.data) == cap(q.data) && q.head > 0 {
		l := q.Len()
		copy(q.data, q.data[q.head:])
		q.head = 0
		q.data = append(q.data[:l], v)
	} else {
		q.data = append(q.data, v)
	}
}

func (q *VertexQueue) Dequeue() graph.Vertex {
	if q.Len() == 0 {
		panic("queue: empty queue")
	}

	var n graph.Vertex
	n, q.data[q.head] = q.data[q.head], nil
	q.head++

	if q.Len() == 0 {
		q.head = 0
		q.data = q.data[:0]
	}

	return n
}

func (q *VertexQueue) Reset() {
	q.head = 0
	q.data = q.data[:0]
}
