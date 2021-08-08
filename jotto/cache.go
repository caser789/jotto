package jotto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// CacheDriver - interface for Motto cache drivers
type CacheDriver interface {
	Get(key string) (value string, err error)
	// GetVia - get the `key` from cache, if not set, call `handler` to
	// get the value and put it into the cache afterwards.
	GetVia(key string, handler func() (value string, expiration time.Duration, err error)) (string, error)
	Set(key, value string, expiration time.Duration) error
	Has(key string) (bool, error)
	Del(keys ...string) (bool, error)
	Flush() (bool, error)
	// Incr - increment the given `key`
	Incr(key string) (int64, error)
	// Expire - set the expire time of `key`
	Expire(key string, expiry time.Duration) (bool, error)
	// Guard - guard the execution of the `handler` function with a lock
	// under the hood, check if the `key` is set in cache, if yes, `handler`
	// will not be executed; otherwise, set the key and execute `handler`.
	Guard(key string, expiration time.Duration, handler func() error) error
}

// RedisDriver implements both the CacheDriver and QueueDriver interface
type RedisDriver struct {
	name     string
	settings *RedisSettings
	client   *redis.Client
}

// NewRedisDriver - create a Redis driver
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

// Get - retrieve `key` from Redis
func (rd *RedisDriver) Get(key string) (value string, err error) {
	return rd.client.Get(key).Result()
}

// GetVia - retrieve `key` from Redis
func (rd *RedisDriver) GetVia(key string, handler func() (string, time.Duration, error)) (value string, err error) {
	value, err = rd.client.Get(key).Result()

	if err != nil && err != redis.Nil {
		return "", err
	}

	value, expiration, err := handler()

	if err != nil {
		return "", err
	}

	return value, rd.Set(key, value, expiration)
}

// Set - put `key` into Redis
func (rd *RedisDriver) Set(key string, value string, expiration time.Duration) (err error) {
	_, err = rd.client.Set(key, value, expiration).Result()

	return
}

// Has - check if `key` exists in Redis
func (rd *RedisDriver) Has(key string) (bool, error) {
	err := rd.client.Get(key).Err()

	return err == nil, err
}

// Del - delete `keys` from Redis
func (rd *RedisDriver) Del(keys ...string) (bool, error) {
	err := rd.client.Del(keys...).Err()

	return err == nil, err
}

// Flush - delete `keys` in the currently selected DB from Redis
func (rd *RedisDriver) Flush() (bool, error) {
	err := rd.client.FlushDB().Err()

	return err == nil, err
}

// Incr - increase the value of `key`
func (rd *RedisDriver) Incr(key string) (int64, error) {
	return rd.client.Incr(key).Result()
}

// Expire - set the expire time of `key`
func (rd *RedisDriver) Expire(key string, expiry time.Duration) (bool, error) {
	return rd.client.Expire(key, expiry).Result()
}

// Guard - guard the execution of `handler` with a lock
func (rd *RedisDriver) Guard(key string, expiration time.Duration, handler func() error) (err error) {
	acquired, err := rd.client.SetNX(key, "guarded", expiration).Result()

	if err != nil {
		return
	}

	if !acquired {
		return fmt.Errorf("acquiring lock failed: %s", key)
	}

	defer func() {
		rd.client.Del(key)
	}()

	return handler()
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

	/*
	 * KEYS[1] = backlog
	 * KEYS[2] = pending
	 * ARGV[1] = uuid
	 * ARGV[2] = job
	 */
	script := redis.NewScript(`
		redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
		return redis.call("lpush", KEYS[2], ARGV[1])
	`)

	_, err = script.Run(rd.client, []string{rd.key(queue, "backlog"), rd.key(queue, "pending")}, job.TraceID, job.Serialize()).Result()
	return
}

// Schedule pushes a new job into the delayed queue so that it will be processed at a later time.
func (rd *RedisDriver) Schedule(queue string, job *Job, at time.Time) (err error) {
	if job.TraceID == "" {
		job.TraceID = GenerateTraceID()
	}

	/*
	 * KEYS[1] = backlog
	 * KEYS[2] = delayed
	 * KEYS[3] = uuid
	 * ARGV[1] = job
	 * ARGV[2] = score
	 */
	script := redis.NewScript(`
		redis.call("hset", KEYS[1], KEYS[3], ARGV[1])
		return redis.call("zadd", KEYS[2], ARGV[2], KEYS[3])
	`)

	keys := []string{rd.key(queue, "backlog"), rd.key(queue, "delayed"), job.TraceID}
	argv := []interface{}{job.Serialize(), at.Unix()}

	fmt.Println(keys)
	fmt.Println(argv)

	_, err = script.Run(rd.client, keys, argv...).Result()
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
	/*
	 * KEYS[1] = working
	 * KEYS[2] = pending
	 * ARGV[1] = uuid
	 */
	script := redis.NewScript(`
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('lpush', KEYS[2], ARGV[1])
	`)

	_, err = script.Run(rd.client, []string{rd.key(queue, "working"), rd.key(queue, "pending")}, job.TraceID).Result()

	return
}

