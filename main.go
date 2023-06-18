package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset for K8s cluster using the config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	// informer factory for k8s pods resource type with resync period set to 10 mins
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute*30)

	// Event Handler for pod resource type to informer factory
	podInformer := informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)

			// compare the Pod Phase
			if oldPod.Status.Phase != newPod.Status.Phase {
				fmt.Println("Pod phase changed from: ", oldPod.Status.Phase, " to: ", newPod.Status.Phase)
			}
		},
	})

	// Start the informer Factory
	informerFactory.Start(stopCh)

	// wait for the stop signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
