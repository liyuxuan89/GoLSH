package vectorDatabase

import (
	"fmt"
	"geeSearch/LSH"
	"math/rand"
	"testing"
	"time"
)

const vecLen = 32
const baseNum = 32

func TestInsert(t *testing.T) {
	lsh := LSH.NewCosDistanceEncoder(vecLen, baseNum)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
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

func makeQuery(db *VectorDb, vectors [][]float64, executionTime chan<- time.Duration) {
	startTime := time.Now()
	for _, v := range vectors {
		_ = db.Search(v, 5)
	}
	elapsedTime := time.Since(startTime)
	executionTime <- elapsedTime
}

func TestSearch(t *testing.T) {
	lsh := LSH.NewCosDistanceEncoder(vecLen, baseNum)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		t.Fatal("init error", err)
		return
	}
	rand.Seed(time.Now().UnixNano())

	parallelism := 1
	queryNum := 1000
	vectors := make([][]float64, queryNum)
	for i := 0; i < queryNum; i++ {
		vec := make([]float64, vecLen)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		vectors[i] = vec
	}
	ch := make(chan time.Duration)
	defer close(ch)

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
	lsh := LSH.NewCosDistanceEncoder(vecLen, baseNum)
	db, err := NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
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
			vec := make([]float64, vecLen)
			for j := 0; j < len(vec); j++ {
				vec[j] = float64(rand.Intn(100) - 50)
			}
			_ = db.Search(vec, 10)
		}
	})
}
