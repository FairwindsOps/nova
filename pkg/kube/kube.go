// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"net/http"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"

	// add all known auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Connection holds a kubernetes.interface as the Client parameter
type Connection struct {
	Client        kubernetes.Interface
	DynamicClient dynamic.Interface
	RESTMapper    meta.RESTMapper
}

var (
	kubeClient *Connection
	once       sync.Once
)

// GetConfigInstance returns a Kubernetes interface based on the current configuration
func GetConfigInstance(context string) *Connection {
	once.Do(func() {
		kubeClient = &Connection{
			Client:        getKubeClient(context),
			DynamicClient: getDynamicKubeClient(context),
			RESTMapper:    getRESTMapper(context),
		}
	})
	return kubeClient
}

func getKubeClient(context string) kubernetes.Interface {
	kubeConf, err := config.GetConfigWithContext(context)
	if err != nil {
		klog.Fatalf("error getting config with context %s: %v", context, err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		klog.Fatalf("error create kubernetes client: %v", err)
	}
	return clientset
}

func getDynamicKubeClient(context string) dynamic.Interface {
	kubeConf, err := config.GetConfigWithContext(context)
	if err != nil {
		klog.Fatalf("error getting config with context %s: %v", context, err)
	}
	dynamicClient, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		klog.Fatalf("error create dynamic kubernetes client: %v", err)
	}
	return dynamicClient
}

func getRESTMapper(context string) meta.RESTMapper {
	kubeConf, err := config.GetConfigWithContext(context)
	if err != nil {
		klog.Fatalf("error getting config with context %s: %v", context, err)
	}

	restMapper, err := apiutil.NewDynamicRESTMapper(kubeConf, &http.Client{})
	if err != nil {
		klog.Fatalf("Error creating REST Mapper: %v", err)
	}
	return restMapper
}

// SetAndGetMock sets the singleton's interface to use a fake ClientSet
func SetAndGetMock() *Connection {
	kc := Connection{
		Client:        fake.NewSimpleClientset(),
		DynamicClient: dfake.NewSimpleDynamicClient(runtime.NewScheme()),
		RESTMapper:    &meta.DefaultRESTMapper{},
	}
	SetInstance(kc)
	return &kc
}

// SetInstance allows the user to set the kubeClient singleton
func SetInstance(kc Connection) {
	kubeClient = &kc
}
