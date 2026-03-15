package json

import (
	"encoding/json"
	"log"
	"os"
	"proxy/internal/config"
	"proxy/internal/storage/users"
	"sync"
	"time"
)

type UserStorage struct {
	users          []users.User
	userFilePath   string
	loadedTime     time.Time
	reloadDuration time.Duration
	mu             sync.RWMutex
}

func (s *UserStorage) Add(user users.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users = append(s.users, user)
	s.save()

	return nil
}

func (s *UserStorage) FindByUsername(username string) (*users.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if time.Since(s.loadedTime) > s.reloadDuration {
		s.load()
	}

	for _, user := range s.users {
		if user.Username == username {
			return &user, nil
		}
	}

	return &users.User{}, nil
}

func (s *UserStorage) GetAll() ([]users.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if time.Since(s.loadedTime) > s.reloadDuration {
		s.load()
	}

	return s.users, nil
}

func (s *UserStorage) load() {
	data, err := os.ReadFile(s.userFilePath)
	if err != nil {
		log.Panicf("Error reading users file: %v", err)
	}

	if err := json.Unmarshal(data, &s.users); err != nil {
		log.Panic("Error parsing users file:", err)
	}
}

func (s *UserStorage) save() {
	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		log.Panicf("Error marshaling users: %v", err)
	}

	if err := os.WriteFile(s.userFilePath, data, 0644); err != nil {
		log.Panicf("Error writing users file: %v", err)
	}
}

func NewStorage(appConfig *config.AppConfig) users.UserStorageInterface {
	p := &UserStorage{
		userFilePath:   appConfig.UserFile,
		loadedTime:     time.Now(),
		reloadDuration: time.Duration(appConfig.UserReloadDuration) * time.Minute,
	}
	p.load()

	return p
}
