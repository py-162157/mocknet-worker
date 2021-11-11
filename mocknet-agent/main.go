package main

import (
	"os"
	"time"

	"mocknet/plugins/controller"
	"mocknet/plugins/etcd"
	"mocknet/plugins/vpp"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/logging/logmanager"
	"go.ligato.io/cn-infra/v2/logging/logrus"
)

const (
	defaultStartupTimeout = 45 * time.Second
)

type MocknetAgent struct {
	name       string
	LogManager *logmanager.Plugin

	ETCD       *etcd.Plugin
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

	ControllerPlugin := controller.NewPlugin(controller.UseDeps(func(deps *controller.Deps) {
		deps.Vpp = VppPlugin
	}))

	EtcdPlugin := etcd.NewPlugin(etcd.UseDeps(func(deps *etcd.Deps) {
	}))

	mocknetAgent := &MocknetAgent{
		name:       "mocknetagent",
		LogManager: LogmanagerPlugin,
		ETCD:       EtcdPlugin,
		VPP:        VppPlugin,
		Controller: ControllerPlugin,
	}

	a := agent.NewAgent(agent.AllPlugins(mocknetAgent), agent.StartTimeout(getStartupTimeout()))
	if err := a.Run(); err != nil {
		logrus.DefaultLogger().Fatal(err)
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
