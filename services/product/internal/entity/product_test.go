package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {

	p, err := NewProduct(1, "Product", 2.1)

	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, p.Name, "Product")
	assert.Equal(t, p.Price, 2.1)
}

func TestNewProductWhenNameIsRequired(t *testing.T) {

	p, err := NewProduct(1, "", 2.1)

	assert.NotNil(t, err)
	assert.Nil(t, p)
	assert.Equal(t, err, ErrNameIsRequired)
}
