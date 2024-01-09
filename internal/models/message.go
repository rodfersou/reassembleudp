package models

import (
	"time"

	"crypto/sha256"
	"fmt"
	"golang.org/x/exp/maps"
	"sort"
)

type MessageStatus string

const (
	InProgress MessageStatus = "IN_PROGRESS"
	Valid      MessageStatus = "VALID"
	Invalid    MessageStatus = "INVALID"
)

type Message struct {
	MessageId int               `bson:"_id"        json:"message_id"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
	Status    MessageStatus     `bson:"status"     json:"status"`
	Fragments map[int]*Fragment `bson:"-"          json:"fragments"`
	Hash      string            `bson:"hash"       json:"hash"`
	Size      int               `bson:"size"       json:"size"`
	Holes     []int             `bson:"holes"      json:"holes"`
	Eof       bool              `bson:"eof"        json:"eof"`
}

func (message *Message) FastValidate() bool {
	if !message.Eof {
		return false
	}
	holes := ValidateMessage(message.Fragments, true)
	return len(holes) == 0
}

func (message *Message) AddFragment(fragment *Fragment) {
	message.Fragments[fragment.Offset] = fragment
	if fragment.Eof == 1 {
		message.Eof = true
	}
	message.Update()
}

func (message *Message) Update() {
	if message.Status != InProgress {
		return
	}
	if message.FastValidate() {
		message.Status = Valid
		data := ReassembleMessage(message.Fragments)
		message.Hash = HashMessage(data)
		message.Size = len(data)
		delete(message.Fragments, message.MessageId)
		message.Print()
	} else if message.UpdatedAt.Unix() < time.Now().Add(-30*time.Second).Unix() {
		message.Status = Invalid
		// message.Holes = ValidateMessage(message.Fragments, false)
		message.Holes = ValidateMessage(message.Fragments, true)
		delete(message.Fragments, message.MessageId)
		message.Print()
	}
}

func (message *Message) Print() {
	if message.Status == Valid {
		fmt.Printf(
			"Message #%d length: %d sha256:%s\n",
			message.MessageId,
			message.Size,
			message.Hash,
		)
		return
	} else if message.Status == Invalid {
		for _, hole := range message.Holes {
			fmt.Printf(
				"Message #%d Hole at: %d\n",
				message.MessageId,
				hole,
			)
		}
	}
}

func CreateMessage(fragment *Fragment) *Message {
	return &Message{
		MessageId: fragment.MessageId,
		UpdatedAt: time.Now(),
		Status:    InProgress,
		Fragments: make(map[int]*Fragment),
		Hash:      "",
		Holes:     make([]int, 0),
		Eof:       false,
	}
}

func ValidateMessage(fragments map[int]*Fragment, fast_fail bool) []int {
	// Empty fragment list return hole in index 0
	if len(fragments) == 0 {
		return []int{0}
	}

	// Map is unsorted
	keys := maps.Keys(fragments)
	sort.Ints(keys[:])

	// Starting the loop from the second item
	// comparing 2 in 2 items and keep track of holes
	// in the message
	holes := make([]int, 0)
	for i := 1; i < len(keys); i++ {
		one := fragments[keys[i-1]]
		two := fragments[keys[i]]
		if one.Offset+one.DataSize != two.Offset {
			holes = append(holes, one.Offset+one.DataSize)
			if fast_fail {
				return holes
			}
		}
	}

	// Last fragment need to be Eof
	last := fragments[keys[len(keys)-1]]
	if last.Eof != 1 {
		// fmt.Println(len(fragments), last.Offset, last.DataSize)
		holes = append(holes, last.Offset+last.DataSize)
	}

	return holes
}

func ReassembleMessage(fragments map[int]*Fragment) []byte {
	// Map is unsorted
	keys := maps.Keys(fragments)
	sort.Ints(keys[:])

	data := make([]byte, 0)
	for _, key := range keys {
		fragment := fragments[key]
		data = append(data, fragment.Data[:]...)
	}
	return data
}

func HashMessage(data []byte) string {
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	return hash
}
