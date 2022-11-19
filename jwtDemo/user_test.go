package jwtdemo

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	hashedPassword := MD5Salt("123456", "cheryl", 3)
	fmt.Println(string(hashedPassword))
}

func TestLogin(t *testing.T) {
	store := NewInMemoryUserStore()
	user := NewUser("cheryl", "123456", "guest")
	store.Save(user)
	usr, err := store.Find("cheryl")
	assert.NoError(t, err)
	assert.NotNil(t, usr)
	ans := usr.IsCorrectPassword("123456")
	assert.Equal(t, true, ans)
}

func TestUserService(t *testing.T) {
	userService := NewUserService(
		NewInMemoryUserStore(),
		NewJWTManager("cheryl", 10*time.Second),
	)

	admin := NewUser("admin", "admin", "admin")
	guest := NewUser("guest", "12345", "guest")
	userService.userStore.Save(admin)
	userService.userStore.Save(guest)

	adminToken, err := userService.Login("admin", "admin")
	assert.NoError(t, err)
	guestToken, err := userService.Login("guest", "12345")
	assert.NoError(t, err)
	err = userService.VerifyToken(guestToken)
	assert.Error(t, err)
	err = userService.VerifyToken(adminToken)
	assert.NoError(t, err)
	time.Sleep(11 * time.Second)
	err = userService.VerifyToken(adminToken)
	assert.Error(t, err)
}

type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.StandardClaims
}

func TestGenerate(t *testing.T) {
	mySigningKey := []byte("AllYourBase")

	// Create the Claims
	claims := MyCustomClaims{
		"bar",
		jwt.StandardClaims{
			ExpiresAt: 15000,
			Issuer:    "test",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	fmt.Printf("%v %v", ss, err)
}
