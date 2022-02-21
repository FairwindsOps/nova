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
	"fmt"

	"github.com/fairwindsops/nova/pkg/output"
	version "github.com/mcuadros/go-version"
	"helm.sh/helm/v3/pkg/release"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	helmdriver "helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/klog/v2"
)

// Helm contains a helm version and kubernetes client interface
type Helm struct {
	Kube            *kube
	DesiredVersions []DesiredVersion
}

// DesiredVersion is a specific desired version that overrides the latest from the repository
type DesiredVersion struct {
	Name    string
	Version string
}

// NewHelm returns a basic helm struct with the version of helm requested
func NewHelm(kubeContext string) *Helm {
	return &Helm{
		Kube: getConfigInstance(kubeContext),
	}
}

// GetReleaseOutput returns releases and chart names
func (h *Helm) GetReleaseOutput() ([]*release.Release, []string, error) {
	var chartNames = []string{}
	outputObjects, err := h.GetHelmReleases()
	if err != nil {
		err = fmt.Errorf("could not detect helm 3 charts: %v", err)
	}
	if outputObjects != nil {
		chartNames = make([]string, len(outputObjects))
		for i, release := range outputObjects {
			chartNames[i] = release.Chart.Metadata.Name
		}
	}
	return outputObjects, chartNames, err
}

// GetHelmReleases returns a list of helm releases from the cluster
func (h *Helm) GetHelmReleases() ([]*release.Release, error) {
	hs := helmdriver.NewSecrets(h.Kube.Client.CoreV1().Secrets(""))
	helmClient := helmstorage.Init(hs)
	deployed, err := helmClient.ListDeployed()

	if err != nil {
		return nil, err
	}
	return deployed, nil
}

// OverrideDesiredVersion accepts a list of releases and overrides the version stored in the helm struct where required
func (h *Helm) OverrideDesiredVersion(rls *output.ReleaseOutput) {
	for _, override := range h.DesiredVersions {
		if rls.ChartName == override.Name {
			klog.V(3).Infof("using override: %s=%s", rls.ChartName, override.Version)
			rls.Latest = output.VersionInfo{
				Version:    override.Version,
				AppVersion: "",
			}
			rls.IsOld = version.Compare(rls.Installed.Version, override.Version, "<")
			rls.Overridden = true
		}
	}
}
