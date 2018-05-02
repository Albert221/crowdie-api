package utils

import "crypto/rand"

func GenerateRandomSecret() string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 64

	bytes := make([]byte, length)
	rand.Read(bytes)

	for i, b := range bytes {
		bytes[i] = alphabet[b % byte(len(alphabet))]
	}

	return string(bytes)
}