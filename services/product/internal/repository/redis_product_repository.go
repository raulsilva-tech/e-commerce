package repository

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type ProductCache struct {
	client *redis.Client
}

func NewProductCache(addr string) *ProductCache {

	db := redis.NewClient(&redis.Options{Addr: addr})

	return &ProductCache{client: db}
}

func (c *ProductCache) SetProduct(key string, value interface{}) error {

	data, _ := json.Marshal(value)

	return c.client.Set(key, data, 10*time.Minute).Err()
}

func (c *ProductCache) GetProduct(key string, dest interface{}) error {

	data, err := c.client.Get(key).Bytes()

	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)

}
