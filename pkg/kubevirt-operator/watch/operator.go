/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017, 2018 Red Hat, Inc.
 *
 */

package watch

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	//	virtv1 "kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/controller"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
)

type VirtOperatorController struct {
	clientset        kubecli.KubevirtClient
	Queue            workqueue.RateLimitingInterface
	kubevirtInformer cache.SharedIndexInformer
	recorder         record.EventRecorder
	//podExpectations    *controller.UIDTrackingControllerExpectations
}

func NewVirtOperatorController(kubevirtInformer cache.SharedIndexInformer,
	recorder record.EventRecorder,
	clientset kubecli.KubevirtClient) *VirtOperatorController {

	c := &VirtOperatorController{
		Queue:            workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		kubevirtInformer: kubevirtInformer,
		recorder:         recorder,
		clientset:        clientset,
		//podExpectations:      controller.NewUIDTrackingControllerExpectations(controller.NewControllerExpectations()),
	}

	c.kubevirtInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addKubevirt,
		DeleteFunc: c.deleteKubevirt,
		UpdateFunc: c.updateKubevirt,
	})

	return c
}

func (c *VirtOperatorController) Run(threadiness int, stopCh chan struct{}) {
	defer controller.HandlePanic()
	defer c.Queue.ShutDown()
	log.Log.Info("Starting KubeVirtOperator controller.")

	// Wait for cache sync before we start the pod controller
	cache.WaitForCacheSync(stopCh, c.kubevirtInformer.HasSynced)

	// Start the actual work
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Log.Info("Stopping KubeVirtOperator controller.")
}

func (c *VirtOperatorController) runWorker() {
	for c.Execute() {
	}
}

func (c *VirtOperatorController) Execute() bool {
	key, quit := c.Queue.Get()
	if quit {
		return false
	}
	defer c.Queue.Done(key)
	err := c.execute(key.(string))

	if err != nil {
		log.Log.Reason(err).Infof("reenqueuing VirtualMachineInstance %v", key)
		c.Queue.AddRateLimited(key)
	} else {
		log.Log.V(4).Infof("processed VirtualMachineInstance %v", key)
		c.Queue.Forget(key)
	}
	return true
}

func (c *VirtOperatorController) execute(key string) error {

	// Fetch the latest Vm state from cache
	// obj, exists, err := c.kubevirtInformer.GetStore().GetByKey(key)
	// if err != nil {
	// 	return err
	// }

	// Once all finalizers are removed the virtoperator gets deleted and we can clean all expectations //
	// if !exists {
	// 	c.podExpectations.DeleteExpectations(key)
	// 	return nil
	// }

	// kubevirt := obj.(*virtv1.KubeVirt)
	// logger := log.Log.Object(kubevirt)

	return nil
}

func (c *VirtOperatorController) addKubevirt(obj interface{})                      {}
func (c *VirtOperatorController) deleteKubevirt(obj interface{})                   {}
func (c *VirtOperatorController) updateKubevirt(obj interface{}, obj2 interface{}) {}
