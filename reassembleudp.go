package main

import (
	"context"
	// "crypto/sha256"
	"fmt"
	"math/big"
	"net"
	"os"
	// "sync"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Payload struct {
	eof            int
	flags          int
	data_size      int
	offset         int
	transaction_id int
	data           []int
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}

	jobs := make(chan []byte, 100)
	for i := 1; i <= 4; i++ {
		go worker(i, jobs)
	}

	proto := os.Getenv("PROTO")
	addr := os.Getenv("IP") + ":" + os.Getenv("PORT")
	conn, err := net.ListenPacket(proto, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		jobs <- buf
	}
}

func worker(id int, jobs <-chan []byte) {
	ctx := context.TODO()
	uri := os.Getenv("MONGO_URI")
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
	// client.Database("reassembleudp").Drop(ctx)
	coll := client.Database("reassembleudp").Collection("messages")

	// index_model := mongo.IndexModel{
	//     Keys: bson.D{
	//         {"_id", 1},
	//         {"payloads.offset", 1},
	//     },
	//     Options: options.Index().SetUnique(true),
	// }
	// _, err = coll.Indexes().CreateOne(ctx, index_model)
	// if err != nil {
	//     panic(err)
	// }

	for buf := range jobs {
		payload := createPayload(buf)
		_, err = coll.UpdateOne(
			ctx,
			bson.M{
				"_id": payload.transaction_id,
			},
			bson.M{
				"$addToSet": bson.M{
					"payloads": bson.M{
						"offset":    payload.offset,
						"data_size": payload.data_size,
						"eof":       payload.eof,
						"data":      payload.data,
					},
				},
			},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			panic(err)
		}
		fmt.Println(
			"Inserting message: ",
			id,
			payload.transaction_id,
			payload.offset,
			payload.data_size,
			payload.eof,
			// buf[:13],
		)
		// fmt.Println(result)
	}
}

func createPayload(buf []byte) *Payload {
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
	return &payload
}
