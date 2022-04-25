package main

import (
	"os"
	"time"

	"mocknet/plugins/controller"
	"mocknet/plugins/linux"
	"mocknet/plugins/vpp"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/logging/logmanager"
	"go.ligato.io/cn-infra/v2/logging/logrus"
)

const (
	defaultStartupTimeout = 180 * time.Second
)

type MocknetAgent struct {
	name       string
	LogManager *logmanager.Plugin

	VPP        *vpp.Plugin
	Controller *controller.Plugin
}

func (ma *MocknetAgent) String() string {
	return "MocknetAgent"
}

func (ma *MocknetAgent) Init() error {
	return nil
}

func (ma *MocknetAgent) Close() error {
	return nil
}

func main() {

	LogmanagerPlugin := &logmanager.DefaultPlugin

	VppPlugin := vpp.NewPlugin(vpp.UseDeps(func(deps *vpp.Deps) {
	}))

	LinuxPlugin := linux.NewPlugin(linux.UseDeps(func(deps *linux.Deps) {
		deps.Vpp = VppPlugin
	}))

	ControllerPlugin := controller.NewPlugin(controller.UseDeps(func(deps *controller.Deps) {
		deps.Vpp = VppPlugin
		deps.Linux = LinuxPlugin
	}))

	mocknetAgent := &MocknetAgent{
		name:       "mocknetagent",
		LogManager: LogmanagerPlugin,
		VPP:        VppPlugin,
		Controller: ControllerPlugin,
	}

	a := agent.NewAgent(agent.AllPlugins(mocknetAgent), agent.StartTimeout(getStartupTimeout()))
	if err := a.Start(); err != nil {
		logrus.DefaultLogger().Fatal(err)
	}

	for {
		time.Sleep(time.Second)
		//logrus.DefaultLogger().Infoln("ControllerPlugin.CloseSignal =", ControllerPlugin.CloseSignal)
		if ControllerPlugin.CloseSignal {
			a.Stop()
			break
		}
	}
}

func getStartupTimeout() time.Duration {
	var err error
	var timeout time.Duration

	// valid env value must conform to duration format
	// e.g: 45s
	envVal := os.Getenv("STARTUPTIMEOUT")

	if timeout, err = time.ParseDuration(envVal); err != nil {
		timeout = defaultStartupTimeout
	}

	return timeout
}
