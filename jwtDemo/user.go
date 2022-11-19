package jwtdemo

type User struct {
	Username string
	Password string
	Role     string
}

const ITERATION_TIME = 3

func NewUser(username, password, role string) *User {
	hashedPassword := MD5Salt(password, username, ITERATION_TIME)
	user := &User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}
	return user
}

func (user *User) IsCorrectPassword(password string) bool {
	hashedPassword := MD5Salt(password, user.Username, ITERATION_TIME)
	return string(hashedPassword) == user.Password
}

func (user *User) Clone() *User {
	return &User{
		Username: user.Username,
		Password: user.Password,
		Role:     user.Role,
	}
}
