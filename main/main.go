package main

import (
	"fmt"
	"geeSearch/ImageColor"
	"geeSearch/LSH"
	"geeSearch/vectorDatabase"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func insertData(db *vectorDatabase.VectorDb)  {
	// read from file
	f, err := os.Open("./main/landscape1000.txt")
	if err != nil {
		log.Fatal("open file error " + err.Error())
	}
	fd, err := io.ReadAll(f)
	if err != nil {
		log.Fatal("read file error" + err.Error())
	}
	fdString := string(fd)
	for _, l := range strings.Split(fdString, "\n") {
		url := strings.Trim(l, "\r\n ")
		vec := getData(url)
		if vec == nil {
			continue
		}
		_, _ = db.Insert(vec, url)
	}
}

func getData(url string) []float64 {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("get error " + err.Error() + " " + url)
		return nil
	}
	defer resp.Body.Close()
	im, err := ImageColor.NewIm(resp.Body)
	if err != nil {
		return nil
	}
	hist := im.ExtractColor()
	return hist
}

func main() {
	// username:password@protocol(address)/dbname?param=value
	lsh := LSH.NewCosDistanceEncoder(768, 32)
	db, err := vectorDatabase.NewVectorDb("root:123456@tcp(localhost:3306)/test_db?charset=utf8", lsh)
	if err != nil {
		return
	}
	fs, _ := os.Open("./main/test.jpg")
	im, err := ImageColor.NewIm(fs)
	if err != nil {
		return
	}
	hist := im.ExtractColor()
	ret := db.Search(hist, 10)
	for _, r := range ret {
		fmt.Println(r.Url)
	}

	//insertData(db)
	//fmt.Println(db, err)
	//for i:=0; i<100; i++ {
	//	for j:=0; j<100; j++ {
	//		_, _ = db.Insert([]float64{float64(i)-50, float64(j)-50})
	//	}
	//}
	//ret := db.Search([]float64{1, 3}, 10)
	//fmt.Println(ret)
}
