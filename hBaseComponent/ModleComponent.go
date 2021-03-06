package hBaseComponent

import (
	"custom/happy/hCluster"
	"custom/happy/hConfig"
	"custom/happy/hDataBase/mongo"
	"custom/happy/hDataBase/redis"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"gopkg.in/mgo.v2"
	"reflect"
	"time"
)

var Modle *ModleComponent

type ModleComponent struct {
	hEcs.ComponentBase
	nodeComponent *hCluster.NodeComponent
	M             *mongo.DbOperate
	R             *redis.Rds
}

func (this *ModleComponent) IsUnique() int {
	return hEcs.UNIQUE_TYPE_GLOBAL
}

func (this *ModleComponent) GetRequire() map[*hEcs.Object][]reflect.Type {
	requires := make(map[*hEcs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&hConfig.ConfigComponent{}),
	}
	return requires
}

func (this *ModleComponent) Initialize() error {
	hLog.Info("MongoDB 数据数据库初始化")
	this.initDatabase()
	Modle = this
	return nil
}

func (this *ModleComponent) initDatabase() {
	this.initMongoDB()
	this.initRedis()
}

func (this *ModleComponent) initMongoDB() {
	if hConfig.Config.CustomConfig.Mongo.DbHost == "" {
		return
	}
	this.M = mongo.NewDbOperate(mongo.NewDbCfg(
		hConfig.Config.CustomConfig.Mongo.DbHost,
		hConfig.Config.CustomConfig.Mongo.DbPort,
		hConfig.Config.CustomConfig.Mongo.DbName,
		hConfig.Config.CustomConfig.Mongo.DbUser,
		hConfig.Config.CustomConfig.Mongo.DbPass,
	), 5*time.Second)

	_ = this.M.OpenDB(func(ms *mgo.Session) {
		// 一个连接大概占10M
		ms.SetPoolLimit((50))
	})
}

func (this *ModleComponent) initRedis() {
	if hConfig.Config.CustomConfig.Redis.Host == "" {
		return
	}
	this.R = redis.NewRds(redis.NewDbCfg(
		hConfig.Config.CustomConfig.Redis.Host,
		hConfig.Config.CustomConfig.Redis.Port,
		hConfig.Config.CustomConfig.Redis.Pwd,
		hConfig.Config.CustomConfig.Redis.MaxIdle,
		hConfig.Config.CustomConfig.Redis.IdleTimeout,
		hConfig.Config.CustomConfig.Redis.DbNum))
}

//func (this *ModleComponent) Awake(ctx *hEcs.Context) {
//	hLog.Info("Modle =========== Awake")
//}
//
//func (this *ModleComponent) Start(ctx *hEcs.Context) {
//	hLog.Info("Modle =========== Start")
//}
//
//func (this *ModleComponent) Update(ctx *hEcs.Context) {
//	hLog.Info("Modle =========== Update")
//}

func (this *ModleComponent) Destroy(context *hEcs.Context) {
	this.M.CloseDB()
}
