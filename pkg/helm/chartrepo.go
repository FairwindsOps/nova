package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/fairwindsops/nova/pkg/output"
	version "github.com/mcuadros/go-version"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// Repo represents a Helm chart Repo
type Repo struct {
	URL    string
	Charts *ChartReleases
}

// ChartReleases contains the chart releases of a helm repository
type ChartReleases struct {
	APIVersion string                    `yaml:"apiVersion"`
	Entries    map[string][]ChartRelease `yaml:"entries"`
}

// ChartRelease is a single chart version in a helm repository
type ChartRelease struct {
	APIVersion  string             `yaml:"apiVersion,omitempty"`
	AppVersion  string             `yaml:"appVersion"`
	Created     time.Time          `yaml:"created"`
	Description string             `yaml:"description"`
	Digest      string             `yaml:"digest,omitempty"`
	Maintainers []chart.Maintainer `yaml:"maintainers,omitempty"`
	Name        string             `yaml:"name"`
	Urls        []string           `yaml:"urls"`
	Version     string             `yaml:"version"`
	Home        string             `json:"home"`
	Sources     []string           `json:"sources"`
	Keywords    []string           `json:"keywords"`
	Icon        string             `json:"icon"`
	Deprecated  bool               `json:"deprecated"`
}

// NewRepos returns data about a helm chart repository, given its url
func NewRepos(urls []string) []*Repo {
	var repos []*Repo

	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup
	wg.Add(len(urls))

	klog.V(5).Infof("loading %d chart repositories", len(urls))

	for _, url := range urls {
		klog.V(8).Infof("loading chart repository: %s", url)
		go func(address string) {
			defer wg.Done()
			repo := &Repo{
				URL:    address,
				Charts: &ChartReleases{},
			}
			err := repo.loadReleases()
			if err != nil {
				klog.V(5).Infof("Could not load chart repo %s: %s", address, err)
			} else {
				mutex.Lock()
				repos = append(repos, repo)
				mutex.Unlock()
			}
		}(url)
	}

	wg.Wait()
	return repos
}

func (r *Repo) loadReleases() error {
	response, err := http.Get(fmt.Sprintf("%s/index.yaml", r.URL))
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, r.Charts)
	if err != nil {
		return err
	}
	return nil
}

// NewestVersion returns the newest chart release for the provided release name
func (r *Repo) NewestVersion(releaseName string) *ChartRelease {
	for name, entries := range r.Charts.Entries {
		if name == releaseName {
			var newest ChartRelease
			for _, release := range entries {
				if IsValidRelease(release.Version) {
					if newest.Version == "" {
						newest = release
					}

					foundNewer := version.Compare(release.Version, newest.Version, ">")
					if foundNewer {
						newest = release
					}
				}
			}
			return &newest
		}
	}
	return nil
}

// NewestChartVersion returns the newest chart release for the provided release name and version
func (r *Repo) NewestChartVersion(currentChart *chart.Metadata) *ChartRelease {
	for name, entries := range r.Charts.Entries {
		if name == currentChart.Name {
			var newest ChartRelease
			repoHasCurrentVersion := false
			for _, release := range entries {
				if IsValidRelease(release.Version) {
					if release.Version == currentChart.Version {
						repoHasCurrentVersion = checkChartsSimilarity(currentChart, &release)
					}

					foundNewer := version.Compare(release.Version, newest.Version, ">")
					if foundNewer {
						newest = release
					}
				}
			}
			if repoHasCurrentVersion {
				return &newest
			}

		}
	}
	return nil
}

// TryToFindNewestReleaseByChart will return the newest chart release given a collection of repos
func TryToFindNewestReleaseByChart(chart *release.Release, repos []*Repo) *ChartRelease {
	var newestRelease *ChartRelease
	for _, repo := range repos {
		newestInRepo := repo.NewestChartVersion(chart.Chart.Metadata)
		if newestInRepo == nil {
			continue
		}
		if newestRelease == nil {
			newestRelease = newestInRepo
		} else {
			if version.Compare(newestInRepo.Version, newestRelease.Version, ">") {
				newestRelease = newestInRepo
			}
		}
	}
	return newestRelease
}

// GetHelmReleasesVersion returns a collection of deployed helm version 3 charts in a cluster.
func (h *Helm) GetHelmReleasesVersion(helmRepos []*Repo, helmReleases []*release.Release) []output.ReleaseOutput {
	outputObjects := []output.ReleaseOutput{}

	klog.V(5).Infof("Got %d installed releases in the cluster", len(helmReleases))
	for _, chart := range helmReleases {
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
	return outputObjects
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

func checkChartsSimilarity(currentChartMeta *chart.Metadata, chartFromRepo *ChartRelease) bool {

	if currentChartMeta.Home != chartFromRepo.Home {
		return false
	}

	if currentChartMeta.Description != chartFromRepo.Description {
		return false
	}

	for _, source := range currentChartMeta.Sources {
		if !containsString(chartFromRepo.Sources, source) {
			return false
		}
	}

	chartFromRepoMaintainers := map[string]bool{}
	for _, m := range chartFromRepo.Maintainers {
		chartFromRepoMaintainers[m.Email+";"+m.Name+";"+m.URL] = true
	}
	for _, m := range currentChartMeta.Maintainers {
		if !chartFromRepoMaintainers[m.Email+";"+m.Name+";"+m.URL] {
			return false
		}
	}
	return true
}

// GetNewestReleaseByName will return the newest chart release given a collection of repos
func GetNewestReleaseByName(name string, repos []*Repo) *ChartRelease {
	newestRelease := &ChartRelease{}
	for _, repo := range repos {
		newestInRepo := repo.NewestVersion(name)
		if newestRelease == nil {
			newestRelease = newestInRepo
		} else {
			if version.Compare(newestInRepo.Version, newestRelease.Version, ">") {
				newestRelease = newestInRepo
			}
		}
	}
	return newestRelease
}

// GetChartInfo returns info about a chart with the version specified
func GetChartInfo(name string, version string, repos []*Repo) *ChartRelease {
	for _, repo := range repos {
		for key, chart := range repo.Charts.Entries {
			if key == name {
				for _, release := range chart {
					if release.Version == version {
						return &release
					}
				}
			}
		}
	}
	return nil
}

// IsRepoIncluded check if the repo is included in the list of repos
func IsRepoIncluded(chartName string, repos []*Repo) []*Repo {
	found := []*Repo{}
	for _, repo := range repos {
		if contains(chartName, repo) {
			found = append(found, repo)
		}
	}
	return found
}

func contains(chartName string, repo *Repo) bool {
	for name := range repo.Charts.Entries {
		if name == chartName {
			return true
		}
	}
	return false
}
