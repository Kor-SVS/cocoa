package util

func MapRange(n, srcMin, srcMax, dstMin, dstMax float64) float64 {
	return (n-srcMin)/(srcMax-srcMin)*(dstMax-dstMin) + dstMin
}
