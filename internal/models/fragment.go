package models

import (
	"math/big"
)

type Fragment struct {
	MessageId int    `bson:"message_id" json:"message_id"`
	Offset    int    `bson:"offset"     json:"offset"`
	DataSize  int    `bson:"data_size"  json:"data_size"`
	Eof       int    `bson:"eof"        json:"eof"`
	Flags     int    `bson:"flags"      json:"flags"`
	Data      []byte `bson:"data"       json:"data"`
}

func CreateFragment(buf []byte) *Fragment {
	fragment := Fragment{
		// Ignore first bit for Flags
		Flags:     int(big.NewInt(0).SetBytes([]byte{buf[0] & 127, buf[1]}).Uint64()),
		DataSize:  int(big.NewInt(0).SetBytes(buf[2:4]).Uint64()),
		Offset:    int(big.NewInt(0).SetBytes(buf[4:8]).Uint64()),
		MessageId: int(big.NewInt(0).SetBytes(buf[8:12]).Uint64()),
	}
	// Eof is the first bit
	if buf[0]&128 == 128 {
		fragment.Eof = 1
	} else {
		fragment.Eof = 0
	}

	if len(buf) >= 12+fragment.DataSize {
		fragment.Data = buf[12 : 12+fragment.DataSize]
	}

	return &fragment
}
