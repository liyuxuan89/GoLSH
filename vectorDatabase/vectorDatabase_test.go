package vectorDatabase

import (
	"geeSearch/LSH"
	"math/rand"
	"testing"
	"time"
)

const vecLen = 32
const baseNum = 32

func TestInsert(t *testing.T) {
	lsh := LSH.NewCosDistanceEncoder(vecLen, baseNum)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", "localhost:6379", lsh)
	if err != nil {
		t.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	vecs := make([][]float64, 0, 1000)
	urls := make([]string, 0, 1000)
	for i := 0; i < 1000000; i++ {
		vec := make([]float64, vecLen)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		vecs = append(vecs, vec)
		urls = append(urls, "")
		if (i+1)%1000 == 0 {
			_, _ = db.InsertBatch(vecs, urls)
			vecs = make([][]float64, 0, 1000)
			urls = make([]string, 0, 1000)
		}
	}
}

func BenchmarkSearch(b *testing.B) {
	b.StopTimer()
	lsh := LSH.NewCosDistanceEncoder(vecLen, baseNum)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", "localhost:6379", lsh)
	if err != nil {
		b.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		vec := make([]float64, vecLen)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		_ = db.Search(vec, 10)
	}
}

func BenchmarkSearchParallel(b *testing.B) {
	b.StopTimer()
	lsh := LSH.NewCosDistanceEncoder(vecLen, baseNum)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", "localhost:6379", lsh)
	if err != nil {
		b.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vec := make([]float64, vecLen)
			for j := 0; j < len(vec); j++ {
				vec[j] = float64(rand.Intn(100) - 50)
			}
			_ = db.Search(vec, 10)
		}
	})
}