// Complete removes the job from the queue entirely.
func (rd *RedisDriver) Complete(queue string, job *Job) (err error) {
	/*
	 * KEYS[1] = working
	 * KEYS[2] = backlog
	 * ARGV[1] = uuid
	 */
	script := redis.NewScript(`
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('hdel', KEYS[2], ARGV[1])
	`)

	_, err = script.Run(rd.client, []string{rd.key(queue, "working"), rd.key(queue, "backlog")}, job.TraceID).Result()

	return
}

// Defer moves the job to a deferred queue for processing at a later time.
func (rd *RedisDriver) Defer(queue string, job *Job, after time.Duration) (err error) {
	/*
	 * KEYS[1] = working
	 * KEYS[2] = delayed
	 * ARGV[1] = uuid
	 * ARGV[2] = score
	 */
	script := redis.NewScript(`
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('zadd', KEYS[2], ARGV[2], ARGV[1])
	`)

	_, err = script.Run(rd.client, []string{rd.key(queue, "working"), rd.key(queue, "delayed")}, job.TraceID, time.Now().Add(after).Unix()).Result()
	return
}

// Fail moves the job into a failure list for trouble shooting.
func (rd *RedisDriver) Fail(queue string, job *Job) (err error) {
	/*
	 * KEYS[1] = working
	 * KEYS[2] = failure
	 * ARGV[1] = uuid
	 */
	script := redis.NewScript(`
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('lpush', KEYS[2], ARGV[1])
	`)

	_, err = script.Run(rd.client, []string{rd.key(queue, "working"), rd.key(queue, "failure")}, job.TraceID).Result()
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

// Stats - get the stats of the queue
func (rd *RedisDriver) Stats(queue string) (stats *QueueStats, err error) {
	stats = &QueueStats{}

	if err != nil {
		return nil, err
	}

	/*
	 * KEYS[1] = pending
	 * KEYS[2] = working
	 * KEYS[3] = failure
	 * KEYS[4] = backlog
	 * KEYS[5] = delayed
	 * ARGV[1] = now
	 */
	script := redis.NewScript(`
		local pending = redis.call('llen', KEYS[1])
		local working = redis.call('llen', KEYS[2])
		local failure = redis.call('llen', KEYS[3])
		local backlog = redis.call('hlen', KEYS[4])
		local delayed = redis.call('zcount', KEYS[5], '-inf', '+inf')
		local waiting = redis.call('zcount', KEYS[5], '-inf', ARGV[1])

		return {pending, working, failure, delayed, backlog, waiting}
	`)

	keys := []string{
		rd.key(queue, "pending"),
		rd.key(queue, "working"),
		rd.key(queue, "failure"),
		rd.key(queue, "backlog"),
		rd.key(queue, "delayed"),
	}

	result, err := script.Run(rd.client, keys, time.Now().Unix()).Result()

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

	/*
	 * KEYS[1] = delayed
	 * KEYS[2] = pending
	 * ARGV[1] = now
	 */
	lua := `
		local ready = redis.call('zrangebyscore', KEYS[1], '-inf', ARGV[1])
		local count = 0

		for k,v in pairs(ready) do
			redis.call('zrem', KEYS[1], v)
			redis.call('lpush', KEYS[2], v)
			count = count + 1
		end

		return count
	`

	keys := []string{rd.key(queue, "delayed"), rd.key(queue, "pending")}
	argv := []interface{}{time.Now().Unix()}

	result, err := rd.client.Eval(lua, keys, argv...).Result()
	count = result.(int64)

	return
}

func (rd *RedisDriver) key(queue string, segment string) string {
	return fmt.Sprintf("%s:%s", queue, segment)
}

/* Null cache driver */

// NewNullDriver - creates a null cache driver
func NewNullDriver(name string) *NullDriver {
	return &NullDriver{name: name}
}

// NullDriver - a null cache driver
type NullDriver struct {
	name string
}

// Get - get key
func (nd *NullDriver) Get(key string) (value string, err error) {
	return "", fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Set - set key
func (nd *NullDriver) Set(key string, value string, expiration time.Duration) error {
	return fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Has - check existence
func (nd *NullDriver) Has(key string) (bool, error) {
	return false, fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Del - delete key
func (nd *NullDriver) Del(keys ...string) (bool, error) {
	return false, fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Flush - flush db
func (nd *NullDriver) Flush() (bool, error) {
	return false, fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// GetVia - get from cache otherwise set via `handler`
func (nd *NullDriver) GetVia(key string, handler func() (value string, expiration time.Duration, err error)) (string, error) {
	return "", fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Expire - set the expire time of `key`
func (nd *NullDriver) Expire(key string, expiry time.Duration) (bool, error) {
	return false, fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Incr - increase by 1
func (nd *NullDriver) Incr(key string) (int64, error) {
	return 0, fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}

// Guard - guard execution of `hander` with a lock
func (nd *NullDriver) Guard(key string, expiration time.Duration, handler func() error) error {
	return fmt.Errorf("Cannot find settings of cache named `%s`", nd.name)
}
