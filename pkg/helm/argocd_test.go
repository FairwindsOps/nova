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

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHelm_ArgoApplicationToHelmReleases(t *testing.T) {
	tests := []struct {
		name  string             // Name of test case
		input ArgoApplication    // Input ArgoCD Application
		want  []*release.Release // Helm releases extracted from the application
	}{
		{
			name: "EmptyInput",
			input: ArgoApplication{
				Metadata: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "argocd",
				},
				Spec: ArgoApplicationSpec{
					Destination: ArgoApplicationDestination{
						Namespace: "space",
					},
				},
			},
			want: []*release.Release{},
		},
		{
			name: "NotHelmChart",
			input: ArgoApplication{
				Metadata: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "argocd",
				},
				Spec: ArgoApplicationSpec{
					Destination: ArgoApplicationDestination{
						Namespace: "space",
					},
					Source: &ArgoApplicationSource{
						RepoURL:        "https://example.com/",
						TargetRevision: "1.0.0",
					},
				},
			},
			want: []*release.Release{},
		},
		{
			name: "SimpleChart",
			input: ArgoApplication{
				Metadata: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "argocd",
				},
				Spec: ArgoApplicationSpec{
					Destination: ArgoApplicationDestination{
						Namespace: "space",
					},
					Source: &ArgoApplicationSource{
						RepoURL:        "https://example.com/charts",
						TargetRevision: "1.0.0",
						Chart:          "test",
					},
				},
			},
			want: []*release.Release{
				{
					Name:      "foo",
					Namespace: "space",
					Version:   1,
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name:    "test",
							Version: "1.0.0",
							Sources: []string{
								"https://example.com/charts",
							},
							AppVersion:  "",
							KubeVersion: "",
							Description: "",
						},
					},
				},
			},
		},
		{
			name: "MultipleSources",
			input: ArgoApplication{
				Metadata: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "argocd",
				},
				Spec: ArgoApplicationSpec{
					Destination: ArgoApplicationDestination{
						Namespace: "space",
					},
					Sources: []ArgoApplicationSource{
						{
							RepoURL:        "https://example1.com/charts",
							TargetRevision: "1.0.0",
							Chart:          "test1",
						},
						{
							RepoURL:        "https://example2.com/charts",
							TargetRevision: "1.0.0",
							Chart:          "test2",
						},
					},
				},
			},
			want: []*release.Release{
				{
					Name:      "foo",
					Namespace: "space",
					Version:   1,
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name:    "test1",
							Version: "1.0.0",
							Sources: []string{
								"https://example1.com/charts",
							},
							AppVersion:  "",
							KubeVersion: "",
							Description: "",
						},
					},
				},
				{
					Name:      "foo",
					Namespace: "space",
					Version:   1,
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name:    "test2",
							Version: "1.0.0",
							Sources: []string{
								"https://example2.com/charts",
							},
							AppVersion:  "",
							KubeVersion: "",
							Description: "",
						},
					},
				},
			},
		},
		{
			name: "OverrideReleaseAndNamespace",
			input: ArgoApplication{
				Metadata: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "argocd",
				},
				Spec: ArgoApplicationSpec{
					Destination: ArgoApplicationDestination{
						Namespace: "space",
					},
					Source: &ArgoApplicationSource{
						RepoURL:        "https://example.com/charts",
						TargetRevision: "1.0.0",
						Chart:          "test",
						Helm: &ArgoApplicationSourceHelm{
							ReleaseName: "bar",
							Namespace:   "baz",
						},
					},
				},
			},
			want: []*release.Release{
				{
					Name:      "bar",
					Namespace: "baz",
					Version:   1,
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name:    "test",
							Version: "1.0.0",
							Sources: []string{
								"https://example.com/charts",
							},
							AppVersion:  "",
							KubeVersion: "",
							Description: "",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.input.ToHelmReleases()
			if len(output) != len(tt.want) {
				t.Fatalf("did not convert all argocd helm releases, expected %d releases, instead got %d", len(tt.want), len(output))
			}
			for i, want := range tt.want {
				assert.EqualExportedValues(t, want, output[i])
			}
		})
	}
}
