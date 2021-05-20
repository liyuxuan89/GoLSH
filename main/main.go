package main

import (
	"fmt"
	"geeSearch/LSH"
	"geeSearch/vectorDatabase"
)

func main() {
	// username:password@protocol(address)/dbname?param=value
	lsh := LSH.NewCosDistanceEncoder(2, 32)
	db, err := vectorDatabase.NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		return
	}
	fmt.Println(db, err)
	//for i:=0; i<100; i++ {
	//	for j:=0; j<100; j++ {
	//		_, _ = db.Insert([]float64{float64(i)-50, float64(j)-50})
	//	}
	//}
	ret := db.Search([]float64{1, 3}, 10)
	fmt.Println(ret)
}
