package internal

import "math"

func int16ToFloat32(value int16) float32 {
	if value >= 0 {
		return float32(value) / float32(math.MaxInt16)
	} else {
		return -float32(value) / float32(math.MinInt16)
	}
}

func float32ToInt16(value float32) int16 {
	if value >= 0.0 {
		return int16(value * float32(math.MaxInt16))
	} else {
		return -int16(value * float32(math.MinInt16))
	}
}
