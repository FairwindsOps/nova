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

package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"

	nova_helm "github.com/fairwindsops/nova/pkg/helm"
	"github.com/fairwindsops/nova/pkg/output"
)

var (
	helmVersion      string
	outputFile       string
	repos            []string
	pollHelmHub      bool
	helmHubConfigURL string
	wide             bool
	kubeContext      string
)

func init() {
	rootCmd.AddCommand(clusterCmd)
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output-file", "", "", "Path on local filesystem to write file output to")
	rootCmd.PersistentFlags().StringVar(&kubeContext, "context", "", "A context to use in the kubeconfig.")
	clusterCmd.PersistentFlags().StringVar(&helmVersion, "helm-version", "3", "Helm version in the current cluster (2|3|auto)")
	clusterCmd.PersistentFlags().BoolVar(&wide, "wide", false, "Output chart name and namespace")

	rootCmd.PersistentFlags().StringSliceVarP(&repos, "url", "u", []string{
		"https://charts.fairwinds.com/stable",
		"https://charts.fairwinds.com/incubator",
		"https://kubernetes-charts.storage.googleapis.com",
		"https://kubernetes-charts-incubator.storage.googleapis.com",
		"https://charts.jetstack.io",
	}, "URL for a helm chart repo")

	// helm hub args
	rootCmd.PersistentFlags().BoolVar(&pollHelmHub, "poll-helm-hub", true, "When true, polls all helm repos that publish to helm hub.  Default is true.")
	rootCmd.PersistentFlags().StringVar(&helmHubConfigURL, "helm-hub-config", "https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml", "The URL to the helm hub sync config.")
}

// getRepoURLs combines user specified repos and, if enabled, all helm hub repo URLs
func getRepoURLs() []string {
	if pollHelmHub {
		hc, err := nova_helm.NewHubConfig(helmHubConfigURL)
		if err == nil {
			return append(repos, hc.URLs()...)
		}
	}
	return repos
}

var clusterCmd = &cobra.Command{
	Use:   "find",
	Short: "Find out-of-date deployed releases.",
	Long:  "Find deployed helm releases that have updated charts available in chart repos",
	Run: func(cmd *cobra.Command, args []string) {

		h := nova_helm.NewHelm(helmVersion, kubeContext)
		HelmRepos := nova_helm.NewRepo(getRepoURLs())
		outputObjects, err := h.GetReleaseOutput(HelmRepos)
		out := output.Output{
			HelmReleases: outputObjects,
		}

		if err != nil {
			klog.Fatalf("Error getting helm releases from cluster: %v", err)
		}
		if outputFile != "" {
			err = out.ToFile(outputFile)
			if err != nil {
				panic(err)
			}
		} else {
			out.Print(wide)
		}
	},
}
