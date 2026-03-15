package users

type UserStorageInterface interface {
	Add(user User) error
	FindByUsername(username string) (*User, error)
	GetAll() ([]User, error)
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
