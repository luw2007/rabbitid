// Segment支持顺序发号。每个自增ID业务场景都会在进程中生成一个Segment实例。
// 每个Segment使用circular buffer缓存多个代发的号码段。 每个发号段对应一个buffer对象。
// circular Buffer 的介绍页：https://en.wikipedia.org/wiki/Circular_buffer
package generator

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// A Segment 按照自然递增生成ID
type Segment struct {
	// dc 数据中心ID最高位，使用"|"和自增数字合并成ID
	dc int64
	// db, table 服务名称
	db, table string
	// ring 使用环缓存发号数据。减少缓存对象的生成，以及锁。
	// 通过writeCursor写入位置和readCursor读取位置来处理缓存加入和退出逻辑
	ring        []*Buffer
	writeCursor int32
	readCursor  int32
	// last 最后发号数据，可不精确
	last int64
	// step 批量导入的大小
	step int64
	// lastTimestamp 最后一次添加时间
	updateTime time.Time
}

const (
	// defaultRingSize 缓存的最大数量，需要至少 defaultExpandSize * 2，
	defaultRingSize = 64
	// defaultExpandSize 默认在ring中加载的缓存数量
	defaultExpandSize = 3
	// maxIntBits 支持最大位数 64
	maxIntBits = 64
)

const (
	// dataCenterBits 机房占位4位，最大支持32个
	dataCenterBits = uint(4)
	// DataCenterMask 机房最大数量
	DataCenterMask = int64(-1 ^ (-1 << dataCenterBits))
	// sequenceBits 发号最多支持位数，sequenceMask 表示发号器最大递增数
	sequenceBits    = maxIntBits - 1 - dataCenterBits
	dataCenterShift = sequenceBits + dataCenterBits
	sequenceMask    = int64(-1 ^ (-1 << dataCenterShift))

	segmentStringTemplate = "{dc:%d, db:%s, table:%s, last:%d, step:%d, " +
		"updateTime:%s, ring:{%s}, readCursor:%d, writeCursor:%d}"
)

var (
	// ErrTimeout 获取ID 的时候超时，发生在从存储获取失败
	ErrTimeout = errors.New("Segment Timeout ")
	// ErrEmpty 发号器空了，没有可用的号
	ErrEmpty = errors.New("buff empty")
	// ErrFull 发号器缓存满了
	ErrFull = errors.New("buff full")
	// ErrExpandDuplicated 发号器存储冲突
	ErrExpandDuplicated = errors.New("expand buff duplicated")
)

// NewSegment 新的自增ID
func NewSegment(dataCenter uint8, db, table string, step int64) *Segment {
	dc := (int64(dataCenter) & DataCenterMask) << sequenceBits
	ring := make([]*Buffer, defaultRingSize)
	for i := range ring {
		ring[i] = &Buffer{disabled: 1}
	}
	return &Segment{
		dc:         dc,
		db:         db,
		table:      table,
		ring:       ring,
		step:       step,
		updateTime: time.Now(),
	}
}

// Expand 批量加入一组号码，插入新的buffer，插入后将写游标在环中向后移动一位
// 这里没法写满整个ring， 写游标和读游标最多相差 ringSize - 1
// 发号范围 (min, max]
func (p *Segment) Expand(min, step int64) error {
	// 判断游标位置超出范围, 写满的发生的概率远小于饥饿
	if atomic.LoadInt32(&p.writeCursor)+1 >= atomic.LoadInt32(&p.readCursor)+defaultRingSize {
		return ErrFull
	}
	// 使用atomic 操作游标
	nextCursor := atomic.AddInt32(&p.writeCursor, 1)

	// atomic.AddInt32 得到的游标 -1肯定唯一
	lastCursor := (nextCursor + defaultRingSize - 1) % defaultRingSize
	if !p.ring[lastCursor].IsDisabled() {
		return ErrExpandDuplicated
	}
	p.step = step
	p.updateTime = time.Now()
	setBuffer(min, step, p.ring[lastCursor])
	return nil
}

// ExpandSize 存储的缓存数量
func (p Segment) ExpandSize() int32 {
	// size 存储的缓存数量
	size := p.writeCursor - p.readCursor
	// 如果读写都在当前节点，且可用，size+1
	if p.writeCursor == p.readCursor && !p.ring[p.readCursor%defaultRingSize].IsDisabled() {
		size++
	}
	return size
}

// Last 获取最后一次发号数据
func (p Segment) Last() (last int64) {
	cur := atomic.LoadInt32(&p.readCursor)
	b := p.ring[cur%defaultRingSize]
	return b.Last()
}

// Len 获取缓存剩余的号码量，由于没有锁，此值不精确
// 遍历可用的缓存，过滤不存在或者已经发完的缓存，然后累计剩余的号码
func (p Segment) Len() (count int64) {
	var b *Buffer
	for i := p.readCursor; i < p.writeCursor; i++ {
		b = p.ring[i%defaultRingSize]
		if !b.IsDisabled() {
			count += b.Remainder()
		}
	}
	return count
}

// Max ring中最大的值，用于切换backend
func (p Segment) Max() (max int64) {
	var b *Buffer
	for i := 0; i < defaultRingSize; i++ {
		b = p.ring[i]
		if b.max > max {
			max = b.max
		}
	}
	return max
}

// Table 获取类型名称
func (p Segment) Table() string {
	return p.table
}

// DB 获取类型名称
func (p Segment) DB() string {
	return p.db
}

// NeedExpand 检查剩余数量，判断是否需要加载更多。
// 当buff大于2个，不在加载
func (p Segment) NeedExpand() bool {
	return p.ExpandSize() < defaultExpandSize
}

// Next 获取读取游标 readCursor 位置缓存的Next值，没有错误则表示当前ID可用。
// isDisabled 表示当前缓存已发完处于不可用状态，需要移动读游标 readCursor
func (p *Segment) Next() (id int64, err error) {
	// 不存在初始化
	cur := atomic.LoadInt32(&p.readCursor)
	b := p.ring[cur%defaultRingSize]
	id, isDisabled, err := b.Next()
	if err != nil {
		return 0, err
	}
	if isDisabled {
		// Buffer.Next内部保证只有一个goroutine才能获取到b.max，
		// 所以isDisabled=true只有一个，这里增加读游标不会出现并发问题
		atomic.AddInt32(&p.readCursor, 1)
		return 0, ErrEmpty
	}
	// 合并机房标记位
	return p.dc | id&sequenceMask, nil
}

// Step 获取当前大小
func (p Segment) Step() int64 {
	return p.step
}

// String 打印出内部对象
func (p Segment) String() string {
	s := make([]string, defaultRingSize)
	for i, v := range p.ring {
		if v == nil {
			s[i] = "nil"
			continue
		}
		s[i] = v.String()
	}
	return fmt.Sprintf(segmentStringTemplate, p.dc, p.db, p.table, p.last, p.step,
		p.updateTime.String(), strings.Join(s, ","), p.readCursor,
		p.writeCursor)
}

// UpdateTime 获取更新数据时间
func (p Segment) UpdateTime() time.Time {
	return p.updateTime
}
