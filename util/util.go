package util

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var charsetMap = map[int32]int{}
var seed *rand.Rand

func init() {
	seed = rand.New(rand.NewSource(time.Now().UnixNano()))

	index := 1
	for _, rune := range charset {
		charsetMap[rune] = index
		index++
	}
}

func RandString() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func RandNumber() uint64 {
	return rand.Uint64()
}

func StringToValue(s string) int {
	var sum = 0
	for _, rune := range s {
		sum += charsetMap[rune]
	}
	return sum
}
