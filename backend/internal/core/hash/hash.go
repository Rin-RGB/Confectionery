package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Hasher struct {
	cost int
}

func NewHasher(cost int) *Hasher {
	return &Hasher{cost: cost}
}

func (hasher *Hasher) HashPassword(password []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(password, hasher.cost)
	if err != nil {
		return nil, fmt.Errorf("generate password hash: %w", err)
	}
	return hash, nil
}

func (hasher *Hasher) Compare(password, passwordHash []byte) error {
	return bcrypt.CompareHashAndPassword(passwordHash, password)
}
