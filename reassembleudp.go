package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	// "crypto/sha256"

	"github.com/libp2p/go-reuseport"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const proto, port = "udp", "6789"

type Payload struct {
	eof            int
	flags          int
	data_size      int
	offset         int
	transaction_id int
	data           []int
}

func main() {
	ctx := context.TODO()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	// TODO: Remove database drop
	client.Database("reassembleudp").Drop(ctx)
	coll := client.Database("reassembleudp").Collection("payloads")

	index_model := mongo.IndexModel{
		Keys: bson.D{
			{"transaction_id", 1},
			{"offset", 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll.Indexes().CreateOne(ctx, index_model)
	if err != nil {
		panic(err)
	}

	conn, err := reuseport.ListenPacket(proto, os.Getenv("IP")+":"+port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		_, dst, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		conn.WriteTo(buf, dst)

		payload := createPayload(buf)
		result, err := coll.InsertOne(ctx, bson.D{
			{"eof", payload.eof},
			{"flags", payload.flags},
			{"data_size", payload.data_size},
			{"offset", payload.offset},
			{"transaction_id", payload.transaction_id},
			{"data", payload.data},
		})
		if err != nil {
			panic(err)
		}
		// fmt.Println(payload)
		fmt.Println(result)
	}
}

func createPayload(buf []byte) Payload {
	payload := Payload{
		flags:          int(big.NewInt(0).SetBytes([]byte{buf[0] & 127, buf[1]}).Uint64()),
		data_size:      int(big.NewInt(0).SetBytes(buf[2:4]).Uint64()),
		offset:         int(big.NewInt(0).SetBytes(buf[4:8]).Uint64()),
		transaction_id: int(big.NewInt(0).SetBytes(buf[8:12]).Uint64()),
	}
	if buf[0]&128 == 128 {
		payload.eof = 1
	} else {
		payload.eof = 0
	}
	payload.data = []int{}
	if len(buf) >= 12+payload.data_size {
		payload.data = make([]int, payload.data_size)
		for i, n := range buf[12 : 12+payload.data_size] {
			payload.data[i] = int(n)
		}
	}
	return payload
}
