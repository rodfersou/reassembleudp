package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatePayloadEOF(t *testing.T) {
	buf := make([]byte, 12)
	payload := createPayload(buf)
	assert.Equal(t, payload.Eof, 0)

	buf[0] = 255
	payload = createPayload(buf)
	assert.Equal(t, payload.Eof, 1)

	buf[0] = 128
	payload = createPayload(buf)
	assert.Equal(t, payload.Eof, 1)

	buf[0] = 127
	payload = createPayload(buf)
	assert.Equal(t, payload.Eof, 0)
}

func TestCreatePayloadFlags(t *testing.T) {
	buf := make([]byte, 12)
	payload := createPayload(buf)
	assert.Equal(t, payload.Flags, 0)

	buf[0] = 128
	payload = createPayload(buf)
	assert.Equal(t, payload.Flags, 0)

	buf[0] = 1
	payload = createPayload(buf)
	assert.Equal(t, payload.Flags, 256)

	buf[0] = 0
	buf[1] = 1
	payload = createPayload(buf)
	assert.Equal(t, payload.Flags, 1)
}

func TestCreatePayloadDataSize(t *testing.T) {
	buf := make([]byte, 12)
	payload := createPayload(buf)
	assert.Equal(t, payload.DataSize, 0)

	buf = make([]byte, 12+1)
	buf[3] = 1
	payload = createPayload(buf)
	assert.Equal(t, payload.DataSize, 1)

	buf = make([]byte, 12+256)
	buf[2] = 1
	payload = createPayload(buf)
	assert.Equal(t, payload.DataSize, 256)
}

func TestCreatePayloadOffset(t *testing.T) {
	buf := make([]byte, 12)
	payload := createPayload(buf)
	assert.Equal(t, payload.Offset, 0)

	buf[7] = 1
	payload = createPayload(buf)
	assert.Equal(t, payload.Offset, 1)

	buf[6] = 1
	buf[7] = 0
	payload = createPayload(buf)
	assert.Equal(t, payload.Offset, 256)

	buf[5] = 1
	buf[6] = 0
	payload = createPayload(buf)
	assert.Equal(t, payload.Offset, 65536)

	buf[4] = 1
	buf[5] = 0
	payload = createPayload(buf)
	assert.Equal(t, payload.Offset, 16777216)
}

func TestCreatePayloadTransactionId(t *testing.T) {
	buf := make([]byte, 12)
	payload := createPayload(buf)
	assert.Equal(t, payload.TransactionId, 0)

	buf[11] = 1
	payload = createPayload(buf)
	assert.Equal(t, payload.TransactionId, 1)

	buf[10] = 1
	buf[11] = 0
	payload = createPayload(buf)
	assert.Equal(t, payload.TransactionId, 256)

	buf[9] = 1
	buf[10] = 0
	payload = createPayload(buf)
	assert.Equal(t, payload.TransactionId, 65536)

	buf[8] = 1
	buf[9] = 0
	payload = createPayload(buf)
	assert.Equal(t, payload.TransactionId, 16777216)
}

func TestCreatePayloadData(t *testing.T) {
	buf := make([]byte, 12)
	payload := createPayload(buf)
	assert.Equal(t, payload.Data, []int{})

	// If buffer size is smaller than data_size, data is empty
	buf[3] = 3
	payload = createPayload(buf)
	assert.Equal(t, payload.Data, []int{})

	buf = make([]byte, 12+3)
	buf[3] = 3
	buf[12] = 1
	buf[13] = 2
	buf[14] = 3
	payload = createPayload(buf)
	assert.Equal(t, payload.Data, []int{1, 2, 3})
}
