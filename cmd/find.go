package cmd

import (
	watcher_helm "github.com/fairwindsops/nova/pkg/helm"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var (
	helmVersion string
)

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.PersistentFlags().StringVar(&helmVersion, "helm-version", "3", "Helm version in the current cluster (2|3|auto)")
}

var clusterCmd = &cobra.Command{
	Use:   "find",
	Short: "Find out-of-date deployed releases.",
	Long:  "Find deployed helm releases that have updated charts available in chart repos",
	Run: func(cmd *cobra.Command, args []string) {

		HelmRepos := watcher_helm.NewRepo(getRepoURLs())
		outputObjects, err := watcher_helm.GetReleaseOutput(helmVersion, HelmRepos)

		if err != nil {
			klog.Fatalf("Error getting helm releases from cluster: %v", err)
		}
		send(outputObjects)
	},
}
