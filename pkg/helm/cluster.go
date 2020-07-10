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

// Helm represents all current releases that we can find in the cluster
type Helm struct {
	Version string
	Kube    *kube
}

// NewHelm returns a basic helm struct with the version of helm requested
func NewHelm(version string) *Helm {
	return &Helm{
		Version: version,
		Kube:    getConfigInstance(),
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

	klog.Infof("Got %d installed releases in the cluster", len(deployed))
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
				Deprecated: chart.Chart.Metadata.Deprecated,
				IsOld:      version.Compare(newest.Version, chart.Chart.Metadata.Version, ">"),
			}
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

	klog.Infof("Got %d installed releases in the cluster", len(deployed))
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
				Deprecated: chart.Chart.Metadata.Deprecated,
				IsOld:      version.Compare(newest.Version, chart.Chart.Metadata.Version, ">"),
			}
			outputObjects = append(outputObjects, rls)
		}
	}

	return outputObjects, err
}

// GetReleaseOutput return the expected output or error
func GetReleaseOutput(version string, repos []*Repo) (outputObjects []output.ReleaseOutput, err error) {

	switch version {
	case "2":
		h := NewHelm(version)
		outputObjects, err = h.GetHelmReleasesVersion2(repos)
	case "3":
		h := NewHelm(version)
		outputObjects, err = h.GetHelmReleasesVersion3(repos)
	case "auto":
		h := NewHelm("3")
		outputObjectsVersion3, err3 := h.GetHelmReleasesVersion3(repos)
		if outputObjectsVersion3 != nil {
			outputObjects = append(outputObjects, outputObjectsVersion3...)
		}
		h2 := NewHelm("2")

		outputObjectsVersion2, err2 := h2.GetHelmReleasesVersion2(repos)
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
