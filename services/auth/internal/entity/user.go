package entity

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrIdIsRequired    = errors.New("id is required")
	ErrEmailIsRequired = errors.New("email is required")
	ErrNameIsRequired  = errors.New("name is required")
	ErrInvalidEmail    = errors.New("invalid email address")
)

type User struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func NewUser(id int64, name, email, password string, createdAt time.Time) (*User, error) {

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  string(hashed),
		CreatedAt: createdAt,
	}

	return u, u.Validate()
}

func (u *User) Validate() error {

	if u.Name == "" {
		return ErrNameIsRequired
	}

	if u.Email == "" {
		return ErrEmailIsRequired
	}

	if !strings.Contains(u.Email, "@") {
		return ErrInvalidEmail
	}

	return nil
}
