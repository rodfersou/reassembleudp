package main

import (
	"context"
	"fmt"
	"golang.org/x/exp/maps"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
	"github.com/rodfersou/reassembleudp/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}

	ctx, coll, disconnect := getMongoCollection()
	defer disconnect()

	proto := os.Getenv("PROTO")
	addr := os.Getenv("IP") + ":" + os.Getenv("PORT")
	conn, err := net.ListenPacket(proto, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	createPool(conn, coll, ctx)
}

func getMongoCollection() (context.Context, *mongo.Collection, func()) {
	ctx := context.TODO()
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	coll := client.Database("reassembleudp").Collection("payloads")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"flags", 1},
			{"created_at", 1},
			{"transaction_id", 1},
		},
	}
	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic(err)
	}

	indexModel = mongo.IndexModel{
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
	return ctx, coll, func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}
}

func createPool(conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
	var wg sync.WaitGroup
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			readUDPWorker(i, conn, coll, ctx)
		}()
	}
	go func() {
		defer wg.Done()
		reassembleMessageWorker(coll, ctx)
	}()
	wg.Wait()
}

func reassembleMessageWorker(coll *mongo.Collection, ctx context.Context) {
	mapRetry := make(map[int]bool)
	for {
		retryQueue := maps.Keys(mapRetry)
		sort.Ints(retryQueue[:])
		if len(retryQueue) > 100 {
			delete(mapRetry, retryQueue[0])
			retryQueue = maps.Keys(mapRetry)
			sort.Ints(retryQueue[:])
		}
		var first_payload models.Payload
		err := coll.FindOne(
			ctx,
			bson.M{
				"flags":          0,
				"transaction_id": bson.M{"$nin": retryQueue},
			},
			options.FindOne().SetSort(
				bson.D{
					{"transaction_id", 1},
					{"offset", 1},
				},
			),
		).Decode(&first_payload)
		if err != nil {
			// Still not ready
			continue
		}

		filter := bson.M{
			"flags":          0,
			"transaction_id": first_payload.TransactionId,
			// "created_at": bson.M{"$gte": time.Now().Add(-30 * time.Second)},
		}
		cursor, err := coll.Find(
			ctx,
			filter,
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
		var payloads []models.Payload
		if err = cursor.All(ctx, &payloads); err != nil {
			panic(err)
		}
		if len(payloads) == 0 {
			continue
		}

		flags := 0
		holes := utils.ValidateMessage(payloads)
		if len(holes) == 0 {
			flags = 1
			message := utils.ReassembleMessage(payloads)
			hash := utils.HashMessage(message)
			fmt.Printf(
				"Message #%d length: %d sha256:%s\n",
				first_payload.TransactionId,
				len(message),
				hash,
			)
		} else {
			mapRetry[first_payload.TransactionId] = true
			if payloads[0].CreatedAt.Unix() < time.Now().Add(-30*time.Second).Unix() {
				flags = 2
				fmt.Printf(
					"Message #%d Hole at: %d\n",
					first_payload.TransactionId,
					holes[0],
				)
				// for _, hole := range holes {
				//     fmt.Printf(
				//         "Message #%d Hole at: %d\n",
				//         transactionId,
				//         hole,
				//     )
				// }
			}
		}
		_, err = coll.UpdateMany(
			ctx,
			filter,
			bson.M{
				"$set": bson.M{
					"flags": flags,
				},
			},
		)
		if err != nil {
			panic(err)
		}
	}
}

func readUDPWorker(id int, conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
	payloads := make(chan *models.Payload, 1024)
	go dbInserter(id, coll, ctx, payloads)
	buf := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		payloads <- models.CreatePayload(buf)
	}
}

func dbInserter(id int, coll *mongo.Collection, ctx context.Context, payloads <-chan *models.Payload) {
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
			tid = payload.TransactionId
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
