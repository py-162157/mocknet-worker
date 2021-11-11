package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	namespace := "kube-system"
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	fmt.Printf("There are %d pods in namespace %s\n", len(pods.Items), namespace)
	fmt.Print("the uid of pod is:", pods.Items[0].UID)

	/*mock_pod1 := coreV1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx1",
			Namespace: "default",
			Labels: map[string]string{
				"name": "nginx",
			},
		},
		Spec: coreV1.PodSpec{
			RestartPolicy: coreV1.RestartPolicyAlways,
			Containers: []coreV1.Container{
				coreV1.Container{
					Name:  "nginx1",
					Image: "nginx:latest",
					Ports: []coreV1.ContainerPort{
						coreV1.ContainerPort{
							ContainerPort: 80,
							Protocol:      coreV1.ProtocolTCP,
						},
					},
				},
			},
		},
	}

	mock_pod2 := coreV1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx2",
			Namespace: "default",
			Labels: map[string]string{
				"name": "nginx",
			},
		},
		Spec: coreV1.PodSpec{
			RestartPolicy: coreV1.RestartPolicyAlways,
			Containers: []coreV1.Container{
				coreV1.Container{
					Name:  "nginx2",
					Image: "nginx:latest",
					Ports: []coreV1.ContainerPort{
						coreV1.ContainerPort{
							ContainerPort: 80,
							Protocol:      coreV1.ProtocolTCP,
						},
					},
				},
			},
		},
	}*/

	/*pod1, err1 := clientset.CoreV1().Pods(namespace).Create(&mock_pod1)
	pod2, err2 := clientset.CoreV1().Pods(namespace).Create(&mock_pod2)
	if pod1 != nil && pod2 != nil {

	}
	if err1 != nil && err2 != nil {
		fmt.Println("create failed!")
	}*/

	/*configmap := make_configmap()
	_, err = clientset.CoreV1().ConfigMaps(namespace).Create(&configmap)
	if err != nil {
		fmt.Println("failed to create configmap!")
	} else {
		fmt.Println("successfully to created configmap!")
	}

	deployment := make_deployment(2)
	_, err = clientset.AppsV1().Deployments(namespace).Create(&deployment)
	if err != nil {
		fmt.Println("failed to create deployment!")
		fmt.Println(err)
	} else {
		fmt.Println("successfully to created deployment!")
	}*/
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" { // linux
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func make_configmap() coreV1.ConfigMap {
	EtcdConfData :=
		`insecure-transport: true
dial-timeout: 10000000000
allow-delayed-start: true
endpoints:
  - "contiv-etcd.kube-system.svc.cluster.local:12379"` //格式很重要，务必按照此来，否则vpp-agent会解析失败
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
						"mocknetworker": "true",
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
								ConfigMap: &coreV1.ConfigMapVolumeSource{
									LocalObjectReference: coreV1.LocalObjectReference{
										Name: "etcd-cfg",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
