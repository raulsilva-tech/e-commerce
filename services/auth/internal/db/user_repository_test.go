package db

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/entity"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	DB *sqlx.DB
	suite.Suite
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	suite.DB.Close()
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	dbConn, err := migrateDB()
	suite.NoError(err)
	suite.DB = dbConn
}

func (suite *UserRepositoryTestSuite) TestCreate() {

	u, _ := entity.NewUser(1, "Raul", "raul@gmail.com", "", time.Now())

	repo := NewUserRepository(suite.DB)
	id, err := repo.Create(*u)

	suite.Nil(err)
	suite.NotEmpty(id)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail() {

	u, _ := entity.NewUser(1, "Raul", "raul@gmail.com", "", time.Now())

	repo := NewUserRepository(suite.DB)
	id, err := repo.Create(*u)

	suite.Nil(err)
	suite.NotEmpty(id)

	u2, err := repo.GetByEmail(u.Email)
	suite.Nil(err)
	suite.NotNil(u2)
	suite.Equal(u.ID, u2.ID)

}

func migrateDB() (*sqlx.DB, error) {

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE users (
    id integer PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at DATETIME
);`)

	return db, err
}
