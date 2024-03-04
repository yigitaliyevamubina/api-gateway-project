package redis

import (
	rd "github.com/gomodule/redigo/redis"
	"myproject/api-gateway/storage/repo"
)

type redisRepo struct {
	reds *rd.Pool
}

func NewRedisRepo(rds *rd.Pool) repo.InMemoryStorageI {
	return &redisRepo{reds: rds}
}

func (r *redisRepo) Set(key, value string) (err error) {
	conn := r.reds.Get()
	defer conn.Close()

	_, err = conn.Do("SET", key, value)
	return err
}

func (r *redisRepo) SetWithTTL(key, value string, seconds int) (err error) {
	conn := r.reds.Get()
	defer conn.Close()

	_, err = conn.Do("SETEX", key, seconds, value)
	return err
}

func (r *redisRepo) Get(key string) (interface{}, error) {
	conn := r.reds.Get()
	defer conn.Close()

	return conn.Do("GET", key)
}
