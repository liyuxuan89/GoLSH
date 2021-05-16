package LSH

import (
	"fmt"
	"math"
	"testing"
)

func TestVec(t *testing.T) {
	vec1 := []float64{0, 2}
	vec2 := []float64{3, 0}
	angle := GetAngle(vec1, vec2)
	if angle != math.Pi / 2 {
		t.Fatal(fmt.Sprintf("expect %f, get %f", math.Pi / 2, angle))
	}
}

func TestCos(t *testing.T) {
	cosEncoder := NewCosDistanceEncoder(2, 2)
	enc, _ := cosEncoder.Encode([]float64{1, 4})
	fmt.Println(cosEncoder.base)
	fmt.Println(enc)
}

func BenchmarkVec(b *testing.B) {
	vec1 := make([]float64, 2048)
	vec2 := make([]float64, 2048)
	for i := 0; i < 2048; i++ {
		vec1 = append(vec1, 1)
		vec2 = append(vec2, 1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetAngle(vec1, vec2)
	}
}

