package redis

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

type DbCfg struct {
	Host        string
	Port        int
	Pwd         string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
	DbNum       int
}

func NewDbCfg(host string, port int, pwd string, maxIdle, idleTimeout, dbNum int) *DbCfg {
	return &DbCfg{
		Host:        host,
		Port:        port,
		Pwd:         pwd,
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		DbNum:       dbNum,
	}
}

type Rds struct {
	pool *redis.Pool
}

func NewRds(cfg *DbCfg) *Rds {
	//return &Rds{pool: struct {
	//	//Dial            func() (redis.Conn, error)
	//	//DialContext     func(ctx context.Context) (redis.Conn, error)
	//	//TestOnBorrow    func(c redis.Conn, t time.Time) error
	//	//MaxIdle         int
	//	//MaxActive       int
	//	//IdleTimeout     time.Duration
	//	//Wait            bool
	//	//MaxConnLifetime time.Duration
	//	//chInitialized   uint32
	//	//mu              sync.Mutex
	//	//closed          bool
	//	//active          int
	//	//ch              chan struct{}
	//	//idle            interface{}
	//	//waitCount       int64
	//	//waitDuration    time.Duration
	//	Max
	//}}
	// 3 ,50
	// "vdonggames123"
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	pool := &redis.Pool{
		Wait:        true,
		MaxIdle:     cfg.MaxIdle,
		IdleTimeout: time.Duration(cfg.IdleTimeout) * time.Second,
		MaxActive:   cfg.MaxActive,
		Dial: func() (conn redis.Conn, err error) {
			conn, err = redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			if _, err = conn.Do("AUTH", cfg.Pwd); err != nil {
				_ = conn.Close()
				return nil, err
			}

			_, selecterr := conn.Do("SELECT", cfg.DbNum)
			if selecterr != nil {
				_ = conn.Close()
				return nil, selecterr
			}
			fmt.Println("Redis 连接成功")
			return conn, err
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			if err != nil {
				panic(err)
			}
			return err
		},
	}
	return &Rds{pool: pool}
}

func (this *Rds) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("missing required arguments")
	}
	c := this.pool.Get()
	defer c.Close()

	return c.Do(commandName, args...)
}

func (this *Rds) Get(tbName string) (interface{}, error) {
	return this.do("GET", tbName)
}

func (this *Rds) Set(tbName string, val interface{}) (interface{}, error) {
	return this.do("SET", tbName, val)
}

func (this *Rds) Del(tbName string) (interface{}, error) {
	return this.do("DEL", tbName)
}

/*
* 可以添加多个成员 一个是  score key ...
 */
func (this *Rds) ZAdd(tbName string, score int, key string) (interface{}, error) {
	return this.do("ZADD", tbName, score, key)
}

func (this *Rds) ZLen(tbName string) (interface{}, error) {
	return this.do("ZCARD", tbName)
}

func (this *Rds) ZLenInRange(tbName string, min, max int) (interface{}, error) {
	return this.do("ZCOUNT", tbName, min, max)
}

//ZINCRBY
func (this *Rds) ZIncrby(tbName string, addScore int, key string) (interface{}, error) {
	return this.do("ZINCRBY", tbName, addScore, key)
}

// ZRANGE salary 0 -1 WITHSCORES
// start 和 stop 都以 0 为底，也就是说，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，以此类推。
// 从小到大排序
func (this *Rds) ZRange(tbName string, startIndex, endIndex int, isWithScores bool) (interface{}, error) {
	if isWithScores {
		return this.do("ZRANGE", tbName, startIndex, endIndex, "WITHSCORES")
	} else {
		return this.do("ZRANGE", tbName, startIndex, endIndex)
	}
}

// ZREVRANGE key start stop [WITHSCORES]
func (this *Rds) ZRevRange(tbName string, startIndex, endIndex int, isWithScores bool) (interface{}, error) {
	if isWithScores {
		return this.do("ZREVRANGE", tbName, startIndex, endIndex, "WITHSCORES")
	} else {
		return this.do("ZREVRANGE", tbName, startIndex, endIndex)
	}
}

// 从小到大排名
func (this *Rds) ZShowSlefRanking(tbName string, key string) (interface{}, error) {
	return this.do("ZRANGE", tbName, key)
}

// 从高到低排序
// ZREVRANK key member
func (this *Rds) ZShowSlefRevRanking(tbName string, key string) (interface{}, error) {
	return this.do("ZREVRANK", tbName, key)
}

// 	ZSCORE key member
func (this *Rds) ZScore(tbName string, key string) (interface{}, error) {
	return this.do("ZSCORE", tbName, key)
}

// ZREM key member [member ...]
// 移除成员
func (this *Rds) ZRemove(tbName string, keys ...interface{}) (interface{}, error) {
	return this.do("ZREM", tbName, keys)
}

// ZREMRANGEBYRANK key start stop
// 包含关系 // 降序
func (this *Rds) ZRemoveByIndex(tbName string, startIndex, endIndex int) (interface{}, error) {
	return this.do("ZREMRANGEBYRANK", tbName, startIndex, endIndex)
}

//ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
// 通过分数排序
