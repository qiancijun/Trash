package jwtdemo

import (
	"fmt"
	"sync"
)

type UserStore interface {
	Save(*User) error
	Find(string) (*User, error)
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

var _ UserStore = (*InMemoryUserStore)(nil) // 检查 InMemoryUserStore 是否实现了 UserStore 接口

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}


func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.users[user.Username] != nil {
		return fmt.Errorf("用户已经存在")
	}
	
	store.users[user.Username] = user

	return nil
}

func (store *InMemoryUserStore) Find(username string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.users[username]
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return user, nil
} 