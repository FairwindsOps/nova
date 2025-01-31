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

const useStarCountThreshold = 10

type packageKey struct {
	Name       string
	Repository string
}

// FindBestArtifactHubMatch takes the helm releases found in the cluster and attempts to match those to a package in artifacthub
func FindBestArtifactHubMatch(clusterRelease *release.Release, ahubPackages []ArtifactHubHelmPackage) *output.ReleaseOutput {
	packagesByName := map[packageKey]ArtifactHubHelmPackage{}
	packageScores := map[packageKey]float32{}
	packageStars := map[packageKey]int{}
	var useStars bool
	for _, p := range ahubPackages {
		if p.Name != clusterRelease.Chart.Metadata.Name {
			continue
		}

		key := packageKey{Name: p.Name, Repository: p.Repository.Name}
		packageScores[key] = scoreChartSimilarity(clusterRelease, p)
		packagesByName[key] = p
		packageStars[key] = p.Stars

		if p.Stars >= useStarCountThreshold {
			useStars = true // If any package has more than 10 stars, we add a point to the highest star package
		}
	}

	var highestStarPackageName *packageKey
	var highStars int
	for p, stars := range packageStars {
		if stars > highStars {
			highStars = stars
			highestStarPackageName = &p
		}
	}

	var highScore float32
	var highScorePackage ArtifactHubHelmPackage
	for k, score := range packageScores {
		if useStars && highestStarPackageName != nil && k == *highestStarPackageName {
			klog.V(10).Infof("adding a point to the highest star package: %s:%s", k.Repository, k.Name)
			score++ // Add a point to the highest star package
		}
		if score > highScore {
			highScore = score
			highScorePackage = packagesByName[k]
		}
	}
	klog.V(10).Infof("highScore for '%s': %f, highScorePackage Repo: %s", clusterRelease.Chart.Metadata.Name, highScore, highScorePackage.Repository.Name)
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
			Version:     release.Chart.Metadata.Version,
			AppVersion:  release.Chart.Metadata.AppVersion,
			KubeVersion: release.Chart.Metadata.KubeVersion,
		},
		Latest: output.VersionInfo{
			Version:     pkg.Version,
			AppVersion:  pkg.AppVersion,
			KubeVersion: pkg.KubeVersion,
		},
		IsOld:       version.Compare(release.Chart.Metadata.Version, pkg.Version, "<"),
		Deprecated:  pkg.Deprecated,
		HelmVersion: "3",
	}
}

var preferredRepositories = []string{"bitnami", "fairwinds-stable", "fairwinds-incubator", "ingress-nginx", "cert-manager", "projectcalico",
	"grafana", "prometheus-community", "elastic", "hashicorp", "argo", "metrics-server", "gitlab", "jenkins", "harbor", "minio", "cluster-autoscaler",
	"aws-ebs-csi-driver", "coredns", "datadog", "deliveryhero", "falcosecurity", "kedacore", "kured", "oauth2-proxy", "rimusz"}

func scoreChartSimilarity(release *release.Release, pkg ArtifactHubHelmPackage) float32 {
	var ret float32
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
		pkgMaintainers[m.Name] = true
	}
	matchedMaintainers := 0
	for _, m := range release.Chart.Metadata.Maintainers {
		if pkgMaintainers[m.Name] {
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
	if pkg.Official {
		klog.V(10).Infof("+1 score for %s verified publisher (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if pkg.Repository.Official {
		klog.V(10).Infof("+1 score for %s verified publisher (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if clusterVersionExistsInPackage(release.Chart.Metadata.Version, pkg) {
		klog.V(10).Infof("+1 score for %s, current version exists in available versions (ahub package %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret++
	}
	if containsString(preferredRepositories, pkg.Repository.Name) {
		klog.V(10).Infof("+1.5 score for %s, preferred repo (ahub package repo %s)", release.Chart.Metadata.Name, pkg.Repository.Name)
		ret += 1.5
	}
	klog.V(10).Infof("calculated score repo: %s, release: %s, stars: %d, score: %f\n\n", pkg.Repository.Name, release.Name, pkg.Stars, ret)
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
		"weekly",
		"daily",
	}

	for _, invalid := range specialForms {
		if strings.Contains(version, invalid) {
			return false
		}
	}
	return true
}
