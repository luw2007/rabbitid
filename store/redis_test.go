package store

import (
	"context"
	"fmt"
	"os"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

const (
	testRedis = "127.0.0.1:6379"
)

func TestRedis_Range(t *testing.T) {
	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stderr))
	client := NewRedis(testRedis, logger)

	// 清理旧数据
	biz := fmt.Sprintf(redisPrefix, testDC, testKey)
	value := client.conn.Del(biz)
	assert.NoError(t, value.Err())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	n, err := client.Range(ctx, testDC, testTable, testKey, testSize)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0))

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	cancel()
	n, err = client.Range(ctx, testDC, testTable, testKey, testSize)
	assert.NoError(t, err)
	assert.Equal(t, n, testSize)
}

func TestRedis_Ping(t *testing.T) {
	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stderr))
	client := NewRedis(testRedis, logger)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := client.Ping(ctx)
	cancel()
	assert.NoError(t, err)
}
