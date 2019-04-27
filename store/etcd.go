package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
)

const (
	etcdTPL    = "%s/%d/%s/%s"
	etcdRoot   = "/rabbitid"
	retryTimes = 10
)

// A Etcd 使用etcd 存储发号元数据
type Etcd struct {
	KV  v3.KV
	log *logrus.Entry
}

var (
	// ErrEtcdFail 从etcd中Range出错
	ErrEtcdFail = errors.New("etcd save error")
	// ErrEtcdNotFound 从etcd中获取单个key出错
	ErrEtcdNotFound = errors.New("etcd key not found")
)

// NewEtcd 获取redis实例
func NewEtcd(clientURI string, logger *logrus.Entry) Etcd {
	cli, err := v3.New(v3.Config{
		Endpoints:   strings.Split(clientURI, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalln("client connect error", clientURI, err.Error())
	}
	return Etcd{KV: v3.NewKV(cli), log: logger}
}

// Range 分片分配进度, 返回v 表示可用范围[v, v+size)
func (p Etcd) Range(ctx context.Context, dataCenter uint8, db, table string, size int64) (int64, error) {
	biz := fmt.Sprintf(etcdTPL, etcdRoot, dataCenter, db, table)
	l := p.log.WithFields(logrus.Fields{"action": "range", "biz": biz, "size": size})
	last, err := p.last(ctx, l, biz)
	// 查找旧值，可能不存在
	if err != nil && err != ErrEtcdNotFound {
		l.WithError(err).Error("not found")
		return 0, ErrEtcdFail
	}
	// 存在多进程竞争的问题，这里乐观认为会成功
	for i := 0; i < retryTimes; i++ {
		last, err = p.update(ctx, l, biz, last, size)
		// Txn 查找必定存在，不存在需要抛错
		switch err {
		case nil:
			return last, nil
		case ErrEtcdNotFound:
			l.WithField("last", last).WithError(err).Error("etcd update fail")
			return 0, err
		default:
			l.WithField("last", last).WithError(err).Error("etcd update fail, try again")
			continue
		}
	}
	return 0, ErrEtcdFail
}

// last 获取上一次分配的数据
func (p Etcd) last(ctx context.Context, l *logrus.Entry, biz string) (int64, error) {
	// 获取旧的数据
	resp, err := p.KV.Get(ctx, biz)
	if err != nil {
		l.WithError(err).Error("notfound last")
		return 0, err
	}
	return getKvsByKey(resp.Kvs, biz)
}

// update 更新 etcd 存储数据
func (p Etcd) update(ctx context.Context, l *logrus.Entry, biz string, min int64, size int64) (int64, error) {
	var err error
	var resp *v3.TxnResponse

	now := strconv.FormatInt(min, 10)
	next := strconv.FormatInt(min+size, 10)
	l.WithFields(logrus.Fields{"action": "Txn", "now": now, "next": next})
	// 新增还是更新
	if min == 0 {
		resp, err = p.KV.Txn(ctx).
			If(v3.Compare(v3.CreateRevision(biz), "=", 0)).
			Then(v3.OpPut(biz, next)).
			Else(v3.OpGet(biz)).
			Commit()
	} else {
		resp, err = p.KV.Txn(ctx).
			If(v3.Compare(v3.Value(biz), "=", now)).
			Then(v3.OpPut(biz, next)).
			Else(v3.OpGet(biz)).
			Commit()
	}
	if err != nil {
		l.WithError(err).Error("Txn error")
		return 0, err
	}
	if resp.Succeeded {
		return min, nil
	}
	return getKvsByKey(resp.Responses[0].GetResponseRange().GetKvs(), biz)
}

// Ping 测试连接状态
func (p Etcd) Ping(ctx context.Context) error {
	if p.KV == nil {
		return ErrEtcdFail
	}
	_, err := p.KV.Get(ctx, etcdRoot)
	return err
}

// getKvsByKey 从kv中获取制定的key
func getKvsByKey(kvs []*mvccpb.KeyValue, s string) (int64, error) {
	for _, ev := range kvs {
		if string(ev.Key) == s {
			return strconv.ParseInt(string(ev.Value), 10, 64)
		}
	}
	return 0, ErrEtcdNotFound
}

// Init 这里可以完成初始化方法，保证服务的可用
func (p Etcd) Init(dataCenter uint8) error { return nil }

func (p Etcd) BlockDB(dataCenter uint8, db string) bool { return false }
