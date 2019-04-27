package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
)

// DefaultACL is the default ACL to use for creating znodes.
var (
	DefaultACL            = zk.WorldACL(zk.PermAll)
	ErrInvalidCredentials = errors.New("invalid credentials provided")
	ErrZKFail             = errors.New("zk save error")
)

const (
	zkTPL  = "%s/%d/%s/%s"
	zkRoot = "/rabbitid"
	// DefaultConnectTimeout is the default timeout to establish a connection to
	// a ZooKeeper node.
	DefaultConnectTimeout = 2 * time.Second
	// DefaultSessionTimeout is the default timeout to keep the current
	// ZooKeeper session alive during a temporary disconnect.
	DefaultSessionTimeout = 5 * time.Second
	// defaultFailSleep 失败休眠时间
	defaultFailSleep = 1 * time.Second
)

// A ZK 使用zookeeper做存储
type ZK struct {
	conn    *zk.Conn
	config  zkConfig
	active  bool
	quit    chan struct{}
	blackDB map[string]time.Time
	log     *logrus.Entry
}

// NewZK 获取redis实例
func NewZK(clientURI string, logger *logrus.Entry, options ...Option) ZK {
	servers := strings.Split(clientURI, ",")
	defaultEventHandler := func(event zk.Event) {
		logrus.WithFields(logrus.Fields{
			"eventtype": event.Type.String(),
			"server":    event.Server,
			"state":     event.State.String(),
		}).WithError(event.Err)
	}
	config := zkConfig{
		acl:            DefaultACL,
		connectTimeout: DefaultConnectTimeout,
		sessionTimeout: DefaultSessionTimeout,
		eventHandler:   defaultEventHandler,
		logger:         logger,
	}
	for _, option := range options {
		if err := option(&config); err != nil {
			panic(err)
		}
	}
	// dialer overrides the default ZooKeeper library Dialer so we can configure
	// the connectTimeout. The current library has a hardcoded value of 1 second
	// and there are reports of race conditions, due to slow DNS resolvers and
	// other network latency issues.
	dialer := func(network, address string, _ time.Duration) (net.Conn, error) {
		return net.DialTimeout(network, address, config.connectTimeout)
	}
	conn, _, err := zk.Connect(servers, config.sessionTimeout, withLogger(logger), zk.WithDialer(dialer))
	if err != nil {
		log.Fatal("zk connect error", clientURI)
	}
	return ZK{conn: conn, config: config, active: true, quit: make(chan struct{}), blackDB: make(map[string]time.Time), log: logger}
}

// Range 分片分配进度, 返回v 表示可用范围[v, v+size)
func (p ZK) Range(_ context.Context, dataCenter uint8, db, table string, size int64) (int64, error) {
	if last, ok := p.blackDB[db]; ok {
		if time.Since(last) < defaultFailSleep {
			return 0, ErrDBNotExists
		}
		p.blackDB[db] = time.Now()
		fmt.Println("black", last, "sice", time.Since(last))
	}
	biz := fmt.Sprintf(zkTPL, zkRoot, dataCenter, db, table)
	l := p.log.WithFields(logrus.Fields{
		"action": "Get",
		"biz":    biz,
		"size":   size,
	})

	var min int64
	var next string
	var data []byte
	var stat *zk.Stat
	var err error
	// 存在多进程竞争的问题，这里乐观认为会成功
	for i := 0; i < retryTimes; i++ {
		if i > 0 {
			l.WithFields(logrus.Fields{
				"next":   next,
				"action": "save",
				"min":    min,
				"retry":  i,
			})
		}
		data, stat, err = p.conn.Get(biz)
		switch err {
		default:
			// err !=nil && err != zk.ErrNoNode
			l.WithError(err).Error("can't catch")
			continue
		case nil:
			if min, err = strconv.ParseInt(string(data), 10, 64); err != nil {
				l.WithField("data", string(data)).WithError(err).Error("parse error")
				return 0, err
			}
			next = strconv.FormatInt(min+size, 10)
			_, err = p.conn.Set(biz, []byte(next), stat.Version)
		case zk.ErrNoNode:
			next = strconv.FormatInt(size, 10)
			_, err = p.conn.Create(biz, []byte(next), 0, p.config.acl)
		}
		switch err {
		default:
			l.WithField("action", "save").WithError(err).Error()
			return 0, ErrZKFail
		case zk.ErrNodeExists, zk.ErrBadVersion:
			continue
		case zk.ErrNoNode:
			if !p.checkDB(dataCenter, db) {
				p.blackDB[db] = time.Now()
			}
			return 0, ErrDBNotExists
		case nil:
			return min + size, nil
		}
	}
	return 0, ErrZKFail
}

// Ping 测试连接状态
func (p ZK) Ping(_ context.Context) error {
	if p.active {
		_, _, err := p.conn.Get(zkRoot)
		return err
	}
	return nil
}

// Stop implements the ZooKeeper Client interface.
func (p *ZK) Stop() {
	p.active = false
	close(p.quit)
	p.conn.Close()
}

// Init 这里可以完成初始化方法，保证服务的可用
func (p ZK) Init(dataCenter uint8) error {
	biz := fmt.Sprintf("%s/%d", zkRoot, dataCenter)
	_, _, err := p.conn.Get(biz)
	if err != nil {
		panic(err)
	}
	return nil
}

func (p ZK) checkDB(dataCenter uint8, db string) bool {
	biz := fmt.Sprintf("%s/%d/%s", zkRoot, dataCenter, db)
	_, _, err := p.conn.Get(biz)
	if err != nil {
		fmt.Println("checkDB fail")
		return false
	}
	return true
}

func (p ZK) BlockDB(dataCenter uint8, db string) bool {
	if last, ok := p.blackDB[db]; ok {
		if time.Since(last) < defaultFailSleep {
			return false
		}
	}
	return true
}
