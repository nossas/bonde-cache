package main

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Redis main struct
type Redis struct {
	s    Specification
	pool *redis.Pool
}

// CreatePool are responsible to ensure we have at least three redis active connections
func (r *Redis) CreatePool() *Redis {
	r.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(r.s.RedisURL)
		},
	}
	return r
}

// SaveMobilization is used to save Mobilization cached content
func (r *Redis) SaveMobilization(key string, value Mobilization) bool {
	conn := r.pool.Get()
	defer conn.Close()

	if value.Name == "" {
		log.Printf("Empty mobilization.")
		return false
	}

	if _, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...); err != nil {
		log.Printf("[populate_cache] cache can't update local db %s: %s", key, err)
		return false
	}

	log.Printf("[populate_cache] cache updated at %s, reading from www.%s.bonde.org, to be served in %s ", value.CachedAt, value.Slug, value.CustomDomain)
	return true
}

// ReadMobilization Load Mobilization From Redis based on key
func (r *Redis) ReadMobilization(key string) Mobilization {
	conn := r.pool.Get()
	defer conn.Close()
	var value Mobilization

	reply, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		log.Printf("[populate_cache] can't found key %s into cache: %s", key, err)
		value = Mobilization{Name: ""}
	} else {
		if err2 := redis.ScanStruct(reply, &value); err2 != nil {
			log.Printf("[populate_cache] can't found key %s into cache: %s", key, err2)
			value = Mobilization{Name: ""}
		}
	}

	return value
}
