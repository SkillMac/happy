package main

import (
	"../../happy"
	"../hBaseComponent"
	"../hECS"
	"../hLog"
	"./api"
	"./components"
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

	launcher.AddComponentGroups(map[string][]hEcs.IComponent{
		"gate":  []hEcs.IComponent{gate},
		"login": []hEcs.IComponent{&components.LoginComponent{}},
		"match": []hEcs.IComponent{&components.MatchComponent{}},
		"room":  []hEcs.IComponent{&components.RoomManagerComponent{}},
	})
	launcher.Serve()
}
