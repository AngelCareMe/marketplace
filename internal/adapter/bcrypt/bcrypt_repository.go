package bcrypt

type Hasher interface {
	GenerateHashPassword(password string) (string, error)
	CompareHashPassword(hash, password string) error
}
