package set

type IntSet map[int]struct{}

func (s IntSet) Add(e int) {
	s[e] = struct{}{}
}

func (s IntSet) Has(e int) bool {
	_, ok := s[e]
	return ok
}

func (s IntSet) Remove(e int) {
	delete(s, e)
}

func (s IntSet) Count() int {
	return len(s)
}

func IntSetEqual(a, b IntSet) bool {
	if intSetSame(a, b) {
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
