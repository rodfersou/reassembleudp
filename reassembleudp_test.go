package main

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestCreatePayloadEOF(t *testing.T) {
    buf := make([]byte, 12)
    payload := createPayload(buf)
    assert.Equal(t, payload.eof, 0)

    buf[0] = 255
    payload = createPayload(buf)
    assert.Equal(t, payload.eof, 1)

    buf[0] = 128
    payload = createPayload(buf)
    assert.Equal(t, payload.eof, 1)

    buf[0] = 127
    payload = createPayload(buf)
    assert.Equal(t, payload.eof, 0)
}

func TestCreatePayloadFlags(t *testing.T) {
    buf := make([]byte, 12)
    payload := createPayload(buf)
    assert.Equal(t, payload.flags, 0)

    buf[0] = 128
    payload = createPayload(buf)
    assert.Equal(t, payload.flags, 0)

    buf[0] = 1
    payload = createPayload(buf)
    assert.Equal(t, payload.flags, 256)

    buf[0] = 0
    buf[1] = 1
    payload = createPayload(buf)
    assert.Equal(t, payload.flags, 1)
}

func TestCreatePayloadDataSize(t *testing.T) {
    buf := make([]byte, 12)
    payload := createPayload(buf)
    assert.Equal(t, payload.data_size, 0)

    buf = make([]byte, 12+1)
    buf[3] = 1
    payload = createPayload(buf)
    assert.Equal(t, payload.data_size, 1)

    buf = make([]byte, 12+256)
    buf[2] = 1
    payload = createPayload(buf)
    assert.Equal(t, payload.data_size, 256)
}

func TestCreatePayloadOffset(t *testing.T) {
    buf := make([]byte, 12)
    payload := createPayload(buf)
    assert.Equal(t, payload.offset, 0)

    buf[7] = 1
    payload = createPayload(buf)
    assert.Equal(t, payload.offset, 1)

    buf[6] = 1
    buf[7] = 0
    payload = createPayload(buf)
    assert.Equal(t, payload.offset, 256)

    buf[5] = 1
    buf[6] = 0
    payload = createPayload(buf)
    assert.Equal(t, payload.offset, 65536)

    buf[4] = 1
    buf[5] = 0
    payload = createPayload(buf)
    assert.Equal(t, payload.offset, 16777216)
}

func TestCreatePayloadTransactionId(t *testing.T) {
    buf := make([]byte, 12)
    payload := createPayload(buf)
    assert.Equal(t, payload.transaction_id, 0)

    buf[11] = 1
    payload = createPayload(buf)
    assert.Equal(t, payload.transaction_id, 1)

    buf[10] = 1
    buf[11] = 0
    payload = createPayload(buf)
    assert.Equal(t, payload.transaction_id, 256)

    buf[9] = 1
    buf[10] = 0
    payload = createPayload(buf)
    assert.Equal(t, payload.transaction_id, 65536)

    buf[8] = 1
    buf[9] = 0
    payload = createPayload(buf)
    assert.Equal(t, payload.transaction_id, 16777216)
}

func TestCreatePayloadData(t *testing.T) {
    buf := make([]byte, 12)
    payload := createPayload(buf)
    assert.Equal(t, payload.data, "")

    // If buffer size is smaller than data_size, data is empty
    buf[3] = 3
    payload = createPayload(buf)
    assert.Equal(t, payload.data, "")

    buf = make([]byte, 12+3)
    buf[3] = 3
    buf[12] = 102
    buf[13] = 111
    buf[14] = 111
    payload = createPayload(buf)
    assert.Equal(t, payload.data, "foo")
}
