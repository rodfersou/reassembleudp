package models

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFragmentEOF(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, fragment.Eof, 0)

	buf[0] = 255
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Eof, 1)

	buf[0] = 128
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Eof, 1)

	buf[0] = 127
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Eof, 0)
}

func TestCreateFragmentFlags(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, fragment.Flags, 0)

	buf[0] = 128
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Flags, 0)

	buf[0] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Flags, 256)

	buf[0] = 0
	buf[1] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Flags, 1)
}

func TestCreateFragmentDataSize(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, fragment.DataSize, 0)

	buf = make([]byte, 12+1)
	buf[3] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.DataSize, 1)

	buf = make([]byte, 12+256)
	buf[2] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.DataSize, 256)
}

func TestCreateFragmentOffset(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, fragment.Offset, 0)

	buf[7] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Offset, 1)

	buf[6] = 1
	buf[7] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Offset, 256)

	buf[5] = 1
	buf[6] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Offset, 65536)

	buf[4] = 1
	buf[5] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Offset, 16777216)
}

func TestCreateFragmentTransactionId(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, fragment.MessageId, 0)

	buf[11] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.MessageId, 1)

	buf[10] = 1
	buf[11] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.MessageId, 256)

	buf[9] = 1
	buf[10] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.MessageId, 65536)

	buf[8] = 1
	buf[9] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.MessageId, 16777216)
}

func TestCreateFragmentData(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, fragment.Data, []int{})

	// If buffer size is smaller than data_size, data is empty
	buf[3] = 3
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Data, []int{})

	buf = make([]byte, 12+3)
	buf[3] = 3
	buf[12] = 1
	buf[13] = 2
	buf[14] = 3
	fragment = CreateFragment(buf)
	assert.Equal(t, fragment.Data, []int{1, 2, 3})
}
