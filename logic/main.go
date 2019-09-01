package main

import (
	"../../happy"
	"../hBaseComponent"
	"../hECS"
	"../hLog"
	"./api"
	"./logicComponents"
	"flag"
	"fmt"
)

func main() {
	var nodeConfigName string
	flag.StringVar(&nodeConfigName, "node", "", "node name")
	flag.Parse()

	launcher := happy.DefaultServer()

	if nodeConfigName != "" {
		launcher.OverrideNodeDefine(nodeConfigName)
		hLog.Info(fmt.Sprintf("Override node info: [ %s ]", nodeConfigName))
	}

	gate := &hBaseComponent.DefaultGateComponent{}

	gate.AddNetAPI(api.NewLogicApi())
	launcher.AddComponentGroup("gate", []hEcs.IComponent{gate})
	launcher.AddComponentGroup("login", []hEcs.IComponent{&logicComponents.LoginComponent{}})
	launcher.Serve()
}
