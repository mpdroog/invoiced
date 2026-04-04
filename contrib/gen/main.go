// Package main generates authentication credentials (IV, Salt, Hash, APIKey) for entities.toml.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// NewEncryptionKey generates a random 256-bit key for Encrypt() and
// Decrypt(). It panics if the source of randomness fails.
// https://raw.githubusercontent.com/gtank/cryptopasta/master/encrypt.go
func NewEncryptionKey() *[32]byte {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}

// NewAPIKey generates a random 32-byte API key encoded in base64.
func NewAPIKey() string {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func creds() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, e := term.ReadPassword(syscall.Stdin)
	if e != nil {
		return "", e
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

func main() {
	rnd := fmt.Sprintf("%x", NewEncryptionKey())
	key := rnd[0:32]
	salt := rnd[32:64]

	pass, e := creds()
	if e != nil {
		panic(e)
	}
	// fmt.Printf("Password: %s\n", pass)

	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pass))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	apiKey := NewAPIKey()

	fmt.Printf("\nentities.toml input\n")
	fmt.Printf("IV:     %s\n", key)
	fmt.Printf("Salt:   %s\n", salt)
	fmt.Printf("Hash:   %s\n", hash)
	fmt.Printf("APIKey: %s\n", apiKey)
}
