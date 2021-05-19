package vectorDatabase

import (
	"errors"
	"geeSearch/LSH"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math"
)

type vectorDb struct {
	db *gorm.DB
	lsh LSH.EncoderLSH
}

func NewVectorDb(dataSource string, lsh LSH.EncoderLSH) (*vectorDb, error) {
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{})
	if err != nil {
		log.Fatal("Db error: ", err)
		return nil, err
	}
	_ = db.AutoMigrate(&Vector{})
	return &vectorDb{
		db: db,
		lsh: lsh,
	}, nil
}

func (v *vectorDb) Insert(vec []float64) (id uint, err error) {
	hash, err := v.lsh.Encode(vec)
	if err != nil {
		return
	}
	vector := &Vector{Hash: hash}
	tx := v.db.Create(vector)
	if tx.RowsAffected != 1 {
		err = errors.New("inserting error")
		return
	}
	return vector.ID, nil
}

func (v *vectorDb) Search(vec []float64, topk int) ([]Vector) {
	hash, err := v.lsh.Encode(vec)
	if err != nil {
		return nil
	}
	var vectors []Vector
	v.db.Where("hash = ?", hash).Find(vectors)
	for i:=0 ; i <= v.lsh.Len() && len(vectors) <= topk; i++ {
		var mask uint64
		mask = math.MaxUint64 >> (64 - i + 1)
		v.db.Where("hash BETWEEN ? and ?", mask ^ hash, mask | hash).Find(vectors)
	}
	return vectors
}
