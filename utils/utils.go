package utils

import (
	"crypto/rand"
	"math/big"
	"time"
)

const base62Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const base62Length = 7

var maxValue = new(big.Int).Exp(big.NewInt(62), big.NewInt(base62Length), nil)


func GenerateBase62ID() (string, error) {
	now := time.Now().UnixMilli() 
	n, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return "", err
	}

	combined := now*1000 + n.Int64() 
	return base62encodedString(combined), nil
}

func base62encodedString(n int64) string {
	base := int64(62)
	encoded := make([]byte, base62Length)

	for i := base62Length - 1; i >= 0; i-- {
		remainder := n % base
		encoded[i] = base62Charset[remainder]
		n = n / base
	}

	return string(encoded)
}
