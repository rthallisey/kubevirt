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
	golog "log"

	"kubevirt.io/kubevirt/pkg/controller"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
	"kubevirt.io/kubevirt/pkg/util"
	//"kubevirt.io/kubevirt/pkg/virt-controller/leaderelectionconfig"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	k8coresv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	clientrest "k8s.io/client-go/rest"
)

const (
	controllerThreads = 3
)

type VirtOperatorControllerApp struct {
	clientSet       kubecli.KubevirtClient
	restClient      *clientrest.RESTClient
	informerFactory controller.KubeInformerFactory

	virtOperatorController *VirtOperatorController
	operatorRecorder       record.EventRecorder
	operatorInformer       cache.SharedIndexInformer

	kubevirtNamespace string
}

func Execute() {
	var err error
	var app VirtOperatorControllerApp = VirtOperatorControllerApp{}

	log.InitializeLogging("operator-controller")

	app.clientSet, err = kubecli.GetKubevirtClient()

	if err != nil {
		golog.Fatal(err)
	}

	app.restClient = app.clientSet.RestClient()

	// Bootstrapping. From here on the initialization order is important
	app.kubevirtNamespace, err = util.GetNamespace()
	if err != nil {
		golog.Fatalf("Error searching for namespace: %v", err)
	}
	app.informerFactory = controller.NewKubeInformerFactory(app.restClient, app.clientSet, app.kubevirtNamespace)

	app.operatorInformer = app.informerFactory.VirtOperator()
	app.operatorRecorder = app.getNewRecorder(k8sv1.NamespaceAll, "operator-controller")

	app.initCommon()
	app.Run()
}

func (oca *VirtOperatorControllerApp) getNewRecorder(namespace string, componentName string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&k8coresv1.EventSinkImpl{Interface: oca.clientSet.CoreV1().Events(namespace)})
	return eventBroadcaster.NewRecorder(scheme.Scheme, k8sv1.EventSource{Component: componentName})
}

func (oca *VirtOperatorControllerApp) initCommon() {
	oca.virtOperatorController = NewVirtOperatorController(oca.operatorInformer, oca.operatorRecorder, oca.clientSet)
}

func (oca *VirtOperatorControllerApp) Run() {
	// logger := log.Log

	stop := make(chan struct{})
	defer close(stop)

	// recorder := oca.getNewRecorder(k8sv1.NamespaceAll, leaderelectionconfig.DefaultEndpointName)

	oca.virtOperatorController.Run(controllerThreads, stop)
	panic("unreachable")
}
