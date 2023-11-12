package utils

import (
	"math/rand"
	"time"
)

func Now() string {
	return time.Now().Format("2006-01-02 03:04:05 PM")
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandSeq(leng int) string {
	b := make([]byte, leng)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
