package usecase

import (
	"time"

	"github.com/raulsilva-tech/e-commerce/services/auth/internal/repository"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Email    string
	Password string
}

type SignupInput struct {
	Name     string
	Email    string
	Password string
}

type AuthUseCase struct {
	UserRepository *db.UserRepository
}

func NewAuthUseCase(repository *db.UserRepository) *AuthUseCase {
	return &AuthUseCase{
		UserRepository: repository,
	}
}
func (uc *AuthUseCase) Login(input LoginInput) (*entity.User, error) {

	if input.Email == "" || input.Password == "" {
		return nil, entity.ErrEmailPasswordRequired
	}
	user, err := uc.UserRepository.GetByEmail(input.Email)
	if err != nil || user == nil {
		return nil, entity.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, entity.ErrInvalidCredentials
	}

	return user, nil

}

func (uc *AuthUseCase) Signup(input SignupInput) (int64, error) {

	if input.Name == "" || input.Email == "" || input.Password == "" {
		return 0, entity.ErrNameEmailPasswordRequired
	}

	existingUser, err := uc.UserRepository.GetByEmail(input.Email)
	if err != nil {
		return 0, err
	}
	if existingUser != nil {
		return 0, entity.ErrEmailAlreadyUsed
	}

	user, err := entity.NewUser(0, input.Name, input.Email, input.Password, time.Now())
	if err != nil {
		return 0, err
	}

	id, err := uc.UserRepository.Create(*user)
	if err != nil {
		return 0, err
	}

	return id, nil
}
