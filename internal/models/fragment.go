package models

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

type Fragment struct {
	MessageId int `bson:"message_id" json:"message_id"`
	Offset    int `bson:"offset"     json:"offset"`
	DataSize  int `bson:"data_size"  json:"data_size"`
	Eof       int `bson:"eof"        json:"eof"`
	Flags     int `bson:"flags"      json:"flags"`
}

func (fragment *Fragment) SaveData(data []byte) {
	// if dir don't exist, create
	files, err := filepath.Glob(fmt.Sprintf("inbox/%04d", fragment.MessageId))
	if err != nil {
		panic(err)
	}
	dir := ""
	if len(files) == 0 {
		dir = fmt.Sprintf("inbox/%04d", fragment.MessageId)
		if err = os.Mkdir(dir, os.ModePerm); err != nil {
			// concurrent dir creation
			// panic(err)
		}
	} else {
		dir = files[0]
	}

	// if file don't exist, create
	filename := fmt.Sprintf("%s/%010d", dir, fragment.Offset)
	files, err = filepath.Glob(filename)
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		err := os.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func (fragment *Fragment) LoadData() []byte {
	dir := fmt.Sprintf("inbox/%04d", fragment.MessageId)
	filename := fmt.Sprintf("%s/%010d", dir, fragment.Offset)
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
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
		fragment.SaveData(buf[12 : 12+fragment.DataSize])
	}

	return &fragment
}
