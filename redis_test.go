package main_test

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	mockRedisConn := redigomock.NewConn()
	// ...
	mockRedisPool := redis.NewPool(func() (redis.Conn, error) {
		return mockRedisConn, nil
	}, 10)

	assert.True(t, mockRedisPool.Get().Err() == nil)
}
