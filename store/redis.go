package store

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

const redisPrefix = "rabbitid_%d_%s_%s"

// A Redis 使用redis作存储
type Redis struct {
	conn *redis.Client
	log  *logrus.Entry
}

// NewRedis 获取redis实例
func NewRedis(redisAddr string, logger *logrus.Entry) Redis {
	cli := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err := cli.Ping().Result()
	if err != nil {
		log.Fatal("redis connect error", redisAddr)
	}
	return Redis{conn: cli, log: logger.WithField("store", "redis")}
}

// Range 分片分配进度, 返回v 表示可用范围[v, v+size)
func (p Redis) Range(_ context.Context, dataCenter uint8, db, table string, size int64) (int64, error) {
	biz := fmt.Sprintf(redisPrefix, dataCenter, db, table)
	value := p.conn.HIncrBy(biz, table, size)
	max := value.Val()
	p.log.WithFields(logrus.Fields{"action": "range", "biz": biz, "size": size, "last": max - size})
	return max - size, value.Err()
}

// Ping 测试连接状态
func (p Redis) Ping(_ context.Context) error {
	value := p.conn.Ping()
	return value.Err()
}

// Init 这里可以完成初始化方法，保证服务的可用
func (p Redis) Init(dataCenter uint8) error { return nil }

func (p Redis) BlockDB(dataCenter uint8, db string) bool { return false }
