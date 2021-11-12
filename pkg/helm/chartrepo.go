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
	"strings"
	"time"

	"github.com/fairwindsops/nova/pkg/output"
	version "github.com/mcuadros/go-version"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"

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

func TryToFindNewestReleaseByChart(clusterRelease *release.Release, ahubPackages []ArtifactHubHelmPackage) *output.ReleaseOutput {
	var highScore int
	var highScorePackage ArtifactHubHelmPackage
	for _, p := range ahubPackages {
		score := 0
		if p.Name != clusterRelease.Chart.Metadata.Name {
			continue
		}
		score = scoreChartSimilarity(clusterRelease, p)
		if score > highScore {
			highScore = score
			highScorePackage = p
		}
	}
	// klog.V(5).Infof("Found high schore: %d for installed release '%s':\n%v", highScore, clusterRelease.Name, spew.Sdump(highScorePackage))
	if highScore == 0 {
		for _, p := range ahubPackages {
			if p.Name == clusterRelease.Chart.Metadata.Name {
				klog.V(3).Infof("No scores above 0 for '%s'. Found respository %s/%s", clusterRelease.Chart.Metadata.Name, p.Repository.Name, p.Name)
			}
		}
		klog.V(3).Infof("No scores above 0. Using this one: %v", highScorePackage.Repository.Url)
	}
	return prepareOutput(clusterRelease, highScorePackage)
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

func scoreChartSimilarity(release *release.Release, pkg ArtifactHubHelmPackage) int {
	ret := 0
	var preferredRepositories = []string{
		"bitnami",
		"fairwinds-stable",
		"fairwinds-incubator",
	}
	if release.Chart.Metadata.Home == pkg.HomeURL {
		klog.V(8).Infof("+1 score for %s Home URL (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	for _, source := range pkg.Links {
		if source.Name == "source" {
			if containsString(release.Chart.Metadata.Sources, source.URL) {
				klog.V(8).Infof("+1 score for %s source links (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
				ret++
			}
		}
	}
	pkgMaintainers := map[string]bool{}
	for _, m := range pkg.Maintainers {
		pkgMaintainers[m.Email+";"+m.Name+";"] = true
	}
	matchedMaintainers := 0
	for _, m := range release.Chart.Metadata.Maintainers {
		if pkgMaintainers[m.Email+";"+m.Name+";"] {
			matchedMaintainers++
		}
	}
	if matchedMaintainers > 0 {
		klog.V(8).Infof("+1 score for %s Maintainers (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if pkg.Repository.VerifiedPublisher {
		klog.V(8).Infof("+1 score for %s verified publisher (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if clusterVersionExistsInPackage(release.Chart.Metadata.Version, pkg) {
		klog.V(8).Infof("+1 score for %s, current version exists in available versions (ahub package %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if containsString(preferredRepositories, pkg.Repository.Name) {
		klog.V(8).Infof("+1 score for %s, preffered repo (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	klog.V(8).Infof("Calculated score repo: %s, release: %s, score: %d\n\n", release.Name, pkg.Repository.Name, ret)
	return ret
}

func clusterVersionExistsInPackage(clusterVersion string, pkg ArtifactHubHelmPackage) bool {
	for _, packageVersion := range pkg.AvailableVersions {
		if packageVersion.Version == clusterVersion {
			return true
		}
	}
	return false
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
