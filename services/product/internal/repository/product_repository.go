package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/raulsilva-tech/e-commerce/services/product/internal/entity"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *entity.Product) (int64, error) {

	err := r.db.QueryRowContext(ctx, "INSERT INTO products (name, price) VALUES ($1, $2) RETURNING id", p.Name, p.Price).Scan(&p.ID)
	if err != nil {
		return 0, err
	}

	return p.ID, err
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name, price FROM products WHERE id = $1", id)

	var p entity.Product
	if err := row.Scan(&p.ID, &p.Name, &p.Price); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) GetList(ctx context.Context, limit int) ([]entity.Product, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, price FROM products limit $1", limit)
	if err != nil {
		return []entity.Product{}, err
	}

	var list []entity.Product

	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {

			return []entity.Product{}, err
		}

		list = append(list, p)
	}

	return list, nil
}
