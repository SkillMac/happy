package hBaseComponent

import "custom/happy/hLog"

type LauncherService struct {
	launcher *LauncherComponent
}

func (this *LauncherService) init(launcher *LauncherComponent) {
	this.launcher = launcher
}

func (this *LauncherService) CloseSelf(name string, reply *bool) error {
	hLog.Info("Launcher Service Close Server", name)
	this.launcher.Close <- struct{}{}
	return nil
}
