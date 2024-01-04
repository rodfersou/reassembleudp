package models

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFragmentEOF(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, 0, fragment.Eof)

	buf[0] = 255
	fragment = CreateFragment(buf)
	assert.Equal(t, 1, fragment.Eof)

	buf[0] = 128
	fragment = CreateFragment(buf)
	assert.Equal(t, 1, fragment.Eof)

	buf[0] = 127
	fragment = CreateFragment(buf)
	assert.Equal(t, 0, fragment.Eof)
}

func TestCreateFragmentFlags(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, 0, fragment.Flags)

	buf[0] = 128
	fragment = CreateFragment(buf)
	assert.Equal(t, 0, fragment.Flags)

	buf[0] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, 256, fragment.Flags)

	buf[0] = 0
	buf[1] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, 1, fragment.Flags)
}

func TestCreateFragmentDataSize(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, 0, fragment.DataSize)

	buf = make([]byte, 12+1)
	buf[3] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, 1, fragment.DataSize)

	buf = make([]byte, 12+256)
	buf[2] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, 256, fragment.DataSize)
}

func TestCreateFragmentOffset(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, 0, fragment.Offset)

	buf[7] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, 1, fragment.Offset)

	buf[6] = 1
	buf[7] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, 256, fragment.Offset)

	buf[5] = 1
	buf[6] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, 65536, fragment.Offset)

	buf[4] = 1
	buf[5] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, 16777216, fragment.Offset)
}

func TestCreateFragmentTransactionId(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, 0, fragment.MessageId)

	buf[11] = 1
	fragment = CreateFragment(buf)
	assert.Equal(t, 1, fragment.MessageId)

	buf[10] = 1
	buf[11] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, 256, fragment.MessageId)

	buf[9] = 1
	buf[10] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, 65536, fragment.MessageId)

	buf[8] = 1
	buf[9] = 0
	fragment = CreateFragment(buf)
	assert.Equal(t, 16777216, fragment.MessageId)
}

func TestCreateFragmentData(t *testing.T) {
	buf := make([]byte, 12)
	fragment := CreateFragment(buf)
	assert.Equal(t, []int{}, fragment.LoadData())

	// If buffer size is smaller than data_size, data is empty
	buf[3] = 3
	fragment = CreateFragment(buf)
	assert.Equal(t, []int{}, fragment.LoadData())

	buf = make([]byte, 12+3)
	buf[3] = 3
	buf[12] = 1
	buf[13] = 2
	buf[14] = 3
	fragment = CreateFragment(buf)
	assert.Equal(t, []int{1, 2, 3}, fragment.LoadData())
}
