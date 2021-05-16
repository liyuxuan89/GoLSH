package vectorDatabase

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"log"
)

type vectorDb struct {
	engine *xorm.Engine
}

func NewVectorDb(driverName, dataSource string) (*vectorDb, error) {
	engine, err := xorm.NewEngine(driverName, dataSource)
	if err != nil {
		log.Fatal("Db error: ", err)
		return nil, err
	}
	err = engine.Sync2(new(Vector))

	vec := &Vector{
		Hash: 10,
	}
	_, err = engine.Insert(vec)
	fmt.Println(err)
	return &vectorDb{
		engine: engine,
	}, nil
}



func (v *vectorDb) Close()  {
	_ = v.engine.Close()
}
