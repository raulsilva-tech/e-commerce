package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/raulsilva-tech/e-commerce/services/product/internal/entity"
	"github.com/raulsilva-tech/e-commerce/services/product/internal/repository"
)

type ProductUseCase struct {
	repo  *repository.ProductRepository
	cache *repository.ProductCache
}

func NewProductUseCase(r *repository.ProductRepository, c *repository.ProductCache) *ProductUseCase {
	return &ProductUseCase{repo: r, cache: c}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, name string, price float64) (int64, error) {

	p, err := entity.NewProduct(0, name, price)
	if err != nil {
		return 0, err
	}
	id, err := uc.repo.Create(ctx, p)

	return id, err
}

func (uc *ProductUseCase) ListProducts(ctx context.Context, limit int) ([]entity.Product, error) {

	return uc.repo.GetList(ctx, limit)

}

func (uc *ProductUseCase) GetByProductId(ctx context.Context, id int64) (*entity.Product, error) {

	var prod *entity.Product
	if err := uc.cache.GetProduct(strconv.Itoa(int(id)), prod); err == nil {
		fmt.Println("cache found!")
		return prod, nil
	} else {
		fmt.Println("cache NOT found!")
	}

	prod, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if prod != nil {
		err = uc.cache.SetProduct(strconv.Itoa(int(id)), prod)
		if err != nil {
			log.Println(err, err.Error())
		}
	}

	return prod, nil
}
