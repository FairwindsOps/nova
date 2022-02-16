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

package output

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"k8s.io/klog"
)

// Output is the object that Nova outputs
type Output struct {
	HelmReleases []ReleaseOutput `json:"helm"`
	IncludeAll   bool            `json:"include_all"`
}

// ReleaseOutput represents a release
type ReleaseOutput struct {
	ReleaseName string `json:"release"`
	ChartName   string `json:"chartName"`
	Namespace   string `json:"namespace,omitempty"`
	Description string `json:"description"`
	Home        string `json:"home,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Installed   VersionInfo
	Latest      VersionInfo
	IsOld       bool   `json:"outdated"`
	Deprecated  bool   `json:"deprecated"`
	HelmVersion string `json:"helmVersion"`
	Overridden  bool   `json:"overridden"`
}

// VersionInfo contains both a chart version and an app version
type VersionInfo struct {
	Version    string `json:"version"`
	AppVersion string `json:"appVersion"`
}

// ToFile dispatches a message to file
func (output Output) ToFile(filename string) error {
	output.dedupe()
	data, err := json.Marshal(output)
	if err != nil {
		klog.Errorf("Error marshaling json: %v", err)
		return err
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		klog.Errorf("Error writing to file %s: %v", filename, err)
	}
	return nil
}

// Print sends the output to STDOUT
func (output Output) Print(wide bool) {
	if len(output.HelmReleases) == 0 {
		fmt.Println("No releases found")
	}
	output.dedupe()
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	header := "Release Name\t"
	if wide {
		header += "Chart Name\tNamespace\tHelmVersion\t"
	}
	header += "Installed\tLatest\tOld\tDeprecated"
	fmt.Fprintln(w, header)
	separator := "============\t"
	if wide {
		separator += "==========\t=========\t===========\t"
	}
	separator += "=========\t======\t===\t=========="
	fmt.Fprintln(w, separator)

	for _, release := range output.HelmReleases {
		if !output.IncludeAll && release.Latest.Version == "" {
			continue
		}
		line := release.ReleaseName + "\t"
		if wide {
			line += release.ChartName + "\t"
			line += release.Namespace + "\t"
			line += release.HelmVersion + "\t"
		}
		line += release.Installed.Version + "\t"
		line += release.Latest.Version + "\t"
		line += fmt.Sprintf("%t", release.IsOld) + "\t"
		line += fmt.Sprintf("%t", release.Deprecated) + "\t"
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

// dedupe will remove duplicate releases from the output if both artifacthub and a custom URL to a helm repository find matches.
// this will always overrite any found by artifacthub with the version from a custom helm repo url because those are found last and
// will therefore always be at the end of the output.HelmReleases array.
func (output *Output) dedupe() {
	klog.V(8).Infof("deduplicating releases for output")
	var unique []ReleaseOutput
	type key struct{ releaseName, chartName, namespace string }
	tracker := make(map[key]int)
	for _, release := range output.HelmReleases {
		k := key{release.ReleaseName, release.ChartName, release.Namespace}
		if i, ok := tracker[k]; ok {
			klog.V(8).Infof("found duplicate release: '%s', chart: '%s', namespace: '%s'", release.ReleaseName, release.ChartName, release.Namespace)
			unique[i] = release
		} else {
			tracker[k] = len(unique)
			unique = append(unique, release)
		}
	}
	output.HelmReleases = unique
}
