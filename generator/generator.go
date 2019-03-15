// Package generator 发号包，目前实现顺序发号Segment
package generator

import "time"

// A Generator 发号器
type Generator interface {
	Expand(int64, int64) error
	Last() int64
	Len() int64
	Max() int64
	DB() string
	Table() string
	NeedExpand() bool
	Next() (int64, error)
	Step() int64
	UpdateTime() time.Time
}
