package kubernetes

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"mocknet/plugins/server/rpctest"

	appsv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.ligato.io/cn-infra/v2/logging"
	"k8s.io/client-go/kubernetes"
	k8s_rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Plugin struct {
	Deps

	PluginName       string
	K8sNamespace     string
	ClientSet        *kubernetes.Clientset
	MnNameReflector  map[string]string // key: pods name in mininet, value: pods name in k8s
	K8sNameReflector map[string]string // key: pods name in k8s, value: pods name in mininet
}

type Deps struct {
	Log        logging.PluginLogger
	KubeConfig *k8s_rest.Config
}

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	p.PluginName = "k8s"
	p.K8sNamespace = "default"
	p.K8sNameReflector = make(map[string]string)
	p.MnNameReflector = make(map[string]string)

	// get current kubeconfig
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// uses the current context in kubeconfig
	if config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
		p.Log.Errorln("failed to build config from flags")
		panic(err.Error())
	} else {
		p.KubeConfig = config
	}

	// create the clientset
	if clientset, err := kubernetes.NewForConfig(p.KubeConfig); err != nil {
		p.Log.Errorln("failed to create the clientset")
		panic(err.Error())
	} else {
		p.ClientSet = clientset
	}

	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	return nil
}

func (p *Plugin) String() string {
	return "kubernetes"
}

func (p *Plugin) Close() error {
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" { // linux
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func (p *Plugin) Create_Deployment(message rpctest.Message) error {
	//etcd_config_map := make_configmap()
	deployment := make_deployment(int32(len(message.Command.EmunetCreation.Emunet.Pods))) // 获取节点的个数并创建相应deployment数据

	/*if _, err := p.ClientSet.CoreV1().ConfigMaps(p.K8sNamespace).Create(&etcd_config_map); err != nil {
		p.Log.Errorln("failed to create configmap:", err)
		panic(err)
	} else {
		p.Log.Infoln("successfully created configmap!")
	}*/

	if _, err := p.ClientSet.AppsV1().Deployments(p.K8sNamespace).Create(&deployment); err != nil {
		p.Log.Errorln("failed to create deployment:", err)
		panic(err)
	} else {
		p.Log.Infoln("successfully created deployment!")
	}
	p.Log.Infoln("successfully created pods!")

	var pods *coreV1.PodList
	for {
		temp_pods, err := p.ClientSet.CoreV1().Pods(p.K8sNamespace).List(metav1.ListOptions{})
		if err != nil {
			p.Log.Errorln(err)
			panic(err)
		}
		if len(temp_pods.Items) != len(message.Command.EmunetCreation.Emunet.Pods)+1 {
			p.Log.Infoln("waiting for pods creation finished")
		} else {
			p.Log.Infoln("pods creation finished!")
			pods = temp_pods
			break
		}
		time.Sleep(time.Second)
	}

	i := 0
	for _, pod := range message.Command.EmunetCreation.Emunet.Pods {
		p.K8sNameReflector[pods.Items[i].Name] = pod.Name
		p.MnNameReflector[pod.Name] = pods.Items[i].Name
		i++
	}

	return nil
}

func make_configmap() coreV1.ConfigMap {
	EtcdConfData :=
		`insecure-transport: true
dial-timeout: 10000000000
allow-delayed-start: true
endpoints:`
	return coreV1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-cfg",
			Namespace: "default",
			Labels: map[string]string{
				"name": "etcd-cfg",
			},
		},
		Data: map[string]string{
			"etcd.conf": EtcdConfData,
		},
	}
}

func make_deployment(replica int32) appsv1.Deployment {
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mocknet-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mocknet",
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mocknet",
					},
					Annotations: map[string]string{
						"contivpp.io/custom-if":          "memif1/memif/stub, memif2/memif/stub, memif3/memif/stub, memif4/memif/stub, memif5/memif/stub",
						"contivpp.io/microservice-label": "mocknet",
					},
				},
				Spec: coreV1.PodSpec{
					NodeSelector: map[string]string{
						"mocknetrole": "worker",
					},
					Containers: []coreV1.Container{
						{
							Name:  "vpp-agent",
							Image: "ligato/vpp-agent:latest",
							Env: []coreV1.EnvVar{
								{
									Name:  "ETCD_CONFIG",
									Value: "/etc/etcd/etcd.conf",
								},
								{
									Name:  "MICROSERVICE_LABEL",
									Value: "mocknet",
								},
								{
									Name: "MY_HOST_IP",
									ValueFrom: &coreV1.EnvVarSource{
										FieldRef: &coreV1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
							},
							VolumeMounts: []coreV1.VolumeMount{
								{
									Name:      "etcd-cfg",
									MountPath: "/etc/etcd",
								},
							},
						},
					},
					Volumes: []coreV1.Volume{
						{
							Name: "etcd-cfg",
							VolumeSource: coreV1.VolumeSource{
								HostPath: &coreV1.HostPathVolumeSource{
									Path: "/opt/etcd",
								},
							},
						},
					},
				},
			},
		},
	}
}
