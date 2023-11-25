package utils

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rodfersou/reassembleudp/internal/models"
)

func TestValidateMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/valid.json")
	var payloads []models.Payload
	json.Unmarshal(content, &payloads)

	holes := ValidateMessage(payloads)
	assert.Equal(t, holes, []int{})
}

func TestValidateInvalidMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/invalid.json")
	var payloads []models.Payload
	json.Unmarshal(content, &payloads)

	holes := ValidateMessage(payloads)
	assert.Equal(t, holes, []int{673})
}

func TestValidateLastNotEofMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/last_not_eof.json")
	var payloads []models.Payload
	json.Unmarshal(content, &payloads)

	holes := ValidateMessage(payloads)
	assert.Equal(t, holes, []int{487663})
}

func TestValidateEmptyMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/empty.json")
	var payloads []models.Payload
	json.Unmarshal(content, &payloads)

	holes := ValidateMessage(payloads)
	assert.Equal(t, holes, []int{0})
}

func TestReassembleMessage(t *testing.T) {
	p1 := models.Payload{
		Offset:   0,
		DataSize: 3,
		Eof:      0,
		Data:     []int{0, 1, 2},
	}
	p2 := models.Payload{
		Offset:   3,
		DataSize: 2,
		Eof:      0,
		Data:     []int{3, 4},
	}
	p3 := models.Payload{
		Offset:   5,
		DataSize: 3,
		Eof:      1,
		Data:     []int{5, 6, 7},
	}
	payloads := []models.Payload{p1, p2, p3}
	message := ReassembleMessage(payloads)
	assert.Equal(t, message, []byte{0, 1, 2, 3, 4, 5, 6, 7})
}

func TestHashMessage(t *testing.T) {
	message := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	hash := HashMessage(message)
	assert.Equal(t, hash, "8a851ff82ee7048ad09ec3847f1ddf44944104d2cbd17ef4e3db22c6785a0d45")
}

func TestHashValidMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/valid.json")
	var payloads []models.Payload
	json.Unmarshal(content, &payloads)
	message := ReassembleMessage(payloads)
	hash := HashMessage(message)
	assert.Equal(t, hash, "95e0d042cadb1106b944b49ae05097a8afd4aabd652a64cdfbc6d2f71c7090f2")
}
