package store

import (
	"github.com/samuel/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
)

// withLogger replaces the ZooKeeper library's default logging service with our logger.
func withLogger(logger *logrus.Entry) func(c *zk.Conn) {
	return func(c *zk.Conn) {
		c.SetLogger(wrapLogger{logger})
	}
}

type wrapLogger struct {
	*logrus.Entry
}

func (logger wrapLogger) Printf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}
