package hasher

import "golang.org/x/crypto/bcrypt"

type BCryptHasher struct {
	cost int
}

func NewBCryptHasher(cost int) *BCryptHasher {
	return &BCryptHasher{cost: cost}
}

func (h *BCryptHasher) Hash(plainPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), h.cost)
	return string(bytes), err
}

func (h *BCryptHasher) Compare(hash, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
}
