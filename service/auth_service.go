package service

import (
	"errors"
	"sync"
	"weblog-app/model"
	"weblog-app/repo"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user does not exist, please signup")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUsernameTaken    = errors.New("username is already taken")
	ErrEmptyCredentials = errors.New("username and password cannot be empty")
)

type AuthService interface {
	Signup(username, password string) (*model.User, string, error)
	Login(username, password string) (*model.User, string, error)
	ValidateSession(token string) (int, error)
	Logout(token string)
}

type authService struct {
	userRepo repo.UserRepository
	sessions map[string]int
	mu       sync.RWMutex
}

func NewAuthService(ur repo.UserRepository) AuthService {
	return &authService{
		userRepo: ur,
		sessions: make(map[string]int),
	}
}

func (s *authService) Signup(username, password string) (*model.User, string, error) {
	if username == "" || password == "" {
		return nil, "", ErrEmptyCredentials
	}

	existing, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	token := uuid.New().String()
	s.mu.Lock()
	s.sessions[token] = user.ID
	s.mu.Unlock()

	return user, token, nil
}

func (s *authService) Login(username, password string) (*model.User, string, error) {
	if username == "" || password == "" {
		return nil, "", ErrEmptyCredentials
	}

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	token := uuid.New().String()
	s.mu.Lock()
	s.sessions[token] = user.ID
	s.mu.Unlock()

	return user, token, nil
}

func (s *authService) ValidateSession(token string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, exists := s.sessions[token]
	if !exists {
		return 0, errors.New("invalid session")
	}
	return userID, nil
}

func (s *authService) Logout(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}