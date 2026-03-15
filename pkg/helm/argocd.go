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
	"context"
	"encoding/json"
	"fmt"
	"log"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// ArgoApplicationSourceHelm is a reduced view of an ArgoCD ApplicationSourceHelm struct
type ArgoApplicationSourceHelm struct {
	// ReleaseName is the Helm release name to use. If omitted it will use the application name
	ReleaseName string `json:"releaseName,omitempty"`
	// Namespace is an optional namespace to template with. If left empty, defaults to the app's destination namespace.
	Namespace string `json:"namespace,omitempty"`
}

// ArgoApplicationSource is a reduced view of an ArgoCD ApplicationSource struct
type ArgoApplicationSource struct {
	// RepoURL is the URL to the repository (Git or Helm) that contains the application manifests
	RepoURL string `json:"repoURL"`
	// TargetRevision defines the revision of the source to sync the application to.
	// In case of Helm, this is a semver tag for the Chart's version.
	TargetRevision string `json:"targetRevision"`
	// Helm holds helm specific options
	Helm *ArgoApplicationSourceHelm `json:"helm,omitempty"`
	// Chart is a Helm chart name, and must be specified for applications sourced from a Helm repo.
	Chart string `json:"chart"`
}

// ArgoApplicationDestination is a reduced view of an ArgoCD ApplicationDestination struct
type ArgoApplicationDestination struct {
	// The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace
	Namespace string `json:"namespace,omitempty"`
}

// ArgoApplicationSpec is a reduced view of an ArgoCD ApplicationSpec struct
type ArgoApplicationSpec struct {
	// Destination is a reference to the target Kubernetes server and namespace
	Destination ArgoApplicationDestination `json:"destination"`
	// Source is a reference to the location of the application's manifests or chart
	Source *ArgoApplicationSource `json:"source,omitempty"`
	// Sources is a reference to the location of the application's manifests or chart
	Sources []ArgoApplicationSource `json:"sources,omitempty"`
}

// ArgoApplication is a reduced view of an ArgoCD Application struct
type ArgoApplication struct {
	// Metadata is Kubernetes object metadata
	Metadata metav1.ObjectMeta `json:"metadata"`
	// Spec is the object specification
	Spec ArgoApplicationSpec `json:"spec"`
}

// IsHelmChart checks if the ApplicationSource deploys an Helm chart
func (a *ArgoApplicationSource) IsHelmChart() bool {
	return a.Chart != "" || a.Helm != nil
}

// GetArgoCDApplicationReleases queries all ArgoCD applications in the `argocd` namespace
// and extracts the Helm releases deployed in the given namespace
func (h *Helm) GetArgoCDApplicationReleases(namespace string) ([]*release.Release, error) {
	// Define ArgoCD Application GVR
	applicationGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	dynamicClient := h.Kube.DynamicClient

	appsUnstructured, err := dynamicClient.Resource(applicationGVR).
		Namespace("argocd").
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatalf("Error listing applications: %v", err)
		return nil, err
	}

	// Parse each application
	apps := []ArgoApplication{}
	for _, app := range appsUnstructured.Items {
		parsed, err := parseArgoApplication(app.Object)
		if err != nil {
			log.Printf("Warning: failed to parse application %s: %v", app.GetName(), err)
			continue
		}
		apps = append(apps, *parsed)
	}

	// Extract the Helm releases for each application
	// and filter them by namespace (empty namespace is a catch-all)
	releases := []*release.Release{}
	for _, app := range apps {
		appReleases := app.ToHelmReleases()
		for _, release := range appReleases {
			if namespace == "" || release.Namespace == namespace {
				releases = append(releases, release)
			}
		}
	}

	return releases, nil
}

func parseArgoApplication(obj map[string]any) (*ArgoApplication, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}

	var app ArgoApplication
	if err := json.Unmarshal(jsonBytes, &app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into ArgoApplication: %w", err)
	}

	return &app, nil
}

// ToHelmReleases extracts all the Helm releases deployed by an ArgoCD Application
func (app *ArgoApplication) ToHelmReleases() []*release.Release {
	releases := []*release.Release{}

	// Check if using single source (spec.source)
	if app.Spec.Source != nil && app.Spec.Source.IsHelmChart() {
		rel := createHelmRelease(
			app,
			app.Spec.Source,
		)
		releases = append(releases, rel)
	}

	// Check if using multiple sources (spec.sources)
	// Multi-source apps were introduced in ArgoCD 2.6
	for _, source := range app.Spec.Sources {
		if source.IsHelmChart() {
			rel := createHelmRelease(
				app,
				&source,
			)
			releases = append(releases, rel)
		}
	}

	return releases
}

// createHelmRelease creates a Helm Release from ArgoCD source information
func createHelmRelease(app *ArgoApplication, source *ArgoApplicationSource) *release.Release {
	// Determine the release name
	var name string
	if source.Helm != nil && source.Helm.ReleaseName != "" {
		name = source.Helm.ReleaseName
	} else {
		name = app.Metadata.Name

	}
	// Determine the namespace
	var namespace string
	if source.Helm != nil && source.Helm.Namespace != "" {
		namespace = source.Helm.Namespace
	} else if app.Spec.Destination.Namespace != "" {
		namespace = app.Spec.Destination.Namespace
	} else {
		namespace = "default"
	}

	return &release.Release{
		Name:      name,
		Namespace: namespace,
		Version:   1, // ArgoCD doesn't track revision numbers like Helm does
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    source.Chart,
				Version: source.TargetRevision,
				// The original source is not available from ArgoCD Application spec,
				// but this is the best we can do
				Sources: []string{
					source.RepoURL,
				},
				// AppVersion is not available from ArgoCD Application spec
				// It would need to be fetched from the actual chart
				AppVersion:  "",
				KubeVersion: "",
				Description: "",
			},
		},
	}
}
