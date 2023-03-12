package main

import (
	"fmt"
	podV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	deployV1 "k8s.io/client-go/listers/apps/v1"
	podInformerV1 "k8s.io/client-go/listers/core/v1"
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

func createInformer(clientset *kubernetes.Clientset, listTime time.Duration) (informers.SharedInformerFactory, deployV1.DeploymentLister, podInformerV1.PodLister) {
	//func createInformer(clientset *kubernetes.Clientset, listTime time.Duration) (informers.SharedInformerFactory, deployV1.DeploymentLister) {
	//创建informer
	informerFactory := informers.NewSharedInformerFactory(clientset, listTime)
	//对pod监听
	podInformer := informerFactory.Core().V1().Pods()
	deployInformer := informerFactory.Apps().V1().Deployments()

	pinformer := podInformer.Informer()
	podLister := podInformer.Lister()
	deployLister := deployInformer.Lister()
	pinformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	return informerFactory, deployLister, podLister
	//return informerFactory, deployLister

}
func offDeploymentUnHealthy(deployLister deployV1.DeploymentLister, namespace string, clientset *kubernetes.Clientset) error {
	deploy, err := deployLister.Deployments(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	for _, dp := range deploy {
		//判断deployment是否是属于开启状态 不等于0 为开启
		if dp.Status.UpdatedReplicas != 0 {
			//判断当前运行的pod  等于0就是未运行
			if dp.Status.AvailableReplicas == 0 {
				var replicas int32 = 0
				dp.Spec.Replicas = &replicas
				_, err := clientset.AppsV1().Deployments(dp.Namespace).Update(dp)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getDeploymentName(rsName string) {

}

func getNotReadyPods(podLister podInformerV1.PodLister, namespace string, restartNumber int32) error {
	pod, err := podLister.Pods(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	for _, p := range pod {
		//for _, containerStatuses := range p.Status.ContainerStatuses {
		//	//for _, owner := range p.OwnerReferences {
		//	//	//if containerStatuses.RestartCount >= restartNumber {
		//	//	//	//if containerStatuses.Ready == false {
		//	//	//	//	//containerName := make([]string, 1)
		//	//	//	//	//containerName = owner.Name
		//	//	//	//	//getDeploymentName(owner.Name)
		//	//	//	//}
		//	//	//}
		//	//}
		//}
	}
	return nil
}

func main() {
	clientset := getKubeConfig()
	stopCh := make(chan struct{})
	//informerFactory, deployLister := createInformer(clientset, time.Minute*5)
	informerFactory, deployLister, podLister := createInformer(clientset, time.Minute*5)
	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)
	err := getNotReadyPods(podLister, "", 5)
	if err != nil {
		log.Fatal(err)
	}
	if err := offDeploymentUnHealthy(deployLister, "hjc", clientset); err != nil {
		log.Fatal(err)
	}
	select {}
}

func onAdd(obj interface{}) {}

func onUpdate(old, new interface{}) {}

func onDelete(obj interface{}) {
	pod := obj.(*podV1.Pod) //断言 是否是deployment类型
	fmt.Printf("[off unhealthy app] Delete Pod Namespace:%s Pod Name: %s\n", pod.Namespace, pod.Name)
}
