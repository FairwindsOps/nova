package output

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"k8s.io/klog"
)

const separator = " "

// ReleaseOutput represents a release
type ReleaseOutput struct {
	ReleaseName      string  `json:"release"`
	ChartName        string  `json:"chartName"`
	Namespace        string  `json:"namespace,omitempty"`
	Description      string  `json:"description"`
	Deprecated       bool    `json:"deprecated,omitempty"`
	Home             string  `json:"home,omitempty"`
	Icon             string  `json:"icon,omitempty"`
	Version          string  `json:"version"`
	AppVersion       *string `json:"appVersion,omitempty"`
	NewestVersion    string  `json:"newest"`
	NewestAppVersion *string `json:"newest_appVersion,omitempty"`
	IsOld            bool    `json:"outdated"`
}

type field struct {
	name   string
	length int
	isBool bool
}

var fieldOrder = []field{
	field{"ReleaseName", 25, false},
	field{"ChartName", 25, false},
	field{"Namespace", 20, false},
	field{"Version", 13, false},
	field{"NewestVersion", 14, false},
	field{"IsOld", 8, true},
	field{"Deprecated", 10, true},
}

// Output is the object that Nova outputs
type Output struct {
	HelmReleases []ReleaseOutput `json:"helm_releases"`
}

func (output ReleaseOutput) String() string {
	v := reflect.ValueOf(output)
	values := make([]string, len(fieldOrder))
	for idx, field := range fieldOrder {
		value := v.FieldByName(field.name)
		if field.isBool {
			boolValue := value.Bool()
			if boolValue {
				values[idx] = "True"
			} else {
				values[idx] = " "
			}
		} else {
			values[idx] = value.String()
		}
		for len(values[idx]) < field.length {
			values[idx] += " "
		}
		if len(values[idx]) > field.length {
			values[idx] = values[idx][:field.length-1] + "…"
		}
	}
	return strings.Join(values, separator)
}

// ToMarkdownTable returns a markdown formatted table
func (output *ReleaseOutput) ToMarkdownTable() string {
	if output.AppVersion != nil && output.NewestAppVersion != nil {
		txt := "| | Old | New |\n|-|-|-|\n| Version | %s | %s |\n| AppVersion | %s | %s |"
		return fmt.Sprintf(txt, output.Version, output.NewestVersion, *output.AppVersion, *output.NewestAppVersion)
	}
	return ""
}

// ToFile dispatches a message to file
func (output Output) ToFile(filename string) error {
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

func (output Output) String() string {
	if len(output.HelmReleases) == 0 {
		return "No releases found"
	}
	fieldNames := make([]string, len(fieldOrder))
	for idx, field := range fieldOrder {
		fieldNames[idx] = field.name
		for len(fieldNames[idx]) < field.length {
			fieldNames[idx] += " "
		}
	}
	str := strings.Join(fieldNames, separator)
	releaseStrings := make([]string, len(output.HelmReleases))
	for idx, release := range output.HelmReleases {
		releaseStrings[idx] = release.String()
	}
	return str + "\n" + strings.Join(releaseStrings, "\n")
}
