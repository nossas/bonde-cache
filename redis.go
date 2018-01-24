package main

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisPool are responsible to ensure we have at least three redis active connections
func RedisPool(s Specification) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(s.RedisURL) },
	}
}

var (
	pool *redis.Pool
)

// RedisSaveMobilization is used to save Mobilization cached content
func RedisSaveMobilization(key string, value Mobilization) {
	conn := pool.Get()
	defer conn.Close()

	if value.Name == "" {
		log.Printf("Empty mobilization.")
	}

	if _, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...); err != nil {
		log.Printf("[worker] cache can't update local db %s: %s", key, err)
		return
	}
	log.Printf("[worker] cache updated at %s, reading from www.%s.bonde.org, to be served in %s ", value.CachedAt, value.Slug, value.CustomDomain)
}

// RedisReadMobilization Load Mobilization From Redis based on key
func RedisReadMobilization(key string) Mobilization {
	conn := pool.Get()
	defer conn.Close()
	var value Mobilization

	reply, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		log.Printf("[worker] can't found key %s into cache: %s", key, err)
		value = Mobilization{Name: ""}
	} else {
		if err2 := redis.ScanStruct(reply, &value); err2 != nil {
			log.Printf("[worker] can't found key %s into cache: %s", key, err2)
		}
	}

	return value
}
