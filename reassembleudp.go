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

	ctx, coll, disconnect := getMongoCollection()
	defer disconnect()

	pool.CreatePool(conn, coll, ctx)
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

func getMongoCollection() (context.Context, *mongo.Collection, func()) {
	ctx := context.TODO()
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	coll := client.Database("reassembleudp").Collection("fragments")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"message_id", 1},
			{"offset", 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic(err)
	}

	// indexModel = mongo.IndexModel{
	//  Keys: bson.D{
	//      {"flags", 1},
	//      {"created_at", 1},
	//      {"message_id", 1},
	//  },
	// }
	// _, err = coll.Indexes().CreateOne(ctx, indexModel)
	// if err != nil {
	//  panic(err)
	// }

	return ctx, coll, func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}
}
