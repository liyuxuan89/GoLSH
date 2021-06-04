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

const defaultRetrievalNumber = 1000

type VectorDb struct {
	db    *gorm.DB
	lsh   LSH.EncoderLSH
	count int64
	skip  int
}

func NewVectorDb(dataSource string, lsh LSH.EncoderLSH) (*VectorDb, error) {
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{})
	if err != nil {
		log.Fatal("Db error: ", err)
		return nil, err
	}
	_ = db.AutoMigrate(&Vector{})
	var count int64
	db.Model(&Vector{}).Count(&count)
	skip := lsh.Len() - int(math.Log2(float64(count)/50))
	return &VectorDb{
		db:    db,
		lsh:   lsh,
		count: count,
		skip:  skip,
	}, nil
}

func (v *VectorDb) InsertBatch(vecs [][]float64, urls []string) (ids []uint, errs []error) {
	insertItems := make([]*Vector, len(vecs))
	errs = make([]error, len(vecs))

	for i := 0; i < len(vecs); i++ {
		hash, err := v.lsh.Encode(vecs[i])
		if err != nil {
			errs[i] = err
		}
		bytes := v.vecToByte(vecs[i])
		vector := &Vector{Hash: hash, VecBytes: bytes, Url: urls[i]}
		insertItems[i] = vector
	}
	tx := v.db.Create(insertItems)
	v.count += tx.RowsAffected
	v.skip = v.lsh.Len() - int(math.Log2(float64(v.count)))
	ids = make([]uint, len(vecs))
	for i, item := range insertItems {
		ids[i] = item.ID
	}
	return
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
	v.count += 1
	v.skip = v.lsh.Len() - int(math.Log2(float64(v.count)))
	return vector.ID, nil
}

func (v *VectorDb) Search(vec []float64, topk int) []Vector {
	hash, err := v.lsh.Encode(vec)
	if err != nil {
		return nil
	}
	var vectors []Vector
	var skip int
	if v.skip <= 0 {
		v.db.Where("hash = ?", hash).Find(&vectors)
		skip = 1
	} else {
		skip = v.skip
	}
	//startTime := time.Now()
	for i := skip; i <= v.lsh.Len() && len(vectors) < topk; i++ {
		//fmt.Println(i)
		var mask uint64
		mask = math.MaxUint64 << i
		v.db.Where("hash BETWEEN ? and ?", mask&hash, (^mask)|hash).Limit(defaultRetrievalNumber).Find(&vectors)
	}
	//fmt.Println("query time:", time.Since(startTime), len(vectors))
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
