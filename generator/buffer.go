package generator

import (
	"fmt"
	"sync/atomic"
)

// A Buffer 存储分配ID (offset, max]
// 通过设置disabled为true，表示当前Buffer已经发完
type Buffer struct {
	disabled          bool
	max, offset, step int64
}

// setBuffer 设置buffer，传入最小值min和大小size，返回Buffer对象
// 这里可以分配的数字范围 (min, min+step]
func setBuffer(min, step int64, buff *Buffer) {
	buff.max = min + step
	buff.offset = min
	buff.step = step
	buff.disabled = false
}

// IsDisabled 是否不可用
func (b *Buffer) IsDisabled() bool {
	if b == nil {
		return true
	}
	return b.disabled
}

// Next 获取下一个数字, 返回id 号码，isDisabled 是否可用，err 错误，
// id如果没有错误发生，表示生成的ID。 isDisabled用于提供给ring摘除节点。
// ErrEmpty表示当前缓存中已经没有号码
// 存在边界：
//  * id == b.max，当前id可用，设置缓存不可用，没有错误；
//  * id > b.max，当前id不可用，设置缓存不可用，有错误。发生在并发取到最后一个号时，
//		只有一个goroutine 才能获取到b.max，其他的Goroutine只能获取到大于b.max；
func (b *Buffer) Next() (id int64, isDisabled bool, err error) {
	if b.IsDisabled() {
		return 0, true, ErrEmpty
	}
	id = atomic.AddInt64(&b.offset, 1)
	// 达到或者超出范围，当前buff都要设置为不可用
	if id >= b.max {
		b.SetDisabled()
		// 没有锁，并发下可能会到达这里
		if id > b.max {
			return 0, true, ErrEmpty
		}
	}
	return id, id >= b.max, nil
}

// Remainder 剩余数量
func (b *Buffer) Remainder() int64 {
	remainder := b.max - atomic.LoadInt64(&b.offset)
	if remainder > 0 {
		return remainder
	}
	return 0
}

// SetDisabled 设置Buffer不可用
func (b *Buffer) SetDisabled() {
	b.disabled = true
}

// String 打印出内部对象
func (b Buffer) String() string {
	return fmt.Sprintf("{max:%d, step:%d, offset:%d, disabled:%t}", b.max, b.step, b.offset, b.disabled)
}
