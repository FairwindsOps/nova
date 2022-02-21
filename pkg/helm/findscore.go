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

	"github.com/fairwindsops/nova/pkg/output"
	version "github.com/mcuadros/go-version"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/klog/v2"
)

// FindBestArtifactHubMatch takes the helm releases found in the cluster and attempts to match those to a package in artifacthub
func FindBestArtifactHubMatch(clusterRelease *release.Release, ahubPackages []ArtifactHubHelmPackage) *output.ReleaseOutput {
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
	klog.V(10).Infof("highScore for '%s': %d, highScorePackage Repo: %s", clusterRelease.Chart.Metadata.Name, highScore, highScorePackage.Repository.Name)
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
		"ingress-nginx",
		"cert-manager",
	}
	if release.Chart.Metadata.Home == pkg.HomeURL {
		klog.V(10).Infof("+1 score for %s Home URL (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if release.Chart.Metadata.Description == pkg.Description {
		klog.V(10).Infof("+1 score for %s Description (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	for _, source := range pkg.Links {
		if source.Name == "source" {
			if containsString(release.Chart.Metadata.Sources, source.URL) {
				klog.V(10).Infof("+1 score for %s source links (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
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
		klog.V(10).Infof("+1 score for %s Maintainers (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if pkg.Repository.VerifiedPublisher {
		klog.V(10).Infof("+1 score for %s verified publisher (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if clusterVersionExistsInPackage(release.Chart.Metadata.Version, pkg) {
		klog.V(10).Infof("+1 score for %s, current version exists in available versions (ahub package %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if containsString(preferredRepositories, pkg.Repository.Name) {
		klog.V(10).Infof("+1 score for %s, preffered repo (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	klog.V(10).Infof("calculated score repo: %s, release: %s, score: %d\n\n", pkg.Repository.Name, release.Name, ret)
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
