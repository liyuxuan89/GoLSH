package vectorDatabase

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"geeSearch/LSH"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

const defaultRetrievalNumber = 200

var ctx = context.Background()

type VectorDb struct {
	db    *gorm.DB
	rdb   *redis.Client
	lsh   LSH.EncoderLSH
	count int64
	skip  int
}

func NewVectorDb(dataSource, redisAddr string, lsh LSH.EncoderLSH) (*VectorDb, error) {
	// 1. connect to mysql
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{})
	if err != nil {
		log.Fatal("Db error: ", err)
		return nil, err
	}
	_ = db.AutoMigrate(&Vector{})
	log.Println("connect to mysql ", dataSource, " successfully")
	// 2. connect to redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("redis error: ", err)
		return nil, err
	}
	log.Println("connect to redis ", redisAddr, " successfully")
	// 3. calculate entry number
	var count int64
	db.Model(&Vector{}).Count(&count)
	skip := lsh.Len() - int(math.Log2(float64(count)*50))
	return &VectorDb{
		db:    db,
		rdb:   rdb,
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
	v.skip = v.lsh.Len() - int(math.Log2(float64(v.count)*50))
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
	v.skip = v.lsh.Len() - int(math.Log2(float64(v.count)*50))
	return vector.ID, nil
}

func (v *VectorDb) Search(vec []float64, topk int) []Vector {
	hash, err := v.lsh.Encode(vec)
	if err != nil {
		return nil
	}
	var skip int
	if v.skip <= 0 {
		skip = 0
	} else {
		skip = v.skip
	}
	var vectors []Vector
	//1. query redis first
	var cacheHit bool
	var values []string
	var count int64
	var mask uint64 = math.MaxUint64
	for i := skip; i <= v.lsh.Len() && count < int64(topk); i++ {
		mask = math.MaxUint64 << i
		count = 0
		values, err = v.rdb.ZRangeByScore(ctx, "range",
			&redis.ZRangeBy{Min: strconv.FormatUint(mask&hash, 10),
				Max: strconv.FormatUint((^mask)|hash, 10)}).Result()
		for _, str := range values {
			add, err := strconv.ParseInt(strings.Split(str, "+")[0], 10, 64)
			if err != nil {
				continue
			}
			count += add
		}
		// log.Println(i, count, values, int64(hash&mask), int64((^mask)|hash))
	}
	if count >= int64(topk) {
		queryResults1 := make([]*redis.StringSliceCmd, len(values))
		pipe := v.rdb.Pipeline()
		for i, str := range values {
			queryHash := "hash+" + strings.Split(str, "+")[1]
			queryResults1[i] = pipe.SMembers(ctx, queryHash)
		}
		_, err := pipe.Exec(ctx)
		ids := make([]string, 0)
		if err == nil {
			for _, res := range queryResults1 {
				ids_, err := res.Result()
				if err == nil {
					ids = append(ids, ids_...)
				}
			}
		}
		queryResults2 := make([]*redis.StringStringMapCmd, len(ids))
		if len(ids) >= topk {
			for i, id := range ids {
				queryResults2[i] = pipe.HGetAll(ctx, "vector+"+id)
			}
			_, err = pipe.Exec(ctx)
			if err == nil {
				for i, res := range queryResults2 {
					row, err := res.Result()
					if err == nil {
						id, err := strconv.ParseUint(ids[i], 10, 64)
						vecBytes, err := base64.StdEncoding.DecodeString(row["base64"])
						if err != nil {
							continue
						}
						vectors = append(vectors, Vector{ID: uint(id), VecBytes: vecBytes, Url: row["url"]})
					}
				}

			}
		}
		if len(vectors) >= topk {
			log.Println("cache hit !!!")
			cacheHit = true
		}
	}
	// 2. query database
	if !cacheHit {
		log.Println("cache miss !!!")
		mask = math.MaxUint64
		for i := skip; i <= v.lsh.Len() && len(vectors) < topk; i++ {
			mask = math.MaxUint64 << i
			skip = i
			v.db.Where("hash BETWEEN ? and ?", mask&hash, (^mask)|hash).Limit(defaultRetrievalNumber).Find(&vectors)
			log.Println(skip, len(vectors))
		}
		// add to cache
		sortByHash := make(map[uint64][]Vector)
		for _, vec := range vectors {
			if _, ok := sortByHash[vec.Hash]; !ok {
				sortByHash[vec.Hash] = make([]Vector, 0)
			}
			sortByHash[vec.Hash] = append(sortByHash[vec.Hash], vec)
		}
		pipe := v.rdb.Pipeline()
		for key, val := range sortByHash {
			for _, vec := range val {
				encodeString := base64.StdEncoding.EncodeToString(vec.VecBytes)
				err = pipe.SAdd(ctx, "hash+"+strconv.FormatUint(key, 10),
					strconv.FormatUint(uint64(vec.ID), 10)).Err()
				if err != nil {
					log.Println("redis error:", err)
					continue
				}
				err = pipe.HSet(ctx, "vector+"+strconv.FormatUint(uint64(vec.ID), 10),
					"base64", encodeString, "url", vec.Url).Err()
				if err != nil {
					log.Println("redis error:", err)
					continue
				}
			}
			err = pipe.ZAdd(ctx, "range", &redis.Z{Score: float64(key),
				Member: fmt.Sprintf("%d+%d", len(val), key)}).Err()
		}
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Println("caching fail")
		}
	}
	// 3. order by distance
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
