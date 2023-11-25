package models

import (
	"math/big"
	"time"
)

type Payload struct {
	TransactionId int       `bson:"transaction_id" json:"transaction_id"`
	Offset        int       `bson:"offset"         json:"offset"`
	DataSize      int       `bson:"data_size"      json:"data_size"`
	Eof           int       `bson:"eof"            json:"eof"`
	Flags         int       `bson:"flags"          json:"flags"`
	Data          []int     `bson:"data"           json:"data"`
	CreatedAt     time.Time `bson:"created_at"     json:"created_at"`
}

func CreatePayload(buf []byte) *Payload {
	payload := Payload{
		// Ignore first bit for Flags
		Flags:         int(big.NewInt(0).SetBytes([]byte{buf[0] & 127, buf[1]}).Uint64()),
		DataSize:      int(big.NewInt(0).SetBytes(buf[2:4]).Uint64()),
		Offset:        int(big.NewInt(0).SetBytes(buf[4:8]).Uint64()),
		TransactionId: int(big.NewInt(0).SetBytes(buf[8:12]).Uint64()),
		CreatedAt:     time.Now(),
	}
	// Eof is the first bit
	if buf[0]&128 == 128 {
		payload.Eof = 1
	} else {
		payload.Eof = 0
	}

	// Convert Data []byte to []int for easy lookup at DB
	payload.Data = []int{}
	if len(buf) >= 12+payload.DataSize {
		payload.Data = make([]int, payload.DataSize)
		for i, n := range buf[12 : 12+payload.DataSize] {
			payload.Data[i] = int(n)
		}
	}

	return &payload
}
