package generator

import (
	"fmt"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testDC          = 0
	testSize  int64 = 10
	testDB          = "test"
	testTable       = "test_1"
)

func ExampleNewSegment() {
	seg := NewSegment(testDC, testDB, testTable, 10)
	seg.Expand(0, 10)
	fmt.Println(seg.Len())
	// Output: 10
}

func TestSegment_Expand(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	seg.Expand(0, testSize)
	assert.Equal(t, seg.Len(), testSize)
	got, err := seg.Next()
	assert.NoError(t, err)

	want := int64(1)
	assert.Equal(t, got, want)
	for i := 0; i < int(testSize-1); i++ {
		seg.Next()
	}
	assert.Equal(t, seg.Len(), int64(0))
}

func TestSegment_Expand2(t *testing.T) {
	var testDC2 uint8 = 1
	seg := NewSegment(testDC2, testDB, testTable, testSize)
	seg.Expand(0, testSize)
	assert.Equal(t, seg.Len(), testSize)
	want := (int64(testDC2)&DataCenterMask)<<sequenceBits + 1
	got, err := seg.Next()
	assert.NoError(t, err)
	assert.Equal(t, got, want)
}

func TestSegment_ExpandSize(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)

	// 生产则增多
	seg.Expand(0, 2)
	assert.Equal(t, seg.ExpandSize(), int32(1))

	// 消费则减少
	seg.Next()
	seg.Next()
	assert.Equal(t, seg.ExpandSize(), int32(0))

	// 生产数据最多为defaultRingSize -1
	for i := 0; i < int(defaultRingSize-1); i++ {
		seg.Expand(2, 2)
	}
	assert.Equal(t, seg.ExpandSize(), int32(defaultRingSize-1))
	seg.Expand(2, 2)
	err := seg.Expand(2, 2)
	assert.Error(t, err, ErrFull.Error())
	assert.Equal(t, seg.ExpandSize(), int32(defaultRingSize-1))
}

func TestSegment_Last(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	seg.Expand(0, testSize)
	assert.Equal(t, seg.Len(), testSize)
	got, err := seg.Next()
	assert.NoError(t, err)
	assert.Equal(t, got, seg.Last())
}

func TestSegment_NeedExpand(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	assert.True(t, seg.NeedExpand())
	seg.Expand(0, testSize)
	seg.Expand(0, testSize)
	seg.Expand(0, testSize)
	assert.False(t, seg.NeedExpand())
}

func TestSegment_Next(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	_, err := seg.Next()
	assert.EqualError(t, err, ErrEmpty.Error())
	seg.Expand(0, testSize)
	want := int64(1)
	got, err := seg.Next()
	assert.NoError(t, err)
	assert.Equal(t, got, want)
}

func TestSegment_Next2(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)

	id, err := seg.Next()
	assert.Error(t, ErrEmpty)

	seg.Expand(0, 2)
	seg.Expand(2, 2)

	// 1
	id, err = seg.Next()
	assert.NoError(t, err)
	assert.Equal(t, id, int64(1))

	// 2
	seg.Next()

	// 3
	id, err = seg.Next()
	assert.NoError(t, err)
	assert.Equal(t, id, int64(3))

	// 4
	seg.Next()

	// err
	id, err = seg.Next()
	assert.Error(t, ErrEmpty)
}

func TestSegment_Step(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)

	seg.Expand(0, testSize)
	assert.Equal(t, seg.Step(), testSize)

	seg.Expand(0, testSize*2)
	assert.Equal(t, seg.Step(), testSize*2)
}

func TestSegment_UpdateTime(t *testing.T) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	now := time.Now()
	seg.Expand(0, testSize)
	utime := seg.UpdateTime()
	if now.After(utime) {
		t.Error("更新updateTime失败")
	}
}

func BenchmarkSegment_Next(b *testing.B) {
	var step int64 = 10000
	seg := NewSegment(testDC, testDB, testTable, testSize)
	seg.Expand(int64(0), step)
	for i := 0; i < b.N; i++ {
		//use b.N for looping
		if int64(i)%step == 0 {
			seg.Expand(int64(i)*step, step)
		}
		seg.Next()
	}
}

func BenchmarkSegment_Expand(b *testing.B) {
	var step int64 = 100
	seg := NewSegment(testDC, testDB, testTable, testSize)
	seg.Expand(int64(0), step)
	for i := 0; i < b.N; i++ {
		//use b.N for looping
		seg.Expand(int64(i)*step, step)
		seg.readCursor++
	}
}

func BenchmarkSegment_NeedExpand_Empty(b *testing.B) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	for i := 0; i < b.N; i++ {
		//use b.N for looping
		seg.NeedExpand()
	}
}

func BenchmarkSegment_NeedExpand_Full(b *testing.B) {
	seg := NewSegment(testDC, testDB, testTable, testSize)
	for i := 0; i < defaultExpandSize; i++ {
		seg.Expand(int64(0), 2)
	}
	for i := 0; i < b.N; i++ {
		//use b.N for looping
		seg.NeedExpand()
	}
}
