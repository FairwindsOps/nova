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
