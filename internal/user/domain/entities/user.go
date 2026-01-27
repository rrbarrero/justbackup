package entities

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

func NewUser(username, passwordHash string) *User {
	return &User{
		Username:     username,
		PasswordHash: passwordHash,
	}
}
