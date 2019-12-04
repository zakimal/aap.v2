package set

type Int64Set map[int64]struct{}

func (s Int64Set) Add(e int64) {
	s[e] = struct{}{}
}

func (s Int64Set) Has(e int64) bool {
	_, ok := s[e]
	return ok
}

func (s Int64Set) Remove(e int64) {
	delete(s, e)
}

func (s Int64Set) Count() int {
	return len(s)
}

func Int64SetEqual(a, b Int64Set) bool {
	if int64SetSame(a, b) {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for e := range a {
		if _, ok := b[e]; !ok {
			return false
		}
	}
	return true
}
