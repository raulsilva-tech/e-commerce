package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {

	u, err := NewUser(1, "Raul", "raul@gmail.com", "", time.Now())

	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, u.Name, "Raul")
	assert.Equal(t, u.Email, "raul@gmail.com")
}

func TestNewUserWhenNameIsRequired(t *testing.T) {

	u, err := NewUser(1, "", "raul@gmail.com", "", time.Now())

	assert.NotNil(t, err)
	assert.Nil(t, u)
	assert.Equal(t, err, ErrNameIsRequired)
}

func TestNewUserWhenEmailIsRequired(t *testing.T) {

	u, err := NewUser(1, "Raul", "", "", time.Now())

	assert.NotNil(t, err)
	assert.Nil(t, u)
	assert.Equal(t, err, ErrEmailIsRequired)
}
