package main

import (
	"../../happy"
	"../hBaseComponent"
	"../hECS"
	"./api"
)

func main() {
	launcher := happy.DefaultServer()

	gate := &hBaseComponent.DefaultGateComponent{}

	gate.AddNetAPI(api.NewTestApi())
	launcher.AddComponentGroup("gate", []hEcs.IComponent{gate})
	//launcher.AddComponentGroup("")
	launcher.Serve()
}
