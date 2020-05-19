package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"

	nova_helm "github.com/fairwindsops/nova/pkg/helm"
	"github.com/fairwindsops/nova/pkg/output"
)

var (
	repos            []string
	pollHelmHub      bool
	helmHubConfigURL string
	outputFile       string

	currentVersion string
)

func init() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Set("alsologtostderr", "true")
	flag.Set("logtostderr", "true")
	flag.Parse()

	rootCmd.PersistentFlags().StringSliceVarP(&repos, "url", "u", []string{
		"https://charts.fairwinds.com/stable",
		"https://charts.fairwinds.com/incubator",
		"https://kubernetes-charts.storage.googleapis.com",
		"https://kubernetes-charts-incubator.storage.googleapis.com",
		"https://charts.jetstack.io",
	}, "URL for a helm chart repo")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output-file", "", "output.json", "Path on local filesystem to write file output to")

	// helm hub args
	rootCmd.PersistentFlags().BoolVar(&pollHelmHub, "poll-helm-hub", true, "When true, polls all helm repos that publish to helm hub.  Default is true.")
	rootCmd.PersistentFlags().StringVar(&helmHubConfigURL, "helm-hub-config", "https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml", "The URL to the helm hub sync config.")
}

var rootCmd = &cobra.Command{
	Use:   "nova",
	Short: "nova",
	Long:  "A fairwinds tool to check for updated chart releases.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error, you must specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			fmt.Printf("\n%v\n", err)
		}
		os.Exit(1)
	},
}

// Execute is the main entry point into the command
func Execute(version string) {
	currentVersion = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func send(op []output.ReleaseOutput) {
	out := output.Output{op}
	err := out.ToFile(outputFile)
	if err != nil {
		panic(err)
	}
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
