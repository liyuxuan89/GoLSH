package vectorDatabase

type Vector struct {
	Id int64 `xorm:"pk autoincr"`
	Hash uint32 `xorm:"mediumint"`
}
