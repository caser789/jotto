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

	/*
	 * KEYS[1] = backlog
	 * KEYS[2] = pending
	 * ARGV[1] = uuid
	 * ARGV[2] = job
	 */
	lua := `
		redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
		return redis.call("lpush", KEYS[2], ARGV[1])
	`

	_, err = rd.client.Eval(lua, []string{rd.key(queue, "backlog"), rd.key(queue, "pending")}, job.TraceID, job.Serialize()).Result()
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
	lua := `
		redis.call("hset", KEYS[1], KEYS[3], ARGV[1])
		return redis.call("zadd", KEYS[2], ARGV[2], KEYS[3])
	`

	keys := []string{rd.key(queue, "backlog"), rd.key(queue, "delayed"), job.TraceID}
	argv := []interface{}{job.Serialize(), at.Unix()}

	fmt.Println(keys)
	fmt.Println(argv)

	_, err = rd.client.Eval(lua, keys, argv...).Result()
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
	lua := `
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('lpush', KEYS[2], ARGV[1])
	`

	_, err = rd.client.Eval(lua, []string{rd.key(queue, "working"), rd.key(queue, "pending")}, job.TraceID).Result()

	return
}

// Complete removes the job from the queue entirely.
func (rd *RedisDriver) Complete(queue string, job *Job) (err error) {
	/*
	 * KEYS[1] = working
	 * KEYS[2] = backlog
	 * ARGV[1] = uuid
	 */
	lua := `
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('hdel', KEYS[2], ARGV[1])
	`

	_, err = rd.client.Eval(lua, []string{rd.key(queue, "working"), rd.key(queue, "backlog")}, job.TraceID).Result()

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
	lua := `
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('zadd', KEYS[2], ARGV[2], ARGV[1])
	`

	_, err = rd.client.Eval(lua, []string{rd.key(queue, "working"), rd.key(queue, "delayed")}, job.TraceID, time.Now().Add(after).Unix()).Result()
	return
}

// Fail moves the job into a failure list for trouble shooting.
func (rd *RedisDriver) Fail(queue string, job *Job) (err error) {
	/*
	 * KEYS[1] = working
	 * KEYS[2] = failure
	 * ARGV[1] = uuid
	 */
	lua := `
		redis.call('lrem', KEYS[1], 0, ARGV[1])
		return redis.call('lpush', KEYS[2], ARGV[1])
	`

	_, err = rd.client.Eval(lua, []string{rd.key(queue, "working"), rd.key(queue, "failure")}, job.TraceID).Result()
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

	/*
	 * KEYS[1] = pending
	 * KEYS[2] = working
	 * KEYS[3] = failure
	 * KEYS[4] = backlog
	 * KEYS[5] = delayed
	 * ARGV[1] = now
	 */
	lua := `
		local pending = redis.call('llen', KEYS[1])
		local working = redis.call('llen', KEYS[2])
		local failure = redis.call('llen', KEYS[3])
		local backlog = redis.call('hlen', KEYS[4])
		local delayed = redis.call('zcount', KEYS[5], '-inf', '+inf')
		local waiting = redis.call('zcount', KEYS[5], '-inf', ARGV[1])

		return {pending, working, failure, delayed, backlog, waiting}
	`

	keys := []string{
		rd.key(queue, "pending"),
		rd.key(queue, "working"),
		rd.key(queue, "failure"),
		rd.key(queue, "backlog"),
		rd.key(queue, "delayed"),
	}

	result, err := rd.client.Eval(lua, keys, time.Now().Unix()).Result()

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
