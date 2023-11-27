package utils

import (
	"crypto/sha256"
	"fmt"
	"golang.org/x/exp/maps"
	"sort"

	"github.com/rodfersou/reassembleudp/internal/models"
)

func ValidateMessage(fragments []models.Fragment) []int {
	// Empty fragment list return hole in index 0
	if len(fragments) == 0 {
		return []int{0}
	}

	// Create a map for easy lookup of offsets
	mapOffset := make(map[int]models.Fragment)
	for _, item := range fragments {
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
		// fmt.Println(len(fragments), last.Offset, last.DataSize)
		holes = append(holes, last.Offset+last.DataSize)
	}

	return holes
}

func ReassembleMessage(fragments []models.Fragment) []byte {
	message := make([]byte, 0)
	for _, fragment := range fragments {
		// Convert array of int back to array of byte
		data := make([]byte, fragment.DataSize)
		for i, n := range fragment.Data {
			data[i] = byte(n)
		}
		message = append(message, data[:]...)
	}
	return message
}

func HashMessage(message []byte) string {
	hash := fmt.Sprintf("%x", sha256.Sum256(message))
	return hash
}
