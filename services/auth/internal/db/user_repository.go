package db

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/entity"
)

type UserRepositoryInterface interface {
	Create(user entity.User) (int64, error)
	GetByEmail(email string) (*entity.User, error)
}

type UserRepository struct {
	DB *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}
func (ur *UserRepository) Create(user entity.User) (int64, error) {

	var id int64

	err := ur.DB.QueryRow("INSERT INTO users ( name,email, password) VALUES ($1,$2,$3) RETURNING id", user.Name, user.Email, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func (ur *UserRepository) GetByEmail(email string) (*entity.User, error) {

	var user entity.User
	err := ur.DB.Get(&user, "select id,name,password,created_at from users where email = $1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
	}
	return &user, nil

}
