package jotto

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type CacheDriver interface {
	Get(key string) (value string, err error)
	Set(key, value string, expiration time.Duration) error
	Has(key string) (bool, error)
}

// RedisDriver implements both the CacheDriver and QueueDriver interface
type RedisDriver struct {
	name     string
	settings *RedisSettings
	client   *redis.Client
}

func NewRedisDriver(name string, settings *RedisSettings) *RedisDriver {
	client := redis.NewClient(&redis.Options{
		Addr:         settings.Address,
		Password:     settings.Password,
		DB:           settings.Database,
		DialTimeout:  time.Second * time.Duration(settings.DialTimeout),
		ReadTimeout:  time.Second * time.Duration(settings.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(settings.WriteTimeout),
	})

	return &RedisDriver{
		name:     name,
		settings: settings,
		client:   client,
	}
}

/* CacheDriver */

func (rd *RedisDriver) Get(key string) (value string, err error) {
	return rd.client.Get(key).Result()
}

func (rd *RedisDriver) Set(key string, value string, expiration time.Duration) (err error) {
	_, err = rd.client.Set(key, value, expiration).Result()

	return
}

func (rd *RedisDriver) Has(key string) (bool, error) {
	err := rd.client.Get(key).Err()

	return err == nil, err
}

/* QueueDriver */

func (rd *RedisDriver) Push(queue string, job *Job) error {
	_, err := rd.client.LPush(queue, job.Serialize()).Result()

	return err
}

func (rd *RedisDriver) Pop(queue string) (job *Job, err error) {
	values, err := rd.client.BRPop(time.Second*2, queue).Result()

	if err != nil {
		return
	}

	job = &Job{}
	err = job.Unserialize(values[1])

	if err != nil {
		return nil, err
	}

	return
}

func NewNullDriver(name string) *NullDriver {
	return &NullDriver{name: name}
}

type NullDriver struct {
	name string
}

func (nd *NullDriver) Get(key string) (value string, err error) {
	return "", fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

func (nd *NullDriver) Set(key string, value string, expiration time.Duration) error {
	return fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

func (nd *NullDriver) Has(key string) (bool, error) {
	return false, fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}
