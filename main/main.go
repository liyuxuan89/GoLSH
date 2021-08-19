package main

import (
	"bytes"
	"geeSearch/ImageColor"
	"geeSearch/LSH"
	"geeSearch/vectorDatabase"
	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

type SearchQuery struct {
	DataUrl string `json:"dataUrl"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

func insertData(db *vectorDatabase.VectorDb) {
	// read from file
	f, err := os.Open("./main/urls.txt")
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

	//1. deploy
	//lsh := LSH.NewCosDistanceEncoder(768, 32)
	//db, err := vectorDatabase.NewVectorDb(
	//	"root:123456@tcp(localhost:3306)/serve_db?charset=utf8",
	//	"localhost:6379", lsh)

	// 2. test
	lsh := LSH.NewCosDistanceEncoder(32, 32)
	db, err := vectorDatabase.NewVectorDb(
		"root:123456@tcp(localhost:3306)/test_db?charset=utf8",
		"localhost:6379", lsh)

	if err != nil {
		log.Fatal("error init database !")
	}
	router := gin.Default()
	router.POST("/search", func(c *gin.Context) {
		json := SearchQuery{}
		_ = c.BindJSON(&json)
		dataURL, err := dataurl.DecodeString(json.DataUrl)
		if err != nil {
			c.String(http.StatusInternalServerError, "parse data url error")
			return
		}
		im, err := ImageColor.NewIm(bytes.NewReader(dataURL.Data))
		if err != nil {
			c.String(http.StatusInternalServerError, "parse image error")
			return
		}
		hist := im.ExtractColor()
		ret := db.Search(hist, json.Limit+json.Offset)
		urls := make([]string, json.Limit)
		for i := 0; i < json.Limit; i++ {
			urls[i] = ret[i+json.Offset].Url
		}
		c.JSON(http.StatusOK, gin.H{
			"imageUrls": urls,
		})
	})

	router.POST("/test", func(c *gin.Context) {
		vec := make([]float64, 32)
		rand.Seed(1024)
		for j := 0; j < len(vec); j++ {
			vec[j] = float64(rand.Intn(100) - 50)
		}
		ret := db.Search(vec, 10)
		c.JSON(http.StatusOK, gin.H{
			"retrievals": len(ret),
		})
	})
	_ = router.Run(":8080")
}
