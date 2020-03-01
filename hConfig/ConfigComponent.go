package hConfig

import (
	"custom/happy/hCommon"
	"custom/happy/hDataBase/mongo"
	"custom/happy/hDataBase/redis"
	"custom/happy/hECS"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

var Config *ConfigComponent

type ConfigComponent struct {
	hEcs.ComponentBase
	commonConfigPath  string
	clusterConfigPath string
	customConfigPath  string
	CommonConfig      *CommonConfig
	ClusterConfig     *ClusterConfig
	CustomConfig      *CustomConfig
}

func (this *ConfigComponent) IsUnique() int {
	return hEcs.UNIQUE_TYPE_GLOBAL
}

func (this *ConfigComponent) Initialize() error {
	this.commonConfigPath = "./conf/CommonConfig.json"
	this.clusterConfigPath = "./conf/ClusterConfig.json"
	this.customConfigPath = "./conf/CustomeConfig.json"
	//初始化默认配置
	this.SetDefault()
	//读取配置文件
	this.ReloadConfig()
	//全局共享
	Config = this
	return nil
}

func (this *ConfigComponent) loadConfig(configpath string, cfg interface{}) error {
	data, err := ioutil.ReadFile(configpath)
	if err != nil {
		//文件不存在时创建配置文件，并写入默认值
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(configpath), 0666); err != nil {
				if os.IsPermission(err) {
					return err
				}
			}
			b, err := json.MarshalIndent(cfg, "", "    ")
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(configpath, b, 0666)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		err = json.Unmarshal(data, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

//重新读取配置文件，包括自定义配置文件
func (this *ConfigComponent) ReloadConfig() {
	err := this.loadConfig(this.commonConfigPath, this.CommonConfig)
	if err != nil {
		panic(err)
	}
	err = this.loadConfig(this.clusterConfigPath, this.ClusterConfig)
	if err != nil {
		panic(err)
	}
	if this.customConfigPath != "" {
		err = this.loadConfig(this.customConfigPath, this.CustomConfig)
		if err != nil {
			panic(err)
		}
	}
}

// config.CustomConfig[name] = structure
func (this *ConfigComponent) LoadCustomConfig(path string, structure interface{}) (err error) {
	kind := reflect.TypeOf(structure).Kind()
	if kind != reflect.Ptr && kind != reflect.Map {
		err = errors.New("structure must be pointer or map")
		return
	}
	err = this.loadConfig(path, structure)
	this.CustomConfig = nil
	this.customConfigPath = path
	return err
}

func (this *ConfigComponent) SetDefault() {
	this.CommonConfig = &CommonConfig{
		Debug: true,
		//runtime
		RuntimeMaxWorker: runtime.NumCPU(),
		//log
		LogLevel: 4,
		LogPath:  "./log",
		//LogFileUnit:     hLog.MB,
		LogFileMax:      10,
		LogFileSizeMax:  10, // 这里默认单位设置成了 MB
		LogConsolePrint: true,
	}
	this.ClusterConfig = &ClusterConfig{
		WorkId:        1,
		MasterAddress: "127.0.0.1:6666",
		LocalAddress:  "127.0.0.1:6666",
		AppName:       "defaultApp",
		WorkName:      "defaultWork",
		Role:          []string{"single"},
		NodeDefine: map[string]Node{
			/*
				内置角色：master、child、location
			*/
			//master节点
			"node_master": {WorkId: 1, LocalAddress: "0.0.0.0:6666", Role: []string{"single"}, NetAddr: NetAddr{Addr: "127.0.0.1:5555", Alias: ""}},
			//位置服务节点
			//"node_location": {LocalAddress: "0.0.0.0:6603", Role: []string{"location"}},

			//用户自定义
			//"node_gate":  {LocalAddress: "0.0.0.0:6601", Role: []string{"gate"}},
			//"node_login": {LocalAddress: "0.0.0.0:6602", Role: []string{"login"}},
			//"node_room":  {LocalAddress: "0.0.0.0:6605", Role: []string{"room"}},

			//dubug 或 单服
			"node_single": {WorkId: 1, LocalAddress: "0.0.0.0:6666", Role: []string{"master", "gate"}, NetAddr: NetAddr{
				Addr: "127.0.0.1:5555", Alias: "",
			}}, /*, "gate", "login", "room"}},*/
		},

		ReportInterval:       3000,
		RpcTimeout:           9000,
		RpcCallTimeout:       5000,
		RpcHeartBeatInterval: 3000,
		IsLocationMode:       true,
		LocationSyncInterval: 500,

		NetConnTimeout:         9000,
		NetListenAddress:       "127.0.0.1:5555",
		NetListenAddressAlias:  "",
		SelectNetListenAddress: "127.0.0.1:5556",
	}

	this.CustomConfig = &CustomConfig{
		Mongo: mongo.DbCfg{
			DbHost: "",
			DbPort: 0,
			DbName: "",
			DbUser: "",
			DbPass: "",
		},
		Email: hCommon.EmailParam{
			ServerHost: "smtp.163.com",
			ServerPort: 25,
			FromEmail:  "",
			FromPasswd: "",
			Toers:      "",
			CCers:      "",
		},
		Redis: redis.DbCfg{
			Host:        "",
			Port:        0,
			Pwd:         "",
			MaxIdle:     0,
			MaxActive:   0,
			IdleTimeout: 0,
			DbNum:       0,
		},
	}
}

/*
	Default config
*/
type CommonConfig struct {
	Debug            bool   //是否为Debug模式
	RuntimeMaxWorker int    //runtime最大工作线程
	LogLevel         int    //log等级
	LogPath          string //log的存储根目录
	LogFileMax       int    // log文件最大值
	LogFileSizeMax   int
	LogConsolePrint  bool //是否输出log到控制台

}
type Node struct {
	WorkId       int64
	LocalAddress string
	Role         []string
	NetAddr      NetAddr
}

type NetAddr struct {
	Addr  string
	Alias string
}

type ClusterConfig struct {
	WorkId        int64
	MasterAddress string   //Master 地址,例如:127.0.0.1:8888
	LocalAddress  string   //本节点IP,注意配置文件时，填写正确的局域网地址或者外网地址，不可为0.0.0.0
	AppName       string   //本节点拥有的app
	WorkName      string   // 运行当前app 的 work 名字
	Role          []string //本节点拥有角色
	NodeDefine    map[string]Node

	ReportInterval       int  //子节点节点信息上报间隔，单位秒
	RpcTimeout           int  //tcp链接超时，单位毫秒
	RpcCallTimeout       int  //rpc调用超时
	RpcHeartBeatInterval int  //tcp心跳间隔
	IsLocationMode       bool //是否启用位置服务器
	LocationSyncInterval int  //位置服务同步间隔，单位秒

	//外网
	NetConnTimeout         int    //外网链接超时
	NetListenAddress       string //网关对外服务地址
	NetListenAddressAlias  string //监听网关地址别名
	SelectNetListenAddress string //无状态网关对外服务地址
}

type CustomConfig struct {
	Mongo mongo.DbCfg
	Email hCommon.EmailParam
	Redis redis.DbCfg
}
