package service

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testDB   = "comment"
	testSize = 5
)

type MockStore struct {
	id *int64
	mu *sync.Mutex
}

// Range 获取数据，传入数据中心ID，业务唯一ID和获取连续的区间大小，得到连续的最大ID和错误[id-size, id)
func (p MockStore) Range(_ context.Context, dataCenter uint8, db, table string, size int64) (id int64, err error) {
	p.mu.Lock()
	id = *p.id
	*p.id += size
	p.mu.Unlock()
	return id, err
}

// Ping 检查连接
func (p MockStore) Ping(ctx context.Context) error {
	return nil
}

// Init 这里可以完成初始化方法，保证服务的可用
func (p MockStore) Init(dataCenter uint8) error { return nil }

func (p MockStore) BlockDB(dataCenter uint8, db string) bool { return false }

func TestService_Next(t *testing.T) {
	db := MockStore{
		id: new(int64),
		mu: new(sync.Mutex),
	}
	logger := logrus.NewEntry(logrus.New())
	svc := New(logger, db, testSize, 0, 60, 600)
	id, errMsg := svc.Next(context.TODO(), testDB, "next")
	assert.Equal(t, errMsg, "")
	fmt.Println(svc.Next(context.TODO(), testDB, "1"))
	assert.Equal(t, id, int64(1))

}

func TestService_Last(t *testing.T) {
	db := MockStore{
		id: new(int64),
		mu: new(sync.Mutex),
	}
	logger := logrus.NewEntry(logrus.New())
	svc := New(logger, db, testSize, 0, 60, 600)

	// 没加载报错
	table := "last"
	last, err := svc.Last(context.TODO(), testDB, table)
	assert.Equal(t, err, ErrEmpty.Error())

	id, err := svc.Next(context.TODO(), testDB, table)
	assert.Equal(t, err, "")
	last, err = svc.Last(context.TODO(), testDB, table)
	assert.Equal(t, err, "")
	assert.Equal(t, id, last)
}

func TestService_Remainder(t *testing.T) {
	db := MockStore{
		id: new(int64),
		mu: new(sync.Mutex),
	}
	logger := logrus.NewEntry(logrus.New())
	svc := New(logger, db, testSize, 0, 60, 600)

	// 没加载报错
	testTable := "remainder"
	remainder, err := svc.Remainder(context.TODO(), testDB, testTable)
	assert.Equal(t, err, ErrEmpty.Error())

	svc.Next(context.TODO(), testDB, testTable)
	remainder, err = svc.Remainder(context.TODO(), testDB, testTable)
	assert.Equal(t, err, "")
	assert.Equal(t, remainder, int64(testSize-1))
}

func BenchmarkService_Next(b *testing.B) {
	db := MockStore{
		id: new(int64),
		mu: new(sync.Mutex),
	}
	logger := logrus.NewEntry(logrus.New())
	svc := New(logger, db, 1000, 0, 60, 600)
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		//use b.N for looping
		svc.Next(ctx, testDB, "bench")
	}
}
