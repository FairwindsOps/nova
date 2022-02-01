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
	"testing"

	"github.com/fairwindsops/nova/pkg/output"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

func TestIsValidRelease(t *testing.T) {
	assert.Equal(t, IsValidRelease("v1.0"), true)
	assert.Equal(t, IsValidRelease("1.0-rc3"), false)
}

func Test_containsString(t *testing.T) {
	assert.Equal(t, containsString([]string{"test", "other"}, "test"), true)
	assert.Equal(t, containsString([]string{"other"}, "test"), false)
}

var emptyRelease = release.Release{
	Chart: &chart.Chart{
		Metadata: &chart.Metadata{},
	},
}

var ahubPackage = ArtifactHubHelmPackage{
	Name:        "test",
	Description: "This is a chart.",
	HomeURL:     "https://example.com/charts",
	Links: []Link{
		{
			Name: "source",
			URL:  "https://example.com/charts",
		},
	},
	Maintainers: []Maintainer{
		{
			Email: "me@example.com",
			Name:  "John",
		},
	},
	AvailableVersions: []AvailableVersion{
		{
			Version: "1.0.0",
		},
		{
			Version: "1.0.1",
		},
	},
	Version:    "1.0.1",
	AppVersion: "1.0.1",
	Deprecated: false,
	Repository: ArtifactHubRepository{
		Name:              "fairwinds-stable",
		VerifiedPublisher: true,
	},
}

var helmRelease = &release.Release{
	Name:      "test",
	Namespace: "test",
	Chart: &chart.Chart{
		Metadata: &chart.Metadata{
			Name:        "test",
			Version:     "1.0.0",
			AppVersion:  "1.0.0",
			Home:        "https://example.com/charts",
			Description: "This is a chart.",
			Icon:        "https://example.com/charts/icon.png",
			Sources: []string{
				"https://example.com/charts",
			},
			Maintainers: []*chart.Maintainer{
				{
					Email: "me@example.com",
					Name:  "John",
				},
			},
		},
	},
}

func Test_scoreChartSimilarity(t *testing.T) {
	tests := []struct {
		name    string
		release *release.Release
		pkg     ArtifactHubHelmPackage
		want    int
	}{
		{
			name:    "highest score",
			release: helmRelease,
			pkg:     ahubPackage,
			want:    7,
		},
		{
			name:    "empty pkg struct",
			release: &emptyRelease,
			pkg: ArtifactHubHelmPackage{
				Description: "This is a chart.",
				HomeURL:     "https://example.com/charts",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreChartSimilarity(tt.release, tt.pkg)
			if got != tt.want {
				t.Errorf("scoreChartSimilarity() got =%v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepareOutput(t *testing.T) {
	tests := []struct {
		name    string
		release *release.Release
		pkg     ArtifactHubHelmPackage
		want    *output.ReleaseOutput
	}{
		{
			name:    "proper output",
			release: helmRelease,
			pkg:     ahubPackage,
			want: &output.ReleaseOutput{
				ReleaseName: "test",
				ChartName:   "test",
				Namespace:   "test",
				Description: "This is a chart.",
				Home:        "https://example.com/charts",
				Icon:        "https://example.com/charts/icon.png",
				Installed: output.VersionInfo{
					Version:    "1.0.0",
					AppVersion: "1.0.0",
				},
				Latest: output.VersionInfo{
					Version:    "1.0.1",
					AppVersion: "1.0.1",
				},
				IsOld:       true,
				Deprecated:  false,
				HelmVersion: "3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareOutput(tt.release, tt.pkg)
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("prepareOutput() got: %v, want: %v", got, tt.want)
			}
		})
	}
}
