package vectorDatabase

import (
	"geeSearch/LSH"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkSearch(b *testing.B) {
	b.StopTimer()
	lsh := LSH.NewCosDistanceEncoder(2, 32)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		b.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		vec := make([]float64, 2)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		_ = db.Search(vec, 10)
	}
}

func BenchmarkSearchParallel(b *testing.B) {
	b.StopTimer()
	lsh := LSH.NewCosDistanceEncoder(2, 32)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		b.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vec := make([]float64, 2)
			for j := 0; j < len(vec); j++ {
				vec[j] = float64(rand.Intn(100) - 50)
			}
			_ = db.Search(vec, 10)
		}
	})
}
