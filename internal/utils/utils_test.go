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
	var fragments []models.Fragment
	json.Unmarshal(content, &fragments)

	holes := ValidateMessage(fragments)
	assert.Equal(t, []int{}, holes)
}

func TestValidateInvalidMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/invalid.json")
	var fragments []models.Fragment
	json.Unmarshal(content, &fragments)

	holes := ValidateMessage(fragments)
	assert.Equal(t, []int{673}, holes)
}

func TestValidateLastNotEofMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/last_not_eof.json")
	var fragments []models.Fragment
	json.Unmarshal(content, &fragments)

	holes := ValidateMessage(fragments)
	assert.Equal(t, []int{487663}, holes)
}

func TestValidateEmptyMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/empty.json")
	var fragments []models.Fragment
	json.Unmarshal(content, &fragments)

	holes := ValidateMessage(fragments)
	assert.Equal(t, []int{0}, holes)
}

func TestReassembleMessage(t *testing.T) {
	p1 := models.Fragment{
		Offset:   0,
		DataSize: 3,
		Eof:      0,
		Data:     []int{0, 1, 2},
	}
	p2 := models.Fragment{
		Offset:   3,
		DataSize: 2,
		Eof:      0,
		Data:     []int{3, 4},
	}
	p3 := models.Fragment{
		Offset:   5,
		DataSize: 3,
		Eof:      1,
		Data:     []int{5, 6, 7},
	}
	fragments := []models.Fragment{p1, p2, p3}
	message := ReassembleMessage(fragments)
	assert.Equal(t, []byte{0, 1, 2, 3, 4, 5, 6, 7}, message)
}

func TestHashMessage(t *testing.T) {
	message := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	hash := HashMessage(message)
	assert.Equal(t, "8a851ff82ee7048ad09ec3847f1ddf44944104d2cbd17ef4e3db22c6785a0d45", hash)
}

func TestHashValidMessage(t *testing.T) {
	content, _ := ioutil.ReadFile("./fixtures/valid.json")
	var fragments []models.Fragment
	json.Unmarshal(content, &fragments)
	message := ReassembleMessage(fragments)
	hash := HashMessage(message)
	assert.Equal(t, "95e0d042cadb1106b944b49ae05097a8afd4aabd652a64cdfbc6d2f71c7090f2", hash)
}
