package pool

import (
	"context"
	"fmt"
	// "sort"
	// "sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
	"github.com/rodfersou/reassembleudp/internal/utils"
)

func ReassembleMessageWorker(
	coll *mongo.Collection,
	ctx context.Context,
	// receivingMessage *sync.Map,
) {
	for {
		type Pair struct {
			Key   int
			Value int64
		}
		var oldestMessage []Pair
		// (*receivingMessage).Range(func(k, v interface{}) bool {
		//  intKey, ok := k.(int)
		//  if !ok {
		//      return false
		//  }
		//  int64Value, ok := v.(int64)
		//  if !ok {
		//      return false
		//  }
		//  oldestMessage = append(oldestMessage, Pair{intKey, int64Value})
		//  return true
		// })
		// if len(oldestMessage) == 0 {
		//  time.Sleep(1 * time.Second)
		//  continue
		// }
		// sort.Slice(oldestMessage, func(i, j int) bool {
		//  iv, jv := oldestMessage[i], oldestMessage[j]
		//  switch {
		//  case iv.Value != jv.Value:
		//      return iv.Value < jv.Value
		//  default:
		//      return iv.Key < jv.Key
		//  }
		// })
		for _, kv := range oldestMessage {
			messageId := kv.Key
			lastReceived := kv.Value
			filter := bson.M{
				"message_id": messageId,
			}
			cursor, err := coll.Find(
				ctx,
				filter,
				options.Find().SetSort(
					bson.D{
						{"message_id", 1},
						{"offset", 1},
					},
				),
			)
			if err != nil {
				panic(err)
			}
			var fragments []models.Fragment
			if err = cursor.All(ctx, &fragments); err != nil {
				panic(err)
			}
			if len(fragments) == 0 {
				continue
			}

			holes := utils.ValidateMessage(fragments)
			if len(holes) == 0 {
				// (*receivingMessage).Delete(messageId)
				message := utils.ReassembleMessage(fragments)
				hash := utils.HashMessage(message)
				fmt.Printf(
					"Message #%d length: %d sha256:%s\n",
					messageId,
					len(message),
					hash,
				)
			} else {
				if lastReceived < time.Now().Add(-30*time.Second).Unix() {
					// (*receivingMessage).Delete(messageId)
					fmt.Printf(
						"Message #%d Hole at: %d\n",
						messageId,
						holes[0],
					)
					// for _, hole := range holes {
					//     fmt.Printf(
					//         "Message #%d Hole at: %d\n",
					//         messageId,
					//         hole,
					//     )
					// }
				}
			}
		}
	}
}
