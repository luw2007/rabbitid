package store

import (
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/samuel/go-zookeeper/zk"
)

type zkConfig struct {
	logger          log.Logger
	acl             []zk.ACL
	credentials     []byte
	connectTimeout  time.Duration
	sessionTimeout  time.Duration
	rootNodePayload [][]byte
	eventHandler    func(zk.Event)
}

// Option functions enable friendly APIs.
type Option func(*zkConfig) error

// Credentials returns an Option specifying a user/password combination which
// the client will use to authenticate itself with.
func Credentials(user, pass string) Option {
	return func(c *zkConfig) error {
		if user == "" || pass == "" {
			return ErrInvalidCredentials
		}
		c.credentials = []byte(user + ":" + pass)
		return nil
	}
}

// ConnectTimeout returns an Option specifying a non-default connection timeout
// when we try to establish a connection to a ZooKeeper server.
func ConnectTimeout(t time.Duration) Option {
	return func(c *zkConfig) error {
		if t.Seconds() < 1 {
			return errors.New("invalid connect timeout (minimum value is 1 second)")
		}
		c.connectTimeout = t
		return nil
	}
}

// SessionTimeout returns an Option specifying a non-default session timeout.
func SessionTimeout(t time.Duration) Option {
	return func(c *zkConfig) error {
		if t.Seconds() < 1 {
			return errors.New("invalid session timeout (minimum value is 1 second)")
		}
		c.sessionTimeout = t
		return nil
	}
}

// Payload returns an Option specifying non-default data values for each znode
// created by CreateParentNodes.
func Payload(payload [][]byte) Option {
	return func(c *zkConfig) error {
		c.rootNodePayload = payload
		return nil
	}
}

// EventHandler returns an Option specifying a callback function to handle
// incoming zk.Event payloads (ZooKeeper connection events).
func EventHandler(handler func(zk.Event)) Option {
	return func(c *zkConfig) error {
		c.eventHandler = handler
		return nil
	}
}
