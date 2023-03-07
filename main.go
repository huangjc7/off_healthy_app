package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"
)

func main() {
	var err error
	var config *rest.Config

	kubeconfig := fmt.Sprintf("%s%s", os.Getenv("HOME"), "/.kube/config")
	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			panic(err.Error())
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//创建informer
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute*5)
	//对pod监听
	podInformer := informerFactory.Core().V1().Pods()

	informer := podInformer.Informer()
	podLister := podInformer.Lister()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	pod, err := podLister.Pods("kube-system").List(labels.Everything())
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range pod {
		//获取pod状态
		fmt.Printf("%d -> %v\n", k+1, v.Status.Phase)
		//获取pod重启次数
		for _, v := range v.Status.ContainerStatuses {
			fmt.Println(v.RestartCount)
		}
	}

	<-stopCh
}

func onAdd(obj interface{}) {
	pod := obj.(*v1.Pod) //断言 是否是deployment类型
	fmt.Println("add a pod:", pod.Name)
}

func onUpdate(old, new interface{}) {
	oldpod := old.(*v1.Pod) //断言 是否是deployment类型
	newpod := new.(*v1.Pod)
	fmt.Println("update pod:", oldpod.Name, newpod.Name)
}

func onDelete(obj interface{}) {
	pod := obj.(*v1.Pod) //断言 是否是deployment类型
	fmt.Println("delete a pod:", pod.Name)
}
