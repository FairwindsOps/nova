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
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"

	// add all known auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Connection struct {
	Client kubernetes.Interface
}

var (
	kubeClient *Connection
	once       sync.Once
)

// GetConfigInstance returns a Kubernetes interface based on the current configuration
func GetConfigInstance(context string) *Connection {
	once.Do(func() {
		if kubeClient == nil {
			kubeClient = &Connection{
				Client: getKubeClient(context),
			}
		}
	})
	return kubeClient
}

func getKubeClient(context string) kubernetes.Interface {
	var clientset *kubernetes.Clientset

	kubeConf, err := config.GetConfigWithContext(context)
	if err != nil {
		klog.Errorf("error getting config with context %s: %v", context, err)
		os.Exit(1)
	}

	clientset, err = kubernetes.NewForConfig(kubeConf)
	if err != nil {
		klog.Errorf("error create kubernetes client: %v", err)
		os.Exit(1)
	}
	return clientset
}

// The functions below assist in testing with a fake kube client
// SetAndGetMock sets the singleton's interface to use a fake ClientSet
func SetAndGetMock() *Connection {
	kc := Connection{
		Client: fake.NewSimpleClientset(),
	}
	SetInstance(kc)
	return &kc
}

// SetInstance allows the user to set the kubeClient singleton
func SetInstance(kc Connection) {
	kubeClient = &kc
}
