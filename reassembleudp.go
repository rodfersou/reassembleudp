package main

import (
	"context"
	"net"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/pool"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}

	conn, disconnect := getUDPConnection()
	defer disconnect()

	ctx, coll_messages, coll_fragments, disconnect := getMongoCollection()
	defer disconnect()

	pool.CreatePool(conn, coll_messages, coll_fragments, ctx)
}

func getUDPConnection() (net.PacketConn, func()) {
	proto := os.Getenv("PROTO")
	addr := os.Getenv("IP") + ":" + os.Getenv("PORT")
	conn, err := net.ListenPacket(proto, addr)
	if err != nil {
		panic(err)
	}
	return conn, func() {
		conn.Close()
	}
}

func getMongoCollection() (
	context.Context,
	*mongo.Collection,
	*mongo.Collection,
	func(),
) {
	ctx := context.TODO()
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	coll_messages := client.Database("reassembleudp").Collection("messages")
	coll_fragments := client.Database("reassembleudp").Collection("fragments")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"message_id", 1},
			{"offset", 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll_fragments.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic(err)
	}

	return ctx, coll_messages, coll_fragments, func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}
}
