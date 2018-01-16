package main

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

func newPool(s Specification) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(s.RedisUrl) },
	}
}

var (
	pool *redis.Pool
)

func redisSave(key string, value Mobilization) {
	conn := pool.Get()
	defer conn.Close()

	if _, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...); err != nil {
		log.Printf("[worker] cache can't update local db %s: %s", key, err)
		return
	}
	log.Printf("[worker] cache updated at %s, reading from www.%s.bonde.org, to be served in %s ", value.CachedAt, value.Slug, value.CustomDomain)
}

func redisRead(key string) Mobilization {
	conn := pool.Get()
	defer conn.Close()
	var value Mobilization

	reply, err := redis.Values(conn.Do("HMGET", key))
	if err != nil {
		log.Printf("[worker] can't found key %s into cache: %s", key, err)
		return value
	}

	if _, err := redis.Scan(reply, &value); err != nil {
		log.Printf("[worker] can't found key %s into cache: %s", key, err)
	}
	log.Printf("[worker] cache updated at %s, reading from www.%s.bonde.org, to be served in %s ", value.CachedAt, value.Slug, value.CustomDomain)
	return value
}
