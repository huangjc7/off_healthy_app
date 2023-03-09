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

var err error
var config *rest.Config

func getKubeConfig() *kubernetes.Clientset {

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
	return clientset
}
func createInformer(clientset *kubernetes.Clientset, listTime time.Duration) {

}
func main() {
	clientset := getKubeConfig()
	createInformer(clientset, time.Minute*5)
	//写到这里啦

	//创建informer
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute*5)
	//对pod监听
	podInformer := informerFactory.Core().V1().Pods()
	deployInformer := informerFactory.Apps().V1().Deployments()

	pinformer := podInformer.Informer()
	//podLister := podInformer.Lister()
	//dinformer := deployInformer.Informer()
	deployLister := deployInformer.Lister()

	pinformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	//pod, err := podLister.Pods("default").List(labels.Everything())
	//if err != nil {
	//	log.Fatal(err)
	//}

	deploy, err := deployLister.Deployments("default").List(labels.Everything())
	if err != nil {
		log.Fatal(err)
	}

	//deploySlice := make([]string, 10, 10)

	//var percentage float32
	for _, dp := range deploy {
		//判断deployment是否是属于开启状态 不等于0 为开启
		if dp.Status.UpdatedReplicas != 0 {
			//判断当前运行的pod  等于0就是未运行
			if dp.Status.AvailableReplicas == 0 {
				var replicas int32 = 0
				dp.Spec.Replicas = &replicas
				_, err := clientset.AppsV1().Deployments("default").Update(dp)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	<-stopCh
}

func onAdd(obj interface{}) {
	//pod := obj.(*v1.Pod) //断言 是否是deployment类型
	//fmt.Println("add deployment:", pod.Name)
}

func onUpdate(old, new interface{}) {
	//oldpod := old.(*v1.Pod) //断言 是否是deployment类型
	//newpod := new.(*v1.Pod)
	//fmt.Println("update deployment:", oldpod.Name, newpod.Name)
}

func onDelete(obj interface{}) {
	pod := obj.(*v1.Pod) //断言 是否是deployment类型
	fmt.Println("[off unhealthy app] delete pod:", pod.Name)
}
