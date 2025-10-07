package repository

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/raulsilva-tech/e-commerce/services/product/internal/entity"
	"github.com/stretchr/testify/suite"
)

func migrateDB() (*sqlx.DB, error) {

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE products (
    id integer PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price REAL
);`)

	return db, err
}

type ProductRepositoryTestSuite struct {
	DB *sqlx.DB
	suite.Suite
}

func TestProductRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}

func (suite *ProductRepositoryTestSuite) TearDownSuite() {
	suite.DB.Close()
}

func (suite *ProductRepositoryTestSuite) SetupSuite() {
	dbConn, err := migrateDB()
	suite.NoError(err)
	suite.DB = dbConn
}

func (suite *ProductRepositoryTestSuite) TestCreate() {

	p, _ := entity.NewProduct(1, "Product 1", 2.1)

	repo := NewProductRepository(suite.DB)
	id, err := repo.Create(context.Background(), p)

	suite.Nil(err)
	suite.NotEmpty(id)
}

func (suite *ProductRepositoryTestSuite) TestGetById() {

	p, _ := entity.NewProduct(2, "Product 2", 2.1)

	repo := NewProductRepository(suite.DB)
	id, err := repo.Create(context.Background(), p)

	suite.Nil(err)
	suite.NotEmpty(id)

	p2, err := repo.GetByID(context.Background(), p.ID)
	suite.Nil(err)
	suite.NotNil(p2)
	suite.Equal(p.ID, p2.ID)

}
