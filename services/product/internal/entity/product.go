package entity

import "errors"

var (
	ErrNameIsRequired = errors.New("name is required")
)

type Product struct {
	ID    int64   `db:"id" json:"id"`
	Name  string  `db:"name" json:"name"`
	Price float64 `db:"price" json:"price"`
}

func NewProduct(id int64, name string, price float64) (*Product, error) {

	p := &Product{id, name, price}

	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Product) Validate() error {

	if p.Name == "" {
		return ErrNameIsRequired
	}

	return nil
}
