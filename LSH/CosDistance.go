package LSH

import (
	"fmt"
	"math/rand"
)

type CosDistanceEncoder struct {
	base    [][]float64
	vecLen  int
	baseNum int
}

func NewCosDistanceEncoder(vecLen int, baseNum int) *CosDistanceEncoder {
	base := make([][]float64, baseNum)
	for i := 0; i < baseNum; i++ {
		base[i] = make([]float64, vecLen)
		for j := 0; j < vecLen; j++ {
			rand.Seed(int64(i*vecLen + j))
			rnd := rand.Float64() - 0.5
			base[i][j] = rnd
		}
	}
	return &CosDistanceEncoder{
		base:    base,
		baseNum: baseNum,
		vecLen:  vecLen,
	}
}

func (c *CosDistanceEncoder) Encode(vec []float64) (uint64, error) {
	if len(vec) != c.vecLen {
		return 0, fmt.Errorf("vec should have len %d, but got len %d", c.vecLen, len(vec))
	}
	var enc uint64
	var mask uint64 = 1
	for i := 0; i < c.baseNum; i++ {
		if DotProduct(vec, c.base[i]) > 0 {
			enc = enc | mask
		}
		mask = mask << 1
	}
	return enc, nil
}

func (c *CosDistanceEncoder) Len() int {
	return c.baseNum
}
