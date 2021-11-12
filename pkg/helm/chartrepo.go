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
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fairwindsops/nova/pkg/output"
	version "github.com/mcuadros/go-version"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"

	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

// Repo represents a Helm chart Repo
type Repo struct {
	URL    string
	Charts *ChartReleases
}

// ChartReleases contains the chart releases of a helm repository
type ChartReleases struct {
	APIVersion string                    `yaml:"apiVersion" json:"apiVersion"`
	Entries    map[string][]ChartRelease `yaml:"entries" json:"entries"`
}

// ChartRelease is a single chart version in a helm repository
type ChartRelease struct {
	APIVersion  string             `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	AppVersion  string             `yaml:"appVersion" json:"appVersion"`
	Created     time.Time          `yaml:"created" json:"created"`
	Description string             `yaml:"description" json:"description"`
	Digest      string             `yaml:"digest,omitempty" json:"digest,omitempty"`
	Maintainers []chart.Maintainer `yaml:"maintainers,omitempty" json:"maintainers,omitempty"`
	Name        string             `yaml:"name" json:"name"`
	Urls        []string           `yaml:"urls" json:"urls"`
	Version     string             `yaml:"version" json:"version"`
	Home        string             `yaml:"home" json:"home"`
	Sources     []string           `yaml:"sources" json:"sources"`
	Keywords    []string           `yaml:"keywords" json:"keywords"`
	Icon        string             `yaml:"icon,omitempty" json:"icon,omitempty"`
	Deprecated  bool               `yaml:"deprecated" json:"deprecated"`
}

// NewRepoList returns data about a helm chart repository, given its url
func NewRepoList(urls []string) []*Repo {
	var repos []*Repo

	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, url := range urls {
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
	newestRelease := &ChartRelease{}
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

func clusterVersionExistsInPackage(clusterVersion string, pkg ArtifactHubHelmPackage) bool {
	for _, packageVersion := range pkg.AvailableVersions {
		if packageVersion.Version == clusterVersion {
			return true
		}
	}
	return false
}

func preferVerifiedAndOfficial(pkg ArtifactHubHelmPackage) int {
	ret := 0
	if pkg.Repository.VerifiedPublisher {
		klog.V(3).Infof("Preferring verified publisher for package %s: %s", pkg.Name, pkg.Repository.Name)
		ret++
	}
	if pkg.Repository.Official {
		klog.V(3).Infof("Preferring official publisher for package %s: %s", pkg.Name, pkg.Repository.Name)
		ret++
	}
	return ret
}

func TryToFindNewestReleaseByChartNew(clusterRelease *release.Release, ahubPackages []ArtifactHubHelmPackage) *output.ReleaseOutput {
	var firstMatch string
	var properPackage ArtifactHubHelmPackage
	matchingPackages := map[string]ArtifactHubHelmPackage{}
	for _, p := range ahubPackages {
		if !clusterVersionExistsInPackage(clusterRelease.Chart.Metadata.Version, p) {
			continue
		}
		if !checkChartsSimilarityNew(clusterRelease, p) {
			continue
		}
		matchingPackages[p.PackageID] = p
		if firstMatch == "" {
			firstMatch = p.PackageID
		}
	}
	if len(matchingPackages) == 0 {
		return nil
	}
	var highScore int
	var highScorePackageID string
	for _, pkg := range matchingPackages {
		score := preferVerifiedAndOfficial(pkg)
		if score > highScore {
			highScore = score
			highScorePackageID = pkg.PackageID
		}
	}
	if highScore == 0 {
		for _, p := range ahubPackages {
			if p.Name == clusterRelease.Chart.Metadata.Name {
				klog.V(3).Infof("No high scores for '%s'. Found respository %s/%s", clusterRelease.Chart.Metadata.Name, p.Repository.Name, p.Name)
			}
		}
		klog.V(3).Infof("No High Scores. Using this one: %v", matchingPackages[firstMatch].Repository.Url)
		properPackage = matchingPackages[firstMatch]
	} else {
		properPackage = matchingPackages[highScorePackageID]
	}
	return prepareOutput(clusterRelease, properPackage)
}

func prepareOutput(release *release.Release, pkg ArtifactHubHelmPackage) *output.ReleaseOutput {
	return &output.ReleaseOutput{
		ReleaseName: release.Name,
		ChartName:   release.Chart.Metadata.Name,
		Namespace:   release.Namespace,
		Description: release.Chart.Metadata.Description,
		Home:        release.Chart.Metadata.Home,
		Icon:        release.Chart.Metadata.Icon,
		Installed: output.VersionInfo{
			Version:    release.Chart.Metadata.Version,
			AppVersion: release.Chart.Metadata.AppVersion,
		},
		Latest: output.VersionInfo{
			Version:    pkg.Version,
			AppVersion: pkg.AppVersion,
		},
		IsOld:       version.Compare(release.Chart.Metadata.Version, pkg.Version, "<"),
		Deprecated:  pkg.Deprecated,
		HelmVersion: "3",
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

func checkChartsSimilarityNew(release *release.Release, pkg ArtifactHubHelmPackage) bool {
	if release.Chart.Metadata.Home != pkg.HomeURL {
		return false
	}
	if release.Chart.Metadata.Description != pkg.Description {
		return false
	}
	for _, source := range pkg.Links {
		if source.Name == "source" {
			if !containsString(release.Chart.Metadata.Sources, source.URL) {
				return false
			}
		}
	}
	pkgMaintainers := map[string]bool{}
	for _, m := range pkg.Maintainers {
		pkgMaintainers[m.Email+";"+m.Name+";"] = true
	}
	for _, m := range release.Chart.Metadata.Maintainers {
		if !pkgMaintainers[m.Email+";"+m.Name+";"] {
			return false
		}
	}
	return true
}

func containsString(arr []string, val string) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
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

// IsValidRelease returns a bool indicating whether a version string is valid or not.
func IsValidRelease(version string) bool {
	var specialForms = []string{
		"SNAPSHOT",
		"snapshot",
		"dev",
		"alpha",
		"a",
		"beta",
		"b",
		"RC",
		"rc",
		"#",
		"p",
		"pl",
	}

	for _, invalid := range specialForms {
		if strings.Contains(version, invalid) {
			return false
		}
	}
	return true
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
