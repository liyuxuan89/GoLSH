package vectorDatabase

import (
	"encoding/binary"
	"errors"
	"geeSearch/LSH"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math"
	"sort"
)

type VectorDb struct {
	db  *gorm.DB
	lsh LSH.EncoderLSH
}

func NewVectorDb(dataSource string, lsh LSH.EncoderLSH) (*VectorDb, error) {
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{})
	if err != nil {
		log.Fatal("Db error: ", err)
		return nil, err
	}
	_ = db.AutoMigrate(&Vector{})
	return &VectorDb{
		db:  db,
		lsh: lsh,
	}, nil
}

func (v *VectorDb) Insert(vec []float64, url string) (id uint, err error) {
	hash, err := v.lsh.Encode(vec)
	if err != nil {
		return
	}
	bytes := v.vecToByte(vec)
	vector := &Vector{Hash: hash, VecBytes: bytes, Url: url}
	tx := v.db.Create(vector)
	if tx.RowsAffected != 1 {
		err = errors.New("inserting error")
		return
	}
	return vector.ID, nil
}

func (v *VectorDb) Search(vec []float64, topk int) []Vector {
	hash, err := v.lsh.Encode(vec)
	if err != nil {
		return nil
	}
	var vectors []Vector
	v.db.Where("hash = ?", hash).Find(&vectors)
	for i := 1; i <= v.lsh.Len() && len(vectors) < topk; i++ {
		var mask uint64
		mask = math.MaxUint64 << i
		v.db.Where("hash BETWEEN ? and ?", mask&hash, (^mask)|hash).Find(&vectors)
	}
	for i := 0; i < len(vectors); i++ {
		vectors[i].Vec = v.byteToVec(vectors[i].VecBytes)
		vectors[i].Dis = v.lsh.Distance(vectors[i].Vec, vec) // cos similarity
	}
	sort.Slice(vectors, func(i, j int) bool {
		return vectors[i].Dis < vectors[j].Dis
	})
	return vectors[0:topk]
}

func (v *VectorDb) vecToByte(vec []float64) []byte {
	bytes := make([]byte, 8*len(vec))
	for i, ve := range vec {
		bits := math.Float64bits(ve)
		binary.LittleEndian.PutUint64(bytes[i*8:(i+1)*8], bits)
	}
	return bytes
}

func (v *VectorDb) byteToVec(vec []byte) []float64 {
	vecFloat := make([]float64, len(vec)/8)
	for i := 0; i < len(vec)/8; i++ {
		bits := binary.LittleEndian.Uint64(vec[i*8 : (i+1)*8])
		vecFloat[i] = math.Float64frombits(bits)
	}
	return vecFloat
}
