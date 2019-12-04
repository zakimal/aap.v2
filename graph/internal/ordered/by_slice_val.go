package ordered

type BySliceValues [][]int64

func (c BySliceValues) Len() int { return len(c) }
func (c BySliceValues) Less(i, j int) bool {
	a, b := c[i], c[j]
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for k, v := range a[:l] {
		if v < b[k] {
			return true
		}
		if v > b[k] {
			return false
		}
	}
	return len(a) < len(b)
}
func (c BySliceValues) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
