package util

import "math"

func CalcDistance(x1 int, y1 int, x2 int, y2 int) int {
    return int(math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2)))
}
