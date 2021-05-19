package main

import (
	"fmt"
	"geeSearch/LSH"
	"geeSearch/vectorDatabase"
)

func main() {
	// username:password@protocol(address)/dbname?param=value
	lsh := LSH.NewCosDistanceEncoder(4, 8)
	db, err := vectorDatabase.NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		return
	}
	fmt.Println(db, err)
}
