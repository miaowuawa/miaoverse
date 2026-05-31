package maths

import (
	"crypto/rand"
	"errors"
	"math/big"
)

func RandomIntLimited(min int, max int) (int, error) {
	if max <= min {
		return 0, errors.New("max must be greater than min")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}
