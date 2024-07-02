package abc

import (
	"context"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type NumdFloatd interface {
	Numd | Floatd
}

type StrdNumd interface {
	Stringd | Numd
}

type Numd interface {
	~int | ~int32 | ~int64
}

type Floatd interface {
	~float32 | ~float64
}

type Stringd interface {
	~string
}

var (
	db *gorm.DB

	RDB  *redis.Client
	Rctx context.Context

	RDB14  *redis.Client
	Rctx14 context.Context
	//...
)

//const (
//	Scheme    = "http://"
//	SchemeTLS = "https://"
//	RdbAddr   = "localhost:6379"
//)

func init() {
	bootSql()
}

func Tx() *gorm.DB {
	return db.Begin()
}

// wang meng..
type OAresponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type Data struct {
	SomeThings string `json:"someThings" mapstructure:"someThings"`
}
