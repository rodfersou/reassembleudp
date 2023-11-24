package main

import (
	"context"
	// "crypto/sha256"
	"fmt"
	"math/big"
	"net"
	"os"
	// "reflect"
	"sync"

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
	var wg sync.WaitGroup

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

	proto := os.Getenv("PROTO")
	addr := os.Getenv("IP") + ":" + os.Getenv("PORT")
	conn, err := net.ListenPacket(proto, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// fmt.Println(reflect.TypeOf(conn))

	for i := 1; i <= 4; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			worker(i, conn, coll, ctx)
		}()
	}
	wg.Wait()
}

func worker(id int, conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
	inserts := make(chan *Payload, 1024)
	go db_inserter(id, coll, ctx, inserts)
	buf := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		inserts <- createPayload(buf)
	}
}

func db_inserter(id int, coll *mongo.Collection, ctx context.Context, inserts <-chan *Payload) {
	models := make([]mongo.WriteModel, 1024)
	i := 0
	for payload := range inserts {
		models[i] = mongo.NewInsertOneModel().SetDocument(
			bson.M{
				"transaction_id": payload.transaction_id,
				"offset":         payload.offset,
				"data_size":      payload.data_size,
				"eof":            payload.eof,
				"data":           payload.data,
			},
		)

		i++
		if i == 1024 {
			_, err := coll.BulkWrite(ctx, models)
			if err != nil {
				// Duplicates
				// panic(err)
			}
			fmt.Println(
				id,
				" Inserting! ",
				len(models),
			)
			i = 0
			models = make([]mongo.WriteModel, 1024)
		}
	}
	_, err := coll.BulkWrite(ctx, models)
	if err != nil {
		panic(err)
	}
	fmt.Println(
		id,
		" Inserting! ",
		len(models),
	)
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
