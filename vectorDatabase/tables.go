package vectorDatabase

type Vector struct {
	ID       uint   `gorm:"primaryKey"`
	Hash     uint64 `gorm:"index:idx_vectors_hash"`
	VecBytes []byte
	Url		 string `gorm:"type:varchar(100)"`
	Vec      []float64 `gorm:"-"`
	Dis      float64   `gorm:"-"`
}
