package vectorDatabase

type Vector struct {
	ID uint `gorm:"primaryKey"`
	Hash uint64 `gorm:"index"`
}
