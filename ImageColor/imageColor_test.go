package ImageColor

import (
	"testing"
)

func TestKMeans(t *testing.T) {
	var vectors = make([][]float64, 100)
	for i:=0; i<100; i++ {
		vectors[i] = []float64{float64(i), float64(i)}
	}
	kMeans(vectors, 2, 3, 1000)
}
