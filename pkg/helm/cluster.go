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
	helmstoragev2 "helm.sh/helm/pkg/storage"
	driverv2 "helm.sh/helm/pkg/storage/driver"
	helmstoragev3 "helm.sh/helm/v3/pkg/storage"
	driverv3 "helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/klog"
)

// Helm contains a helm version and kubernetes client interface
type Helm struct {
	Version         string
	Kube            *kube
	DesiredVersions []DesiredVersion
}

// DesiredVersion is a specific desired version that overrides the latest from the repository
type DesiredVersion struct {
	Name    string
	Version string
}

// NewHelm returns a basic helm struct with the version of helm requested
func NewHelm(version string, kubeContext string) *Helm {
	return &Helm{
		Version: version,
		Kube:    getConfigInstance(kubeContext),
	}
}

// GetHelmReleasesVersion3 returns a collection of deployed helm version 3 charts in a cluster.
func (h *Helm) GetHelmReleasesVersion3(helmRepos []*Repo) ([]output.ReleaseOutput, error) {
	outputObjects := []output.ReleaseOutput{}

	hs := driverv3.NewSecrets(h.Kube.Client.CoreV1().Secrets(""))
	helmClient := helmstoragev3.Init(hs)
	deployed, err := helmClient.ListDeployed()

	if err != nil {
		return nil, err
	}

	klog.V(5).Infof("Got %d installed releases in the cluster", len(deployed))
	for _, chart := range deployed {
		validRepos := IsRepoIncluded(chart.Chart.Metadata.Name, helmRepos)
		newest := TryToFindNewestReleaseByChart(chart, validRepos)
		if newest != nil {
			rls := output.ReleaseOutput{
				ReleaseName: chart.Name,
				ChartName:   chart.Chart.Metadata.Name,
				Namespace:   chart.Namespace,
				Description: chart.Chart.Metadata.Description,
				Icon:        chart.Chart.Metadata.Icon,
				Home:        chart.Chart.Metadata.Home,
				Installed: output.VersionInfo{
					Version:    chart.Chart.Metadata.Version,
					AppVersion: chart.Chart.Metadata.AppVersion,
				},
				Latest: output.VersionInfo{
					Version:    newest.Version,
					AppVersion: newest.AppVersion,
				},
				HelmVersion: "v3",
				Deprecated:  chart.Chart.Metadata.Deprecated,
			}
			h.overrideDesiredVersion(&rls)
			rls.IsOld = version.Compare(rls.Latest.Version, chart.Chart.Metadata.Version, ">")
			outputObjects = append(outputObjects, rls)
		}
	}
	return outputObjects, err
}

// GetHelmReleasesVersion2 returns a collection of deployed helm version 2 charts in a cluster.
func (h *Helm) GetHelmReleasesVersion2(helmRepos []*Repo) ([]output.ReleaseOutput, error) {
	outputObjects := []output.ReleaseOutput{}
	hcm := driverv2.NewConfigMaps(h.Kube.Client.CoreV1().ConfigMaps(""))
	helmClient := helmstoragev2.Init(hcm)
	deployed, err := helmClient.ListDeployed()

	if err != nil {
		return nil, err
	}

	klog.V(5).Infof("Got %d installed releases in the cluster", len(deployed))
	for _, chart := range deployed {
		validRepos := IsRepoIncluded(chart.Chart.Metadata.Name, helmRepos)

		newest := TryToFindNewestReleaseByChartVersion2(chart, validRepos)
		if newest != nil {
			rls := output.ReleaseOutput{
				ReleaseName: chart.Name,
				ChartName:   chart.Chart.Metadata.Name,
				Namespace:   chart.Namespace,
				Description: chart.Chart.Metadata.Description,
				Icon:        chart.Chart.Metadata.Icon,
				Home:        chart.Chart.Metadata.Home,
				Installed: output.VersionInfo{
					Version:    chart.Chart.Metadata.Version,
					AppVersion: chart.Chart.Metadata.AppVersion,
				},
				Latest: output.VersionInfo{
					Version:    newest.Version,
					AppVersion: newest.AppVersion,
				},
				HelmVersion: "v2",
				Deprecated:  chart.Chart.Metadata.Deprecated,
			}
			h.overrideDesiredVersion(&rls)
			rls.IsOld = version.Compare(rls.Latest.Version, chart.Chart.Metadata.Version, ">")
			outputObjects = append(outputObjects, rls)
		}
	}

	return outputObjects, err
}

func (h *Helm) overrideDesiredVersion(rls *output.ReleaseOutput) {
	for _, override := range h.DesiredVersions {
		if rls.ChartName == override.Name {
			klog.V(3).Infof("using override: %s=%s", rls.ChartName, override.Version)
			rls.Latest = output.VersionInfo{
				Version:    override.Version,
				AppVersion: "",
			}
			rls.Overridden = true
		}
	}
}

// GetReleaseOutput return the expected output or error
func (h *Helm) GetReleaseOutput(repos []*Repo) (outputObjects []output.ReleaseOutput, err error) {
	switch h.Version {
	case "2":
		outputObjects, err = h.GetHelmReleasesVersion2(repos)
	case "3":
		outputObjects, err = h.GetHelmReleasesVersion3(repos)
	case "auto":
		outputObjectsVersion3, err3 := h.GetHelmReleasesVersion3(repos)
		if outputObjectsVersion3 != nil {
			outputObjects = append(outputObjects, outputObjectsVersion3...)
		}

		outputObjectsVersion2, err2 := h.GetHelmReleasesVersion2(repos)
		if outputObjectsVersion2 != nil {
			outputObjects = append(outputObjects, outputObjectsVersion2...)
		}

		if err2 != nil && err3 != nil {
			err = fmt.Errorf("Could not detect helm 2 or helm 3 charts.\nHelm 2: %v\nHelm 3: %v", err2, err3)
		}

	default:
		err = fmt.Errorf("helm version either not specified or incorrect (use 2,3 or auto)")
	}
	return

}
