package jotto

import (
	"encoding/json"
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

// queue:pending (list, uuid)
// queue:working (list, uuid)
// queue:failure (list, uuid)

// queue:delayed (sorted set, uuid by timestamp)
// queue:backlog (hash, uuid => job)

// Enqueue pushes a new job into the queue
func (rd *RedisDriver) Enqueue(queue string, job *Job) (err error) {
	if job.TraceID == "" {
		job.TraceID = GenerateTraceID()
	}

	lua := `
		local queue = KEYS[1]
		local uuid = KEYS[2]
		local job = ARGV[1]

		redis.call("hset", queue..":backlog", uuid, job)
		return redis.call("lpush", queue..":pending", uuid)
	`

	_, err = rd.client.Eval(lua, []string{queue, job.TraceID}, job.Serialize()).Result()
	return
}

// Schedule pushes a new job into the delayed queue so that it will be processed at a later time.
func (rd *RedisDriver) Schedule(queue string, job *Job, at time.Time) (err error) {
	if job.TraceID == "" {
		job.TraceID = GenerateTraceID()
	}

	lua := `
		local queue = KEYS[1]
		local uuid  = KEYS[2]
		local job   = ARGV[1]
		local score = ARGV[2]

		redis.call("hset", queue..":backlog", uuid, job)
		return redis.call("zadd", queue..":delayed", score, uuid)
	`

	_, err = rd.client.Eval(lua, []string{queue, job.TraceID}, job.Serialize(), at.Unix()).Result()
	return
}

// Dequeue retrieves a job from the queue
func (rd *RedisDriver) Dequeue(queue string) (job *Job, err error) {
	/*
	 * 1. BRPOPLPUSH queue:pending queue:working
	 * 2. Fetch jobs from queue:backlog
	 */
	jobID, err := rd.client.BRPopLPush(rd.key(queue, "pending"), rd.key(queue, "working"), time.Duration(rd.settings.ReadTimeout)*time.Second).Result()

	if err != nil {
		return
	}

	serialized, err := rd.client.HGet(rd.key(queue, "backlog"), jobID).Result()

	if err != nil {
		return
	}

	job = &Job{}

	err = json.Unmarshal([]byte(serialized), job)

	return
}

// Attempt requeues the job
func (rd *RedisDriver) Attempt(queue string, job *Job) (err error) {
	job.Attempt()

	_, err = rd.client.HSet(rd.key(queue, "backlog"), job.TraceID, job.Serialize()).Result()

	return
}

// Requeue requeues the job
func (rd *RedisDriver) Requeue(queue string, job *Job) (err error) {
	lua := `
		local queue = KEYS[1]
		local uuid = KEYS[2]
		local job = ARGV[1]

		redis.call('lrem', queue..':working', 0, uuid)
		return redis.call('lpush', queue..':pending', uuid)
	`

	_, err = rd.client.Eval(lua, []string{queue, job.TraceID}, job.Serialize()).Result()

	return
}

// Complete removes the job from the queue entirely.
func (rd *RedisDriver) Complete(queue string, job *Job) (err error) {
	lua := `
		local queue = KEYS[1]
		local uuid  = KEYS[2]
		local job   = ARGV[1]

		redis.call('lrem', queue..':working', 0, uuid)
		return redis.call('hdel', queue..':backlog', uuid)
	`

	_, err = rd.client.Eval(lua, []string{queue, job.TraceID}, job.Serialize()).Result()

	return
}

// Defer moves the job to a deferred queue for processing at a later time.
func (rd *RedisDriver) Defer(queue string, job *Job, after time.Duration) (err error) {
	lua := `
		local queue = KEYS[1]
		local uuid  = KEYS[2]
		local score = ARGV[1]

		redis.call('lrem', queue..':working', 0, uuid)
		return redis.call('zadd', queue..':delayed', score, uuid)
	`

	_, err = rd.client.Eval(lua, []string{queue, job.TraceID}, time.Now().Add(after).Unix()).Result()
	return
}

// Fail moves the job into a failure list for trouble shooting.
func (rd *RedisDriver) Fail(queue string, job *Job) (err error) {
	lua := `
		local queue = KEYS[1]
		local uuid  = ARGV[1]

		redis.call('lrem', queue..':working', 0, uuid)
		return redis.call('lpush', queue..':failure', uuid)
	`

	_, err = rd.client.Eval(lua, []string{queue}, job.TraceID).Result()
	return
}

// Truncate discards everything (!!DANGER!!) currently stored in the queue
func (rd *RedisDriver) Truncate(queue string) (err error) {
	deleted, err := rd.client.Del(
		rd.key(queue, "pending"),
		rd.key(queue, "working"),
		rd.key(queue, "failure"),
		rd.key(queue, "backlog"),
		rd.key(queue, "delayed"),
	).Result()

	fmt.Println("deleted", deleted, queue, err)
	return
}

func (rd *RedisDriver) Stats(queue string) (stats *QueueStats, err error) {
	stats = &QueueStats{}

	if err != nil {
		return nil, err
	}

	lua := `
		local queue = KEYS[1]
		local now = ARGV[1]

		local pending = redis.call('llen', queue..':pending')
		local working = redis.call('llen', queue..':working')
		local failure = redis.call('llen', queue..':failure')
		local delayed = redis.call('zcount', queue..':delayed', '-inf', '+inf')
		local backlog = redis.call('hlen', queue..':backlog')
		local waiting = redis.call('zcount', queue..':delayed', '-inf', now)

		return {pending, working, failure, delayed, backlog, waiting}
	`

	result, err := rd.client.Eval(lua, []string{queue}, time.Now().Unix()).Result()

	if err != nil {
		return
	}

	counts := result.([]interface{})

	return &QueueStats{
		Pending: counts[0].(int64),
		Working: counts[1].(int64),
		Failure: counts[2].(int64),
		Delayed: counts[3].(int64),
		Backlog: counts[4].(int64),
		Waiting: counts[5].(int64),
	}, nil
}

// ScheduleDeferred moves deferred jobs that are ready for processing to the pending queue
func (rd *RedisDriver) ScheduleDeferred(queue string) (count int64, err error) {

	lua := `
		local queue = KEYS[1]
		local now   = ARGV[1]

		local ready = redis.call('zrangebyscore', queue..':delayed', '-inf', now)
		local count = 0

		for k,v in pairs(ready) do
			redis.call('zrem', queue..':delayed', v)
			redis.call('lpush', queue..':pending', v)
			count = count + 1
		end

		return count
	`

	result, err := rd.client.Eval(lua, []string{queue}, time.Now().Unix()).Result()
	count = result.(int64)

	return
}

func (rd *RedisDriver) key(queue string, segment string) string {
	return fmt.Sprintf("%s:%s", queue, segment)
}

/* Null cache driver */

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
