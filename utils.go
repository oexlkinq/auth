package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func genBcryptHash(value []byte) []byte {
	hash, err := bcrypt.GenerateFromPassword(value, 10)
	if err != nil {
		panic(fmt.Errorf("gen bcrypt hash: %w", err))
	}

	return hash
}
