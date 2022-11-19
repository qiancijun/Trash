package jwtdemo

import "fmt"

type UserService struct {
	userStore  UserStore
	jwtManager *JWTManager
}

func NewUserService(userStore UserStore, jwtManager *JWTManager) *UserService {
	return &UserService{
		userStore:  userStore,
		jwtManager: jwtManager,
	}
}

func (us *UserService) Login(username, password string) (string, error) {
	loginUser, err := us.userStore.Find(username)
	if err != nil {
		return "", err
	}
	if loginUser == nil || !loginUser.IsCorrectPassword(password) {
		return "", fmt.Errorf("用户不存在或密码错误")
	}
	// 密码正确
	token, err := us.jwtManager.Generate(loginUser)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (us *UserService) VerifyToken(token string) error {
	claims, err := us.jwtManager.Verify(token)
	if err != nil {
		return fmt.Errorf("token 不合法")
	}
	if claims.Role != "admin" {
		return fmt.Errorf("权限不足")
	}
	return nil
}