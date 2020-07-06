package cmd

import (
	"fmt"
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
)

func init() {
	rootCmd.AddCommand(clusterCmd)
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output-file", "", "", "Path on local filesystem to write file output to")
	clusterCmd.PersistentFlags().StringVar(&helmVersion, "helm-version", "3", "Helm version in the current cluster (2|3|auto)")

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

		HelmRepos := nova_helm.NewRepo(getRepoURLs())
		outputObjects, err := nova_helm.GetReleaseOutput(helmVersion, HelmRepos)
		out := output.Output{outputObjects}

		if err != nil {
			klog.Fatalf("Error getting helm releases from cluster: %v", err)
		}
		if outputFile != "" {
			err = out.ToFile(outputFile)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("\n\n")
			fmt.Println(out)
		}
	},
}
