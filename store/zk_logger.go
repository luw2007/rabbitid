package store

import (
	"fmt"

	kitlog "github.com/go-kit/kit/log"
	"github.com/samuel/go-zookeeper/zk"
)

// withLogger replaces the ZooKeeper library's default logging service with our
// own Go kit logger.
func withLogger(logger kitlog.Logger) func(c *zk.Conn) {
	return func(c *zk.Conn) {
		c.SetLogger(wrapLogger{logger})
	}
}

type wrapLogger struct {
	kitlog.Logger
}

func (logger wrapLogger) Printf(format string, args ...interface{}) {
	logger.Log("msg", fmt.Sprintf(format, args...))
}
