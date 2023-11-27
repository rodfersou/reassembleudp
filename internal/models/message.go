package models

import (
	"time"
)

type MessageStatus string

const (
	InProgress MessageStatus = "IN_PROGRESS"
	Valid      MessageStatus = "VALID"
	Invalid    MessageStatus = "INVALID"
)

// type MessageStatus int
// const (
//     InProgress MessageStatus = iota
//     Valid
//     Invalid
// )
// mapStringsMessageStatus := map[string]MessageStatus{
//     "IN_PROGRESS": InProgress,
//     "VALID":       Valid,
//     "INVALID":     Invalid,
// }
// func (status MessageStatus) String() string {
//     return messageStatusString[status]
// }
// essageStatus

type Message struct {
	MessageId int           `bson:"_id"        json:"message_id"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
	Status    MessageStatus `bson:"status"     json:"status"`
	Data      []int         `bson:"data"       json:"data"`
	Holes     []int         `bson:"holes"      json:"holes"`
	CheckSum  string        `bson:"checksum"   json:"checksum"`
}

func CreateMessage(id int) *Message {
	return &Message{
		MessageId: id,
		UpdatedAt: time.Now(),
		Status:    InProgress,
		Data:      []int{},
		Holes:     []int{},
		CheckSum:  "",
	}
}
