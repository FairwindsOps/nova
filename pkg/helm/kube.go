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

package helm

import (
	"os"
	"sync"
	"fmt"

	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	// add all known auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type kube struct {
	Client kubernetes.Interface
}

var kubeClient *kube
var once sync.Once

// Modify kubeconfig raw context
func SwitchContext(context string, kubeconfig string) (err error) {

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return err
	}
	if rawConfig.Contexts[context] == nil {
		return fmt.Errorf("context %s doesn't exists", context)
	}
	rawConfig.CurrentContext = context
	err = clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), rawConfig, true)
	if err != nil {
		return fmt.Errorf("Error modifying config at %s", kubeconfig)
	}
	return
}

// Get raw config file path
func getRawConfig(context string) (err error) {
	var kubeconfig string

	// Kubeconfig path, argo CLI implementation expects the default path
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = "/root/.kube/config"
	}

	err = SwitchContext(context, kubeconfig)

	if err != nil {
		return err
	}
	return
}

// GetConfigInstance returns a Kubernetes interface based on the current configuration
func getConfigInstance(context string, argo bool) *kube {
	once.Do(func() {
		if kubeClient == nil {
			kubeClient = &kube{
				Client: getKubeClient(context, argo),
			}
		}
	})
	return kubeClient
}

func getKubeClient(context string, argo bool) kubernetes.Interface {
	var clientset *kubernetes.Clientset

	if argo && len(context) > 0 {
		err := getRawConfig(context)
		if err != nil {
			klog.Errorf("error modifying config with context %s: %v", context, err)
			os.Exit(1)
		}
	}

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
