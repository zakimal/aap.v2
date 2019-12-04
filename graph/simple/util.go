package simple

import "math"

func isSame(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}
