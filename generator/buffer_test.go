package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer_IsDisabled(t *testing.T) {
	b := new(Buffer)
	setBuffer(0, 10, b)
	assert.Equal(t, b.IsDisabled(), false)
	b.SetDisabled()
	assert.Equal(t, b.IsDisabled(), true)
}

func TestBuffer_Next(t *testing.T) {
	b := new(Buffer)
	setBuffer(0, 2, b)
	id, isEmpty, err := b.Next()
	assert.Equal(t, id, int64(1))
	assert.Equal(t, isEmpty, false)
	assert.NoError(t, err)

	id, isEmpty, err = b.Next()
	assert.Equal(t, id, int64(2))
	assert.Equal(t, isEmpty, true)

	id, isEmpty, err = b.Next()
	assert.Error(t, err, ErrEmpty)
}

func TestBuffer_Remainder(t *testing.T) {
	var min, step int64
	step = 2
	b := new(Buffer)
	setBuffer(min, step, b)
	n := b.Remainder()
	assert.Equal(t, n, step)

	b.Next()
	n = b.Remainder()
	assert.Equal(t, n, step-1)

	b.Next()
	n = b.Remainder()
	assert.Equal(t, n, step-2)

	b.Next()
	n = b.Remainder()
	assert.Equal(t, n, step-2)

}
