package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testDB             = "test"
	testTable          = "test_1"
	testKey            = "test|test_1"
	testBenchKey       = "bench_1"
	testURI            = "127.0.0.1:2379"
	timeout            = time.Second * 50
	testSize     int64 = 100
	testDC       uint8 = 0
)

func TestEtcd_Range(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	client := NewEtcd(testURI, log)

	// 清理旧数据
	biz := fmt.Sprintf(etcdTPL, etcdRoot, testDC, testDB, testBenchKey)
	_, err := client.KV.Delete(context.TODO(), biz)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	n, err := client.Range(ctx, testDC, testDB, testTable, testSize)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0))

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	n, err = client.Range(ctx, testDC, testDB, testTable, testSize)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, n, testSize)
}

func TestEtcd_Ping(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	client := NewEtcd(testURI, logger)
	err := client.Ping(context.TODO())
	assert.NoError(t, err)
}

func BenchmarkEtcd_Range(b *testing.B) {
	logger := logrus.NewEntry(logrus.New())
	client := NewEtcd(testURI, logger)

	// 清理旧数据
	biz := fmt.Sprintf(etcdTPL, etcdRoot, testDC, testTable, testBenchKey)
	del, err := client.KV.Delete(context.TODO(), biz)
	assert.NoError(b, err)
	assert.NotEqual(b, del.Deleted, int64(0))

	for i := 0; i < b.N; i++ {
		//use b.N for looping
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		client.Range(ctx, testDC, testDB, testBenchKey, testSize)
		cancel()
	}
}
