package helper

import (
	"math"
)

func Float64Equals(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}
