package float

import "math"

// Round returns a float rounded according to "Round to nearest, ties away from zero" IEEE floaing point rounding rule
func Round(x float64) float64 {
	var absx, y float64
	absx = math.Abs(x)
	y = math.Floor(absx)
	if absx-y >= 0.5 {
		y += 1.0
	}
	return math.Copysign(y, x)
}
