package main

import (
	"fmt"
	"strings"
	"syscall"

	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ssh/terminal"
	"io"
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

func creds() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, e := terminal.ReadPassword(int(syscall.Stdin))
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
	//fmt.Printf("Password: %s\n", pass)

	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pass))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	fmt.Printf("\nentities.toml input\n")
	fmt.Printf("IV:   %s\n", key)
	fmt.Printf("Salt: %s\n", salt)
	fmt.Printf("Hash: %s\n", hash)
}
