package Redis

import (
	"app/config"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Cache struct {
	p        *redis.Pool // redis connection pool
	conninfo string
	dbNum    int
	key      string
	password string
	maxIdle  int

	//the timeout to a value less than the redis server's timeout.
	timeout time.Duration
}

func (rc *Cache) Load() error {
	conf := config.GetConfAll()
	rc.conninfo = conf.RedisHost + ":" + conf.RedisPort
	rc.dbNum, _ = strconv.Atoi(conf.RedisDb)
	rc.password = conf.RedisPassWord
	rc.maxIdle = 20
	rc.key = "app"
	rc.timeout = 180 * time.Second
	rc.ConnectInit()
	c := rc.p.Get()
	defer c.Close()
	return c.Err()
}

func (rc *Cache) ConnectInit() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rc.conninfo)
		if err != nil {
			return nil, err
		}
		if rc.password != "" {
			if _, err := c.Do("AUTH", rc.password); err != nil {
				c.Close()
				return nil, err
			}
		}
		_, selecterr := c.Do("SELECT", rc.dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}
	// initialize a new pool
	rc.p = &redis.Pool{
		MaxIdle:     rc.maxIdle,
		IdleTimeout: rc.timeout,
		Dial:        dialFunc,
	}
}

// actually do the redis cmds, args[0] must be the key name.
func (rc *Cache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("missing required arguments")
	}
	args[0] = rc.associate(args[0])
	c := rc.p.Get()
	defer c.Close()
	return c.Do(commandName, args...)
}

// Get cache from redis.
func (rc *Cache) Get(key string) string {
	if v, err := rc.do("GET", key); err == nil {
		if v == nil {
			return ""
		} else {
			return string(v.([]byte))
		}
	}
	return ""
}

// Get cache ttl from redis.
func (rc *Cache) Ttl(key string) int64 {
	if v, err := rc.do("TTL", key); err == nil {
		if v == nil {
			return -1
		} else {
			return v.(int64)
		}
	}
	return -1
}

// Set cache ttl from redis.
func (rc *Cache) Expire(key string, timeout time.Duration) error {
	_, err := rc.do("EXPIRE", key, int64(timeout/time.Second))
	return err
}

// GetMulti get cache from redis.
func (rc *Cache) GetMulti(keys []string) []interface{} {
	c := rc.p.Get()
	defer c.Close()
	var args []interface{}
	for _, key := range keys {
		args = append(args, rc.associate(key))
	}
	values, err := redis.Values(c.Do("MGET", args...))
	if err != nil {
		return nil
	}
	return values
}

// Put put cache to redis.
func (rc *Cache) Put(key string, val interface{}, timeout time.Duration) error {
	_, err := rc.do("SETEX", key, int64(timeout/time.Second), val)
	return err
}

// Delete delete cache in redis.
func (rc *Cache) Delete(key string) error {
	_, err := rc.do("DEL", key)
	return err
}

// IsExist check cache's existence in redis.
func (rc *Cache) IsExist(key string) bool {
	v, err := redis.Bool(rc.do("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

// Incr increase counter in redis.
func (rc *Cache) Incr(key string) error {
	_, err := redis.Bool(rc.do("INCRBY", key, 1))
	return err
}

// Decr decrease counter in redis.
func (rc *Cache) Decr(key string) error {
	_, err := redis.Bool(rc.do("INCRBY", key, -1))
	return err
}

// ClearAll clean all cache in redis. delete this redis collection.
func (rc *Cache) ClearAll() error {
	cachedKeys, err := rc.Scan(rc.key + ":*")
	if err != nil {
		return err
	}
	c := rc.p.Get()
	defer c.Close()
	for _, str := range cachedKeys {
		if _, err = c.Do("DEL", str); err != nil {
			return err
		}
	}
	return err
}

// Scan scan all keys matching the pattern. a better choice than `keys`
func (rc *Cache) Scan(pattern string) (keys []string, err error) {
	c := rc.p.Get()
	defer c.Close()
	var (
		cursor uint64 = 0 // start
		result []interface{}
		list   []string
	)
	for {
		result, err = redis.Values(c.Do("SCAN", cursor, "MATCH", pattern, "COUNT", 1024))
		if err != nil {
			return
		}
		list, err = redis.Strings(result[1], nil)
		if err != nil {
			return
		}
		keys = append(keys, list...)
		cursor, err = redis.Uint64(result[0], nil)
		if err != nil {
			return
		}
		if cursor == 0 { // over
			return
		}
	}
}

func (rc *Cache) HmSet(key string, v interface{}) error {
	_, err := rc.do("HMSET", redis.Args{}.Add(key).AddFlat(v)...)
	if err != nil {
		return err
	}
	return nil
}

func (rc *Cache) HmGet(key, v string) string {
	if value, err := redis.Values(rc.do("HMGET", key, v)); err == nil {
		if value == nil {
			return ""
		} else {
			for _, o := range value {
				if o == nil {
					return ""
				}
				return string(o.([]byte))
			}
		}
	}
	return ""
}

func (rc *Cache) HmGetAll(key string) ([]interface{}, error) {
	return redis.Values(rc.do("HGETALL", key))
}

func (rc *Cache) ExistsScript(sha string) (error, int64) {
	c := rc.p.Get()
	defer c.Close()
	data, err := redis.Values(c.Do("SCRIPT", "EXISTS", sha))
	if err != nil {
		return err, 0
	}
	for _, o := range data {
		return nil, o.(int64)
	}
	return nil, 0
}

func (rc *Cache) Script(sha string) (error, string) {
	c := rc.p.Get()
	defer c.Close()
	nameByte, err := c.Do("SCRIPT", "LOAD", sha)
	if err != nil {
		return err, ""
	}
	return nil, string(nameByte.([]byte))
}

func (rc *Cache) EvalSha(sha string, data ...interface{}) (interface{}, error) {
	c := rc.p.Get()
	defer c.Close()

	var args []interface{}
	args = make([]interface{}, 2+len(data))
	args[0] = sha
	args[1] = len(data)
	copy(args[2:], data)

	val, err := c.Do("EVALSHA", args...)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Set set cache to redis.
func (rc *Cache) Set(key string, val interface{}) error {
	_, err := rc.do("SET", key, val)
	return err
}

// associate with config key.
func (rc *Cache) associate(originKey interface{}) string {
	return fmt.Sprintf("%s:%s", rc.key, originKey)
}

func (rc *Cache) Lock(key string, timeout time.Duration) (bool, error) {
	res, err := rc.do("SET", key, 1, "EX", int64(timeout/time.Second), "NX")
	if err != nil {
		return false, err
	}
	return res == "OK", nil
}

func (rc *Cache) UnLock(key string) error {
	_, err := rc.do("del", key)
	return err
}
