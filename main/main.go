package main

import (
	"fmt"
	"geeSearch/vectorDatabase"
)

func main() {

	db, err := vectorDatabase.NewVectorDb("mysql", "root:123456@tcp(localhost:3306)/test_db?charset=utf8")
	if err != nil {
		return
	}
	defer db.Close()
	fmt.Println(db, err)
}
