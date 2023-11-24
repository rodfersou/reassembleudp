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
	TransactionId int   `bson:"transaction_id"`
	Offset        int   `bson:"offset"`
	DataSize      int   `bson:"data_size"`
	Eof           int   `bson:"eof"`
	Flags         int   `bson:"flags"`
	Data          []int `bson:"data"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}

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

	create_pool(conn, coll, ctx)
}

func create_pool(conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
	var wg sync.WaitGroup
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
	payloads := make(chan *Payload, 1024)
	go db_inserter(id, coll, ctx, payloads)
	buf := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		payloads <- createPayload(buf)
	}
}

func db_inserter(id int, coll *mongo.Collection, ctx context.Context, payloads <-chan *Payload) {
	models := make([]mongo.WriteModel, 1024)
	i := 0
	tid := 1
	for payload := range payloads {
		models[i] = mongo.NewInsertOneModel().SetDocument(
			payload,
		)

		i++
		if i == 1024 {
			_, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
			if err != nil {
				panic(err)
			}
			if payload.TransactionId != tid {
				fmt.Println("Finished transactions from ", tid, "to ", payload.TransactionId-1)
				tid = payload.TransactionId
			}
			// fmt.Println(
			//  id,
			//  " Inserting! ",
			//  len(models),
			// )
			i = 0
			models = make([]mongo.WriteModel, 1024)
		}
	}
	if i > 0 {
		_, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
		if err != nil {
			panic(err)
		}
		// fmt.Println(
		//  id,
		//  " Inserting! ",
		//  len(models),
		// )
		fmt.Println("Last transaction ", tid)
	}
}

func createPayload(buf []byte) *Payload {
	payload := Payload{
		Flags:         int(big.NewInt(0).SetBytes([]byte{buf[0] & 127, buf[1]}).Uint64()),
		DataSize:      int(big.NewInt(0).SetBytes(buf[2:4]).Uint64()),
		Offset:        int(big.NewInt(0).SetBytes(buf[4:8]).Uint64()),
		TransactionId: int(big.NewInt(0).SetBytes(buf[8:12]).Uint64()),
	}
	if buf[0]&128 == 128 {
		payload.Eof = 1
	} else {
		payload.Eof = 0
	}
	payload.Data = []int{}
	if len(buf) >= 12+payload.DataSize {
		payload.Data = make([]int, payload.DataSize)
		for i, n := range buf[12 : 12+payload.DataSize] {
			payload.Data[i] = int(n)
		}
	}
	return &payload
}
