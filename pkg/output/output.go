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
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/tabwriter"

	"github.com/fairwindsops/nova/pkg/containers"
	"helm.sh/helm/v3/pkg/release"

	"k8s.io/klog/v2"
)

const (
	// JSONFormat json output format
	JSONFormat = "json"
	// TableFormat table/csv output format
	TableFormat = "table"
)

// Output is the object that Nova outputs
type Output struct {
	HelmReleases []ReleaseOutput `json:"helm"`
	IncludeAll   bool            `json:"include_all"`
}

// ContainersOutput represents the output data we need for displaying a table of out of date container images
type ContainersOutput struct {
	ContainerImages   []ContainerOutput          `json:"container_images"`
	ErrImages         []*containers.ErroredImage `json:"err_images"`
	IncludeAll        bool                       `json:"include_all"`
	LatestStringFound bool                       `json:"latest_string_found"`
}

// ContainersOutput represents the output data we need for displaying a table of out of date container images
type HelmAndContainersOutput struct {
	Helm      Output
	Container ContainersOutput
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

// ContainerOutput represents all the data we need for a single container image
type ContainerOutput struct {
	Name               string `json:"name"`
	CurrentVersion     string `json:"current_version"`
	LatestVersion      string `json:"latest_version"`
	LatestMinorVersion string `json:"latest_minor_version"`
	LatestPatchVersion string `json:"latest_patch_version"`
	IsOld              bool   `json:"outdated"`
}

// VersionInfo contains both a chart version and an app version
type VersionInfo struct {
	Version    string `json:"version"`
	AppVersion string `json:"appVersion"`
}

// NewOutputWithHelmReleases creates a new output object with the given helm releases pre-populated with the installed version
func NewOutputWithHelmReleases(helmReleases []*release.Release) Output {
	var output Output
	for _, helmRelease := range helmReleases {
		var release ReleaseOutput
		release.ChartName = helmRelease.Chart.Metadata.Name
		release.ReleaseName = helmRelease.Name
		release.Namespace = helmRelease.Namespace
		release.Description = helmRelease.Chart.Metadata.Description
		release.Home = helmRelease.Chart.Metadata.Home
		release.Icon = helmRelease.Chart.Metadata.Icon
		release.Installed = VersionInfo{helmRelease.Chart.Metadata.Version, helmRelease.Chart.Metadata.AppVersion}
		output.HelmReleases = append(output.HelmReleases, release)
	}
	return output
}

// ToFile dispatches a message to file
func (output Output) ToFile(filename string) error {
	output.dedupe()
	extension := path.Ext(filename)
	switch extension {
	case ".json":
		data, err := json.Marshal(output)
		if err != nil {
			klog.Errorf("Error marshaling json: %v", err)
			return err
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			klog.Errorf("Error writing to file %s: %v", filename, err)
		}
	case ".csv":
		file, err := os.Create(filename)
		defer file.Close()
		if err != nil {
			return err
		}
		w := csv.NewWriter(file)
		defer w.Flush()
		header := []string{"Release Name", "Chart Name", "Namespace", "HelmVersion", "Installed", "Latest", "Old", "Deprecated"}
		var data [][]string
		data = append(data, header)
		for _, rl := range output.HelmReleases {
			row := []string{rl.ReleaseName, rl.ChartName, rl.Namespace, rl.HelmVersion, rl.Installed.Version, rl.Latest.Version, strconv.FormatBool(rl.IsOld), strconv.FormatBool(rl.Deprecated)}
			data = append(data, row)
		}
		w.WriteAll(data)
	default:
		return errors.New("File format is not supported. The supported file format are json and csv only")
	}
	return nil
}

// Print sends the output to STDOUT
func (output Output) Print(format string, wide, showOld bool) {
	if len(output.HelmReleases) == 0 {
		fmt.Println("No releases found")
		return
	}
	output.dedupe()
	switch format {
	case JSONFormat:
		data, _ := json.Marshal(output.HelmReleases)
		fmt.Fprintln(os.Stdout, string(data))
	case TableFormat:
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
			if (!output.IncludeAll && release.Latest.Version == "") || (showOld && !release.IsOld) {
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
	default:
		klog.Errorf("Output format is not supported. The supported formats are json and table only")
	}
}

// dedupe will remove duplicate releases from the output if both artifacthub and a custom URL to a helm repository find matches.
// this will always override any found by artifacthub with the version from a custom helm repo url because those are found last and
// will therefore always be at the end of the output.HelmReleases array.
func (output *Output) dedupe() {
	var unique []ReleaseOutput
	type key struct{ releaseName, chartName, namespace string }
	tracker := make(map[key]int)
	for _, release := range output.HelmReleases {
		k := key{release.ReleaseName, release.ChartName, release.Namespace}
		if i, ok := tracker[k]; ok {
			klog.V(8).Infof("found duplicate release output, deduping: '%s', chart: '%s', namespace: '%s'", release.ReleaseName, release.ChartName, release.Namespace)
			unique[i] = release
		} else {
			tracker[k] = len(unique)
			unique = append(unique, release)
		}
	}
	output.HelmReleases = unique
}

// NewContainersOutput creates a new ContainersOutput object ready to be printed
func NewContainersOutput(containers []*containers.Image, errImages []*containers.ErroredImage, showNonSemver, showErrored, includeAll bool) *ContainersOutput {
	var output ContainersOutput
	output.IncludeAll = includeAll
	for _, container := range containers {
		if container == nil {
			continue
		}
		if !showNonSemver && !container.StrictSemver {
			continue
		}
		var containerOutput ContainerOutput
		prefix := container.Prefix
		containerOutput.Name = container.Name
		containerOutput.CurrentVersion = prefix + container.Current.Value
		containerOutput.LatestVersion = prefix + container.Current.Value
		containerOutput.LatestMinorVersion = prefix + container.Current.Value
		containerOutput.LatestPatchVersion = prefix + container.Current.Value
		containerOutput.IsOld = false
		if container.Newest != nil {
			containerOutput.LatestVersion = prefix + container.Newest.Value
			containerOutput.IsOld = true
		}
		if container.NewestMinor != nil {
			containerOutput.LatestMinorVersion = prefix + container.NewestMinor.Value
		}
		if container.NewestPatch != nil {
			containerOutput.LatestPatchVersion = prefix + container.NewestPatch.Value
		}
		if containerOutput.CurrentVersion == "latest" {
			output.LatestStringFound = true
		}
		output.ContainerImages = append(output.ContainerImages, containerOutput)
	}
	if showErrored {
		output.ErrImages = errImages
	}
	return &output
}

// Print prints the ContainersOutput to STDOUT
func (output ContainersOutput) Print(format string) {
	if len(output.ContainerImages) == 0 && len(output.ErrImages) == 0 {
		fmt.Println("No images found")
		return
	}
	switch format {
	case JSONFormat:
		data, _ := json.Marshal(output)
		fmt.Fprintln(os.Stdout, string(data))
	case TableFormat:
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
		if len(output.ContainerImages) != 0 {
			header := "Container Name\tCurrent Version\tOld\tLatest\tLatest Minor\tLatest Patch"
			fmt.Fprintln(w, header)
			separator := "==============\t===============\t===\t======\t=============\t============="
			fmt.Fprintln(w, separator)

			for _, c := range output.ContainerImages {
				if !output.IncludeAll && c.LatestVersion == c.CurrentVersion {
					continue
				}
				line := c.Name + "\t"
				line += c.CurrentVersion + "\t"
				line += fmt.Sprintf("%t", c.IsOld) + "\t"
				line += c.LatestVersion + "\t"
				line += c.LatestMinorVersion + "\t"
				line += c.LatestPatchVersion + "\t"
				fmt.Fprintln(w, line)
			}
		}

		if len(output.ErrImages) == 0 {
			w.Flush()
			return
		}
		fmt.Fprintln(w, "\n\nErrors:")
		errHeader := "Container Name\tError"
		fmt.Fprintln(w, errHeader)
		errSeparator := "==============\t====="
		fmt.Fprintln(w, errSeparator)
		for _, e := range output.ErrImages {
			line := e.Image + "\t"
			line += e.Err + "\t"
			fmt.Fprintln(w, line)
		}
		w.Flush()
		if output.LatestStringFound {
			fmt.Printf("Found a container utilizing the 'latest' tag. This is bad practice and should be avoided.\n\n")
		}
	default:
		klog.Errorf("Output format is not supported. The supported formats are json and table only")
	}
}

// CombinedOutputFormat has both helm releases and containers info in a backwards compatible way
type CombinedOutputFormat struct {
	Helm       []ReleaseOutput `json:"helm"`
	IncludeAll bool            `json:"include_all"`
	Container  struct {
		ContainersOutput
	} `json:"container"`
}

// NewHelmAndContainersOutput creates a new HelmAndContainersOutput object ready to be printed
func NewHelmAndContainersOutput(helm Output, container ContainersOutput) *HelmAndContainersOutput {
	return &HelmAndContainersOutput{
		Helm:      helm,
		Container: container,
	}
}

// Print prints the HelmAndContainersOutput to STDOUT
func (output HelmAndContainersOutput) Print(format string, wide, showOld bool) {
	switch format {
	case TableFormat:
		output.Helm.Print(format, wide, showOld)
		fmt.Println("")
		output.Container.Print(format)
	case JSONFormat:
		outputFormat := CombinedOutputFormat{
			Helm:       output.Helm.HelmReleases,
			Container:  struct{ ContainersOutput }{ContainersOutput: output.Container},
			IncludeAll: output.Helm.IncludeAll,
		}
		data, _ := json.Marshal(outputFormat)
		fmt.Fprintln(os.Stdout, string(data))
	}
}

// Print prints the HelmAndContainersOutput to STDOUT
func (output HelmAndContainersOutput) ToFile(filename string) error {
	output.Helm.dedupe()
	extension := path.Ext(filename)
	switch extension {
	case ".json":
		outputFormat := CombinedOutputFormat{
			Helm:       output.Helm.HelmReleases,
			Container:  struct{ ContainersOutput }{ContainersOutput: output.Container},
			IncludeAll: output.Helm.IncludeAll,
		}
		data, err := json.Marshal(outputFormat)
		if err != nil {
			klog.Errorf("Error marshaling json: %v", err)
			return err
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			klog.Errorf("Error writing to file %s: %v", filename, err)
		}
	default:
		// TODO - when both flags are used should it have CSV output?!
		return errors.New("File format is not supported. The supported file format is json only")
	}
	return nil
}
