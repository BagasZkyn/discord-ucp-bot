package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// GeneratePIN menghasilkan PIN 6 digit acak
func GeneratePIN() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return int(n.Int64()) + 100000
}

// GenerateSalt menghasilkan salt hex 16 karakter (8 bytes)
func GenerateSalt() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HashPassword menghasilkan SHA256(password+salt) dalam uppercase hex
func HashPassword(password, salt string) string {
	h := sha256.Sum256([]byte(password + salt))
	return strings.ToUpper(fmt.Sprintf("%x", h))
}
