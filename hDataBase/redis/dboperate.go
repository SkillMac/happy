package redis

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

type DbCfg struct {
	Host        string
	Port        string
	Pwd         string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
}

func NewDbCfg() *DbCfg {
	return &DbCfg{
		Host:        "",
		Port:        "",
		Pwd:         "",
		MaxIdle:     0,
		IdleTimeout: 0,
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
	addres := "101.200.51.151:6379"
	pool := &redis.Pool{
		Wait:        true,
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		MaxActive:   50,
		Dial: func() (conn redis.Conn, err error) {
			conn, err = redis.Dial("tcp", addres)
			if err != nil {
				return nil, err
			}
			if _, err = conn.Do("AUTH", "vdonggames123"); err != nil {
				conn.Close()
				return nil, err
			}
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

func (this *Rds) Get() {

}
