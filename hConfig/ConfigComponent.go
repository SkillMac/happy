package hConfig

import (
	"custom/happy/hCommon"
	"custom/happy/hDataBase/mongo"
	"custom/happy/hECS"
	"custom/happy/hLog"
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
		LogLevel: hLog.DEBUG,
		LogPath:  "./log",
		LogMode:  hLog.ROLLFILE,
		//LogFileUnit:     hLog.MB,
		LogFileMax:      10,
		LogFileSizeMax:  10, // 这里默认单位设置成了 MB
		LogConsolePrint: true,
	}
	this.ClusterConfig = &ClusterConfig{
		MasterAddress: "127.0.0.1:6666",
		LocalAddress:  "127.0.0.1:6666",
		AppName:       "defaultApp",
		Role:          []string{"single"},
		NodeDefine: map[string]Node{
			/*
				内置角色：master、child、location
			*/
			//master节点
			"node_master": {LocalAddress: "0.0.0.0:6666", Role: []string{"single"}},
			//位置服务节点
			//"node_location": {LocalAddress: "0.0.0.0:6603", Role: []string{"location"}},

			//用户自定义
			//"node_gate":  {LocalAddress: "0.0.0.0:6601", Role: []string{"gate"}},
			//"node_login": {LocalAddress: "0.0.0.0:6602", Role: []string{"login"}},
			//"node_room":  {LocalAddress: "0.0.0.0:6605", Role: []string{"room"}},

			//dubug 或 单服
			"node_single": {LocalAddress: "0.0.0.0:6666", Role: []string{"master", "gate"}}, /*, "gate", "login", "room"}},*/
		},

		ReportInterval:       3000,
		RpcTimeout:           9000,
		RpcCallTimeout:       5000,
		RpcHeartBeatInterval: 3000,
		IsLocationMode:       true,
		LocationSyncInterval: 500,

		NetConnTimeout:   9000,
		NetListenAddress: "127.0.0.1:5555",

		//IsActorModel: true,
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
	}
}

/*
	Default config
*/
type CommonConfig struct {
	Debug            bool          //是否为Debug模式
	RuntimeMaxWorker int           //runtime最大工作线程
	LogLevel         hLog.LEVEL    //log等级
	LogPath          string        //log的存储根目录
	LogMode          hLog.ROLLTYPE //log文件存储模式，分为按文件大小分割，按日期分割
	//LogFileUnit      hLog.UNIT     //log文件大小单位
	LogFileMax      int32 // log文件最大值
	LogFileSizeMax  int64
	LogConsolePrint bool //是否输出log到控制台

}
type Node struct {
	LocalAddress string
	Role         []string
}

type ClusterConfig struct {
	MasterAddress string   //Master 地址,例如:127.0.0.1:8888
	LocalAddress  string   //本节点IP,注意配置文件时，填写正确的局域网地址或者外网地址，不可为0.0.0.0
	AppName       string   //本节点拥有的app
	Role          []string //本节点拥有角色
	NodeDefine    map[string]Node

	ReportInterval       int  //子节点节点信息上报间隔，单位秒
	RpcTimeout           int  //tcp链接超时，单位毫秒
	RpcCallTimeout       int  //rpc调用超时
	RpcHeartBeatInterval int  //tcp心跳间隔
	IsLocationMode       bool //是否启用位置服务器
	LocationSyncInterval int  //位置服务同步间隔，单位秒

	//外网
	NetConnTimeout   int    //外网链接超时
	NetListenAddress string //网关对外服务地址
}

type CustomConfig struct {
	Mongo mongo.DbCfg
	Email hCommon.EmailParam
}
