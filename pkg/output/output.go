package output

import "fmt"

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

// ToMarkdownTable returns a markdown formatted table
func (output *ReleaseOutput) ToMarkdownTable() string {
	if output.AppVersion != nil && output.NewestAppVersion != nil {
		txt := "| | Old | New |\n|-|-|-|\n| Version | %s | %s |\n| AppVersion | %s | %s |"
		return fmt.Sprintf(txt, output.Version, output.NewestVersion, *output.AppVersion, *output.NewestAppVersion)
	}
	return ""
}
