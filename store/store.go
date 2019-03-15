// Package store 存储发号进度
package store

import (
	"context"
	"log"

	"errors"

	kitlog "github.com/go-kit/kit/log"
)

// A Store 存储接口
type Store interface {
	// Range 获取数据，传入数据中心ID，业务唯一ID和获取连续的区间大小，得到连续的最大ID和错误[id-size, id)
	Range(ctx context.Context, dataCenter uint8, db, table string, size int64) (id int64, err error)
	// Init 这里可以完成初始化方法，保证服务的可用
	Init(dataCenter uint8) error

	BlockDB(dataCenter uint8, db string) bool
	// Ping 检查连接
	Ping(ctx context.Context) error
}

var (
	ErrDBNotExists = errors.New("zk: db does not exist")
)

func NewStore(storeType, uri string, dataCenter uint8, logger kitlog.Logger) Store {
	var db Store
	switch storeType {
	default:
		log.Fatalln("store type Error", storeType)
	case "redis":
		db = NewRedis(uri, logger)
	case "etcd":
		db = NewEtcd(uri, logger)
	case "zk":
		db = NewZK(uri, logger)
	}
	db.Init(dataCenter)
	return db
}
