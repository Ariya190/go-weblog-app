package repo

import (
	"database/sql"
	"weblog-app/model"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByUsername(username string) (*model.User, error)
	GetByID(id int) (*model.User, error)
	GetByNames(usernames []string) ([]model.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(user *model.User) error {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRow(query, user.Username, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepo) GetByUsername(username string) (*model.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`
	var u model.User
	err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByID(id int) (*model.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE id = $1`
	var u model.User
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByNames(usernames []string) ([]model.User, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var users []model.User
	for _, name := range usernames {
		if name == "" {
			continue
		}
		u, err := r.GetByUsername(name)
		if err == nil && u != nil {
			users = append(users, *u)
		}
	}
	return users, nil
}