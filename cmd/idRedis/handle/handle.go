package handle

import (
	"sync"

	"strings"

	"context"
	"os"
	"time"

	klog "github.com/go-kit/kit/log"

	"github.com/tidwall/redcon"

	"github.com/luw2007/rabbitid/cmd/idHttp/conf"
	"github.com/luw2007/rabbitid/cmd/idHttp/service"
	"github.com/luw2007/rabbitid/store"
)

type Handler struct {
	svc    service.Service
	db     store.Store
	logger klog.Logger
}

const (
	slowTime = 10 * time.Millisecond
	defautDB = "_NO_APP_"
)

var (
	mu    sync.Mutex
	conns = make(map[redcon.Conn]bool)
)

func (p *Handler) Serve(conn redcon.Conn, cmd redcon.Command) {
	defer func(begin time.Time) {
		if time.Since(begin) > slowTime {
			p.logger.Log("slow", time.Since(begin), "duration", time.Since(begin))
		}
	}(time.Now())
	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
	case "hincrby":
		p.HINCRBY(conn, cmd)
	case "incr":
		p.INCR(conn, cmd)
	case "max":
		p.Max(conn, cmd)
	case "get":
		p.Last(conn, cmd)
	case "len":
		p.Remainder(conn, cmd)
	case "ping":
		p.Ping(conn, cmd)
	case "quit":
		conn.WriteString("OK")
		conn.Close()
	case "shutdown":
		p.ShutDown()
	}
}

func (p *Handler) HINCRBY(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	db := string(cmd.Args[1])
	table := string(cmd.Args[2])
	id, msg := p.svc.Next(context.TODO(), db, table)
	if msg != "" {
		conn.WriteError(msg)
		return
	}
	conn.WriteInt64(id)
}

func (p *Handler) INCR(conn redcon.Conn, cmd redcon.Command) {
	var db, table string
	switch len(cmd.Args) {
	default:
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	case 2:
		names := strings.SplitN(string(cmd.Args[1]), ":", 2)
		if len(names) == 2 {
			db = names[0]
			table = names[1]
		} else {
			db = defautDB
			table = names[0]
		}
	case 3:
		db = string(cmd.Args[1])
		table = string(cmd.Args[2])
	}
	id, msg := p.svc.Next(context.TODO(), db, table)
	if msg != "" {
		conn.WriteError(msg)
		return
	}
	conn.WriteInt64(id)
}

func (p *Handler) Remainder(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	name := string(cmd.Args[1])
	id, msg := p.svc.Remainder(context.TODO(), name)
	if msg != "" {
		conn.WriteError(msg)
		return
	}
	conn.WriteInt64(id)
}

func (p *Handler) Last(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	name := string(cmd.Args[1])
	id, msg := p.svc.Last(context.TODO(), name)
	if msg != "" {
		conn.WriteError(msg)
		return
	}
	conn.WriteInt64(id)
}

func (p *Handler) Max(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	name := string(cmd.Args[1])
	id, msg := p.svc.Max(context.TODO(), name)
	if msg != "" {
		conn.WriteError(msg)
		return
	}
	conn.WriteInt64(id)
}

func (p *Handler) Ping(conn redcon.Conn, cmd redcon.Command) {
	err := p.db.Ping(context.TODO())
	if err != nil {
		conn.WriteError(err.Error())
	}
	conn.WriteString("PONG")
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
	var logger klog.Logger
	logger = klog.NewLogfmtLogger(klog.NewSyncWriter(os.Stderr))
	logger = klog.With(logger, "ts", klog.DefaultTimestampUTC)

	db := store.NewStore(config.Store.Type, config.Store.URI, config.Generate.DataCenter, logger)
	svc := service.New(logger, db, config.Generate.Step, config.Generate.DataCenter, config.Store.Min, config.Store.Max)
	return &Handler{svc: svc, db: db, logger: logger}
}
