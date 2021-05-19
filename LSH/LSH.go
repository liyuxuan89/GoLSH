package LSH

import (
	"math"
)

type EncoderLSH interface {
	Encode(vec []float64) (uint64, error)
	Len() int
}

func GetAngle(vec1, vec2 []float64) float64 {
	dotProduct := DotProduct(vec1, vec2)
	lenVec1 := GetLength(vec1)
	lenVec2 := GetLength(vec2)
	angle := dotProduct / (lenVec1 * lenVec2)
	angle = math.Acos(angle)
	return angle
}

func DotProduct(vec1, vec2 []float64) float64 {
	var dotProduct float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
	}
	return dotProduct
}

func GetLength(vec []float64) float64 {
	var lenVec float64
	for i := 0; i < len(vec); i++ {
		lenVec += vec[i] * vec[i]
	}
	return math.Sqrt(lenVec)
}
