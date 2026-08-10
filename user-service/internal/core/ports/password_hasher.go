package ports

type PasswordHasher interface {
	Hash(plainPassword string) (string, error)
	Compare(hash, plainPassword string) error
}
