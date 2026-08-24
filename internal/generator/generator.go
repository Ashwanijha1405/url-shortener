package generator

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

const DefaultLength = 7

func Generate(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}

	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(alphabet))),
		)
		if err != nil {
			return "", err
		}

		result[i] = alphabet[n.Int64()]
	}

	return string(result), nil
}
