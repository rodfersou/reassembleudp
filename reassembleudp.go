package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"golang.org/x/exp/maps"
	"math/big"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	coll := client.Database("reassembleudp").Collection("payloads")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"transaction_id", 1},
			{"offset", 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll.Indexes().CreateOne(ctx, indexModel)
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

	createPool(conn, coll, ctx)
}

func createPool(conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
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
	go dbInserter(id, coll, ctx, payloads)
	buf := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		payloads <- createPayload(buf)
	}
}

// var visited = make(map[int]bool)
func validateDbMessage(id int, coll *mongo.Collection, ctx context.Context, transactionIds <-chan int) {
	for transactionId := range transactionIds {
		// _, ok := visited[transactionId]
		// if ok {
		//     continue
		// }
		// visited[transactionId] = true
		cursor, err := coll.Find(
			ctx,
			bson.M{
				"transaction_id": transactionId,
			},
			options.Find().SetSort(
				bson.D{
					{"transaction_id", 1},
					{"offset", 1},
				},
			),
		)
		if err != nil {
			panic(err)
		}
		var payloads []Payload
		if err = cursor.All(ctx, &payloads); err != nil {
			panic(err)
		}

		holes := validateMessage(payloads)
		if len(holes) == 0 {
			message := reassembleMessage(payloads)
			hash := hashMessage(message)
			fmt.Println(
				"Message #",
				transactionId,
				" length: ",
				len(message),
				" sha256:",
				hash,
			)
		} else {
			for _, hole := range holes {
				fmt.Println(
					"Message #",
					transactionId,
					" Hole at: ",
					hole,
				)
			}
		}
	}
}

func dbInserter(id int, coll *mongo.Collection, ctx context.Context, payloads <-chan *Payload) {
	transactionIds := make(chan int, 1024)
	go validateDbMessage(id, coll, ctx, transactionIds)
	models := make([]mongo.WriteModel, 1024)
	i := 0
	tid := 1
	for payload := range payloads {
		models[i] = mongo.NewInsertOneModel().SetDocument(
			payload,
		)

		i++
		if i == 1024 {
			// Unordered Bulk inserts skip duplicates when the unique index raise error
			_, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
			if err != nil {
				panic(err)
			}
			if payload.TransactionId != tid {
				for i := tid; i < payload.TransactionId; i++ {
					transactionIds <- i
				}
				tid = payload.TransactionId
			}
			i = 0
			models = make([]mongo.WriteModel, 1024)
		}
	}
	if i > 0 {
		// Unordered Bulk inserts skip duplicates when the unique index raise error
		_, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
		if err != nil {
			panic(err)
		}
		fmt.Println("Last transaction ", tid)
	}
}

func createPayload(buf []byte) *Payload {
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

func validateMessage(payloads []Payload) []int {
	// Empty payload list return hole in index 0
	if len(payloads) == 0 {
		return []int{0}
	}

	// Create a map for easy lookup of offsets
	mapOffset := make(map[int]Payload)
	for _, item := range payloads {
		mapOffset[item.Offset] = item
	}

	// Map is unsorted
	keys := maps.Keys(mapOffset)
	sort.Ints(keys[:])

	// Starting the loop from the second item
	// comparing 2 in 2 items and keep track of holes
	// in the message
	holes := make([]int, 0)
	for i := 1; i < len(keys); i++ {
		one := mapOffset[keys[i-1]]
		two := mapOffset[keys[i]]
		if one.Offset+one.DataSize != two.Offset {
			holes = append(holes, one.Offset+one.DataSize)
		}
	}

	// Last fragment need to be Eof
	last := mapOffset[keys[len(keys)-1]]
	if last.Eof != 1 {
		holes = append(holes, last.Offset+last.DataSize)
	}

	return holes
}

func reassembleMessage(payloads []Payload) []byte {
	message := make([]byte, 0)
	for _, payload := range payloads {
		// Convert array of int back to array of byte
		data := make([]byte, payload.DataSize)
		for i, n := range payload.Data {
			data[i] = byte(n)
		}
		message = append(message, data[:]...)
	}
	return message
}

func hashMessage(message []byte) string {
	hash := fmt.Sprintf("%x", sha256.Sum256(message))
	return hash
}
