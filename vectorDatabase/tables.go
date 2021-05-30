package vectorDatabase

type Vector struct {
	ID       uint   `gorm:"primaryKey"`
	Hash     uint64 `gorm:"index"`
	VecBytes []byte
	Url		 string
	Vec      []float64 `gorm:"-"`
	Dis      float64   `gorm:"-"`
}
