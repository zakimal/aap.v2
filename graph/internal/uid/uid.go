package uid

import "github.com/zakimal/aap.v2/graph/internal/set"

const MAX = int64(^uint64(0) >> 1)

type IDSet struct {
	maxID      int64
	used, free set.Int64Set
}

func NewIDSet() IDSet {
	return IDSet{
		maxID: -1,
		used:  make(set.Int64Set),
		free:  make(set.Int64Set),
	}
}

func (s *IDSet) NewID() int64 {
	for id := range s.free {
		return id
	}
	if s.maxID != MAX {
		return s.maxID + 1
	}
	for id := int64(0); id <= s.maxID+1; id++ {
		if !s.used.Has(id) {
			return id
		}
	}
	panic("unreachable")
}

func (s *IDSet) Use(id int64) {
	s.used.Add(id)
	s.free.Remove(id)
	if id > s.maxID {
		s.maxID = id
	}
}

func (s *IDSet) Release(id int64) {
	s.free.Add(id)
	s.used.Remove(id)
}
