package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testRedis = "127.0.0.1:6379"
)

func TestRedis_Range(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	client := NewRedis(testRedis, log)

	// 清理旧数据
	biz := fmt.Sprintf(redisPrefix, testDC, testDB, testTable)
	value := client.conn.Del(biz)
	assert.NoError(t, value.Err())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	n, err := client.Range(ctx, testDC, testDB, testTable, testSize)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0))

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	cancel()
	n, err = client.Range(ctx, testDC, testDB, testTable, testSize)
	assert.NoError(t, err)
	assert.Equal(t, n, testSize)
}

func TestRedis_Ping(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	client := NewRedis(testRedis, log)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := client.Ping(ctx)
	cancel()
	assert.NoError(t, err)
}
