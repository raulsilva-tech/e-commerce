package entity

import "errors"

var (
	ErrProductIdIsRequired = errors.New("product id is required")
	ErrQuantityIsRequired  = errors.New("quantity is required")
)

type Order struct {
	ID        int64
	ProductID int64
	Quantity  int
	Total     float64
}

func NewOrder(id int64, productID int64, qt int, total float64) (*Order, error) {

	o := &Order{id, productID, qt, total}

	err := o.Validate()
	if err != nil {
		return nil, err
	}

	return o, nil

}

func (o *Order) Validate() error {

	if o.ProductID == 0 {
		return ErrProductIdIsRequired
	}
	if o.Quantity == 0 {
		return ErrQuantityIsRequired
	}

	return nil
}
