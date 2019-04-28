package handle

import (
	"fmt"
	"runtime"
	"sync"

	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/redcon"

	"github.com/luw2007/rabbitid/cmd/idHttp/conf"
	"github.com/luw2007/rabbitid/cmd/idHttp/service"
	"github.com/luw2007/rabbitid/store"
)

type Handler struct {
	svc    service.Service
	db     store.Store
	logger *logrus.Entry
}

const (
	slowTime      = 20 * time.Millisecond
	cancelTimeout = 500 * time.Millisecond
	defautDB      = "_NO_APP_"
)

var (
	mu    sync.Mutex
	conns = make(map[redcon.Conn]bool)
)

func (p *Handler) Serve(conn redcon.Conn, cmd redcon.Command) {
	name := strings.ToLower(string(cmd.Args[0]))
	defer func(begin time.Time, name string) {
		if time.Since(begin) > slowTime {
			p.logger.WithFields(
				logrus.Fields{
					"slow":     time.Since(begin),
					"duration": time.Since(begin),
				})
		}
		if r := recover(); r != nil {
			var err error
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("%v", r)
			}
			stack := make([]byte, 4<<10) // 4 KB
			length := runtime.Stack(stack, true)
			p.logger.Errorf("[PANIC RECOVER] %s: %s %s\n", name, err, stack[:length])
			conn.WriteError("[PANIC RECOVER]" + name)
		}
	}(time.Now(), name)
	switch name {
	default:
		var (
			db    string
			table string
			id    int64
			msg   string
		)
		switch len(cmd.Args) {
		default:
			conn.WriteError("ERR wrong number of arguments for '" + name + "' command.")
			return
		case 2:
			names := strings.SplitN(string(cmd.Args[1]), ":", 2)
			if len(names) > 1 {
				db, table = names[0], names[1]
			} else {
				db, table = defautDB, names[0]
			}
		case 3:
			db = string(cmd.Args[1])
			table = string(cmd.Args[2])
		}
		ctx, cancel := context.WithTimeout(context.Background(), cancelTimeout)
		defer cancel()
		switch name {
		default:
			conn.WriteError("ERR unknown command '" + name + "'")
			return
		case "get", "incr", "next":
			id, msg = p.svc.Next(ctx, db, table)
		case "max":
			id, msg = p.svc.Max(ctx, db, table)
		case "last":
			id, msg = p.svc.Last(ctx, db, table)
		case "remainder":
			id, msg = p.svc.Remainder(ctx, db, table)
		}
		if msg != "" {
			conn.WriteError(msg)
			return
		}
		conn.WriteInt64(id)
	case "check":
		ctx, cancel := context.WithTimeout(context.Background(), cancelTimeout)
		defer cancel()
		err := p.db.Ping(ctx)
		if err != nil {
			conn.WriteError(err.Error())
		}
		conn.WriteString("OK")

	case "ping":
		conn.WriteString("PONG")
	case "help":
		conn.WriteString(`usage: [cmd] DB TABLE
		next
		last
		max
		remainder
	`)
	case "quit":
		conn.WriteString("OK")
		conn.Close()
	case "shutdown":
		p.ShutDown()
	}
}
func (p *Handler) usage(conn redcon.Conn, name string) {
	conn.WriteError("ERR wrong number of arguments for '" + name + "' command.")

}
func (p *Handler) Connected(conn redcon.Conn) bool {
	mu.Lock()
	conns[conn] = true
	mu.Unlock()
	return true
}

func (p *Handler) Closed(conn redcon.Conn, _ error) {
	mu.Lock()
	delete(conns, conn)
	mu.Unlock()
}

func (p *Handler) ShutDown() {
	mu.Lock()
	for conn := range conns {
		conn.Close()
	}
	mu.Unlock()
}

func NewRedisHandler(config conf.Config) *Handler {
	logger := config.Logger.WithField("app", "redis")
	db := store.NewStore(config.Store.Type, config.Store.URI, config.Generate.DataCenter, logger)
	svc := service.New(logger, db, config.Generate.Step, config.Generate.DataCenter, config.Store.Min, config.Store.Max)
	return &Handler{svc: svc, db: db, logger: logger}
}
