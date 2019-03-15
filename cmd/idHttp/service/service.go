// Package service 生成ID
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	klog "github.com/go-kit/kit/log"
	"github.com/petermattis/goid"

	"github.com/luw2007/rabbitid/generator"
	"github.com/luw2007/rabbitid/store"
)

// A Service 发号器接口
type Service interface {
	// NextID 通过服务名称获取自增ID和错误
	Next(ctx context.Context, db, table string) (id int64, msg string)
	// Last 通过服务名获取最后发放的ID和错误
	Last(ctx context.Context, name string) (id int64, msg string)
	// Last 通过服务名获取剩余的ID数量和错误
	Remainder(ctx context.Context, name string) (id int64, msg string)
	// Max 通过服务名获取可生成的最大ID和错误
	Max(ctx context.Context, name string) (id int64, msg string)
}

// A service 递增生成ID
type service struct {
	Generator  *sync.Map
	Store      store.Store
	Step       int64
	DataCenter uint8
	// minBufferTime, maxBufferTime 表示缓存最长和最短支持时间，用来调整每次缓存数量
	minBufferTime time.Duration
	maxBufferTime time.Duration
	log           klog.Logger
}

const (
	// retries 重试次数
	retries = 5
	// processTaskTicker 后台任务间隔
	processTaskTicker = time.Millisecond * 50
	// 默认存储获取超时时间
	defaultGeneratorLoadTimeout = time.Millisecond * 200
)

// ErrEmpty 查询ID 的类型不存在
var ErrEmpty = errors.New("类型不存在")

// New 生成新的ID服务
func New(logger klog.Logger, store store.Store, size int64, dc uint8, min, max time.Duration) Service {
	if int64(dc) > generator.DataCenterMask {
		log.Fatalln("dateCenter critical:", dc)
	}
	logger = klog.With(logger, "svc", "id", "gid", goid.Get(), "dataCenter", dc)
	service := &service{
		Generator:     new(sync.Map),
		Store:         store,
		Step:          size,
		DataCenter:    dc,
		minBufferTime: min,
		maxBufferTime: max,
		log:           logger,
	}
	logger.Log("new id service")
	go service.process()
	return service
}

// expand 加载更多的数据
func (p *service) expand(ctx context.Context, g generator.Generator) (int64, error) {
	size := p.newSize(g)
	min, err := p.Store.Range(ctx, p.DataCenter, g.DB(), g.Table(), size)
	if err != nil {
		return 0, err
	}
	if err = g.Expand(min, size); err != nil {
		p.log.Log("action", "expand", "err", err.Error())
	}
	return min, nil
}

// Last 获取上次分配的ID
func (p *service) Last(ctx context.Context, name string) (int64, string) {
	gs, ok := p.Generator.Load(name)
	if !ok {
		return 0, ErrEmpty.Error()
	}
	g := gs.(generator.Generator)
	return g.Last(), ""
}

// newSize 重新计算size
func (p service) newSize(g generator.Generator) int64 {
	size := g.Step()
	duration := time.Since(g.UpdateTime())

	// [0, minBufferTime) 表示当前消费者饥饿，增加获取数量
	if duration < p.minBufferTime {
		size *= 2
		// [maxBufferTime, ∞) 表示当前消费者饱和，减少获取数量
	} else if duration > p.maxBufferTime {
		size /= 2
	}
	// 每次加载不小于初始值
	if size < p.Step || size > p.Step*1024 {
		size = g.Step()
	}
	return size
}

// NextID 获取新的ID, 没有初始化从store中获取
func (p *service) Next(ctx context.Context, db, table string) (v int64, msg string) {
	var g generator.Generator
	name := fmt.Sprintf("%s|%s", db, table)
	gs, ok := p.Generator.Load(name)
	// 不存在初始化
	if ok {
		g = gs.(generator.Generator)
	} else {
		g = generator.NewSegment(p.DataCenter, db, table, p.Step)
		// 防止竞争生成多个generator
		old, loaded := p.Generator.LoadOrStore(name, g)
		if loaded {
			g = old.(generator.Generator)
		} else {
			p.log.Log("db", db, "table", table, "expand", "init")
			p.expand(ctx, g)
		}
	}
	var err error
	for i := 0; i < retries; i++ {
		v, err = g.Next()
		switch err {
		case nil:
			return v, ""
		case generator.ErrEmpty:
			// 可用数据为空的时候再检查一次，防止并发导致多次expand
			if g.NeedExpand() {
				p.log.Log("name", name, "expand", "front", "cause", "empty")
				p.expand(ctx, g)
			}
			continue
		default:
			p.log.Log("name", name, "expand", "front", "cause", err.Error(), "len", g.Len())
			return v, err.Error()
		}
	}
	p.log.Log("name", name, "expand", "front", "cause", err.Error())
	return v, err.Error()
}

// Remainder 余数
func (p *service) Remainder(ctx context.Context, name string) (int64, string) {
	gs, ok := p.Generator.Load(name)
	if !ok {
		return 0, ErrEmpty.Error()
	}
	g := gs.(generator.Generator)
	return g.Len(), ""
}

// Max 可生成的最大值
func (p *service) Max(ctx context.Context, name string) (int64, string) {
	gs, ok := p.Generator.Load(name)
	if !ok {
		return 0, ErrEmpty.Error()
	}
	g := gs.(generator.Generator)
	return g.Max(), ""
}

// process 后台任务，加载更多数据
func (p *service) process() {
	p.log.Log("msg", "process start")
	for {
		time.Sleep(processTaskTicker)
		ctx, cancel := context.WithTimeout(context.Background(), defaultGeneratorLoadTimeout)
		err := p.Store.Ping(ctx)
		if err != nil {
			p.log.Log("expand", "ping", "err", err.Error())
		}
		cancel()
		p.Generator.Range(func(key, value interface{}) bool {
			g := value.(generator.Generator)
			if p.Store.BlockDB(p.DataCenter, g.DB()) {
				return true
			}
			// 当剩余数量多少的时候，再次填充
			if !g.NeedExpand() {
				return true
			}
			ctx, cancel := context.WithTimeout(context.Background(), defaultGeneratorLoadTimeout)
			p.log.Log("db", g.DB(), "table", g.Table(), "expand", "process", "len", g.Len(), "size", g.Step(), "last", g.Last())
			_, err := p.expand(ctx, g)
			cancel()
			if err != nil {
				p.log.Log("expand", "expand", "err", err.Error())
			}
			return true
		})
	}
}
