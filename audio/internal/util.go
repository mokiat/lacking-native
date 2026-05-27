package internal

import "math"

func float32ToInt16(value float32) int16 {
	value = max(-1.0, min(value, 1.0)) // prevent overflow
	if value >= 0.0 {
		return int16(value * float32(math.MaxInt16))
	} else {
		return int16(-value * float32(math.MinInt16))
	}
}
