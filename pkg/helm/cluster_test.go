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
	"helm.sh/helm/v3/pkg/release"
)

func TestHelm_OverrideDesiredVersion(t *testing.T) {
	tests := []struct {
		name            string
		desiredVersions []DesiredVersion
		input           *output.ReleaseOutput
		want            *output.ReleaseOutput
	}{
		{
			name:            "Empty desiredVersions",
			desiredVersions: []DesiredVersion{},
			input: &output.ReleaseOutput{
				ChartName: "test-chart",
				Latest: output.VersionInfo{
					Version:    "0.0.0",
					AppVersion: "0.0.0",
				},
			},
			want: &output.ReleaseOutput{
				ChartName: "test-chart",
				Latest: output.VersionInfo{
					Version:    "0.0.0",
					AppVersion: "0.0.0",
				},
			},
		},
		{
			name: "Override desiredVersions",
			desiredVersions: []DesiredVersion{
				{
					Name:    "test-chart",
					Version: "1.1.1",
				},
			},
			input: &output.ReleaseOutput{
				ChartName: "test-chart",
				Latest: output.VersionInfo{
					Version:    "0.0.0",
					AppVersion: "0.0.0",
				},
			},
			want: &output.ReleaseOutput{
				ChartName: "test-chart",
				Latest: output.VersionInfo{
					Version:    "1.1.1",
					AppVersion: "",
				},
				IsOld:      true,
				Overridden: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Helm{
				DesiredVersions: tt.desiredVersions,
			}
			h.OverrideDesiredVersion(tt.input)
			assert.EqualValues(t, tt.want, tt.input)
		})
	}
}

func TestHelm_filterIgnoredReleases(t *testing.T) {
	tests := []struct {
		name              string             // Name of test case
		releaseIgnoreList []string           // List of release names to be ignored
		chartIgnoreList   []string           // List of charts to be ignored
		input             []*release.Release // Input to filtering function
		want              []*release.Release // Output from filtering function
	}{
		{
			name:              "EmptyInput",
			releaseIgnoreList: []string{},
			chartIgnoreList:   []string{},
			input:             []*release.Release{},
			want:              []*release.Release{},
		},
		{
			name:              "NoIgnoredReleasesOrCharts",
			releaseIgnoreList: []string{},
			chartIgnoreList:   []string{},
			input: []*release.Release{
				{
					Name: "foo",
				},
			},
			want: []*release.Release{
				{
					Name: "foo",
				},
			},
		},
		{
			name: "AllIgnoredReleases",
			releaseIgnoreList: []string{
				"foo",
			},
			chartIgnoreList: []string{},
			input: []*release.Release{
				{
					Name: "foo",
				},
			},
			want: []*release.Release{},
		},
		{
			name: "SomeIgnoredReleases",
			releaseIgnoreList: []string{
				"foo",
			},
			chartIgnoreList: []string{},
			input: []*release.Release{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
			want: []*release.Release{{
				Name: "bar",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := filterIgnoredReleases(tt.input, tt.releaseIgnoreList, tt.chartIgnoreList)
			if len(output) != len(tt.want) {
				t.Fatalf("filtering did not catch all cases, expected %d releases, instead got %d", len(tt.want), len(output))
			}
		})
	}
}
