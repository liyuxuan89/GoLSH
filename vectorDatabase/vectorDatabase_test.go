package vectorDatabase

import (
	"fmt"
	"geeSearch/LSH"
	"math/rand"
	"testing"
	"time"
)

func makeQuery(db *VectorDb, vectors [][]float64, executionTime chan<- time.Duration) {
	startTime := time.Now()
	for _, v := range vectors {
		_ = db.Search(v, 10)
	}
	elapsedTime := time.Since(startTime)
	executionTime <- elapsedTime
}

func TestSearch(t *testing.T) {
	lsh := LSH.NewCosDistanceEncoder(768, 32)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		t.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())

	parallelism := 2
	queryNum := 500
	vectors := make([][]float64, queryNum)
	for i := 0; i < queryNum; i++ {
		vec := make([]float64, 768)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		vectors[i] = vec
	}
	ch := make(chan time.Duration)
	defer close(ch)
	//wg := &sync.WaitGroup{}

	for i := 0; i < parallelism; i++ {
		go makeQuery(db, vectors, ch)
	}
	recvNum := 0
	var totalTime time.Duration
	for t := range ch {
		totalTime += t
		recvNum += 1
		if recvNum == parallelism {
			break
		}
	}
	fmt.Println(totalTime.Seconds() / (float64(parallelism * queryNum)))
}

func BenchmarkSearch(b *testing.B) {
	b.StopTimer()
	lsh := LSH.NewCosDistanceEncoder(768, 32)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		b.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		vec := make([]float64, 768)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		_ = db.Search(vec, 10)
	}
}

func BenchmarkSearchParallel(b *testing.B) {
	b.StopTimer()
	lsh := LSH.NewCosDistanceEncoder(768, 32)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		b.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vec := make([]float64, 768)
			for j := 0; j < len(vec); j++ {
				vec[j] = float64(rand.Intn(100) - 50)
			}
			_ = db.Search(vec, 10)
		}
	})
}
