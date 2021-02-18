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
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	nova_helm "github.com/fairwindsops/nova/pkg/helm"
	"github.com/fairwindsops/nova/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

var (
	version       string
	versionCommit string
	cfgFile       string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(genConfigCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file to use. If empty, flags will be used instead")

	rootCmd.PersistentFlags().String("output-file", "", "Path on local filesystem to write file output to")
	viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("output-file"))

	rootCmd.PersistentFlags().StringToStringP("desired-versions", "d", nil, "A map of chart=override_version to override the helm repository when checking.")
	viper.BindPFlag("desired-versions", rootCmd.PersistentFlags().Lookup("desired-versions"))

	rootCmd.PersistentFlags().StringSliceP("url", "u", []string{
		"https://charts.fairwinds.com/stable",
		"https://charts.fairwinds.com/incubator",
		"https://kubernetes-charts.storage.googleapis.com",
		"https://kubernetes-charts-incubator.storage.googleapis.com",
		"https://charts.jetstack.io",
	}, "URL for a helm chart repo")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))

	// helm hub args
	rootCmd.PersistentFlags().Bool("poll-helm-hub", true, "When true, polls all helm repos that publish to helm hub.  Default is true.")
	viper.BindPFlag("poll-helm-hub", rootCmd.PersistentFlags().Lookup("poll-helm-hub"))

	rootCmd.PersistentFlags().String("helm-hub-config", "https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml", "The URL to the helm hub sync config.")
	viper.BindPFlag("helm-hub-config", rootCmd.PersistentFlags().Lookup("helm-hub-config"))

	rootCmd.PersistentFlags().String("context", "", "A context to use in the kubeconfig.")
	viper.BindPFlag("context", rootCmd.PersistentFlags().Lookup("context"))

	rootCmd.PersistentFlags().String("helm-version", "3", "Helm version in the current cluster (2|3|auto)")
	viper.BindPFlag("helm-version", rootCmd.PersistentFlags().Lookup("helm-version"))

	rootCmd.PersistentFlags().Bool("wide", false, "Output chart name and namespace")
	viper.BindPFlag("wide", rootCmd.PersistentFlags().Lookup("wide"))

	klog.InitFlags(nil)
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func initConfig() {
	if cfgFile == "" {
		klog.V(2).Infof("config not set, using flags only")
		return
	}

	if strings.Contains(cfgFile, "http") {
		klog.V(2).Infof("detected URL for config location")
		var err error
		cfgFile, err = downloadConfig(cfgFile)
		if err != nil {
			klog.Fatalf("failed to download config: %s", err.Error())
		}
		defer os.Remove(cfgFile)
	}

	// Read config
	viper.SetConfigFile(cfgFile)
	klog.V(2).Infof("using config file: %s", cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		klog.V(2).Infof("could not read config file %s - ignoring it", err.Error())
	}
}

func downloadConfig(cfgURL string) (string, error) {
	fileURL, err := url.Parse(cfgURL)
	if err != nil {
		return "", err
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	file, err := ioutil.TempFile("", fmt.Sprintf("*-%s", fileName))
	if err != nil {
		return "", err
	}

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(cfgURL)
	if err != nil {
		return "", &errors.StatusError{}
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	tmpConfig := file.Name()
	klog.V(2).Infof("downloaded config file %s with size %d", tmpConfig, size)

	return tmpConfig, nil
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

// getRepoURLs combines user specified repos and, if enabled, all helm hub repo URLs
func getRepoURLs() []string {
	repos := viper.GetStringSlice("url")
	if viper.GetBool("poll-helm-hub") {
		hc, err := nova_helm.NewHubConfig(viper.GetString("helm-hub-config"))
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
		h := nova_helm.NewHelm(viper.GetString("helm-version"), viper.GetString("context"))

		klog.V(4).Infof("Settings: %v", viper.AllSettings())
		klog.V(4).Infof("All Keys: %v", viper.AllKeys())

		if viper.IsSet("desired-versions") {
			klog.V(3).Infof("desired-versions is set - attempting to load them")
			klog.V(7).Infof("raw desired-versions: %v", viper.Get("desired-versions"))

			desiredVersion := viper.GetStringMapString("desired-versions")
			for k, v := range desiredVersion {
				klog.V(2).Infof("version override for %s: %s", k, v)
				h.DesiredVersions = append(h.DesiredVersions, nova_helm.DesiredVersion{
					Name:    k,
					Version: v,
				})
			}
		}
		HelmRepos := nova_helm.NewRepo(getRepoURLs())
		outputObjects, err := h.GetReleaseOutput(HelmRepos)
		out := output.Output{
			HelmReleases: outputObjects,
		}

		if err != nil {
			klog.Fatalf("Error getting helm releases from cluster: %v", err)
		}
		outputFile := viper.GetString("output-file")
		if outputFile != "" {
			err = out.ToFile(outputFile)
			if err != nil {
				panic(err)
			}
		} else {
			out.Print(viper.GetBool("wide"))
		}
	},
}

var genConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Generate a config file.",
	Long:  "Generate a configuration file with all of the default configuration values.",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.SafeWriteConfigAs(cfgFile)
		if err != nil {
			klog.Fatal(err)
		}
	},
}

// Execute is the main entry point into the command
func Execute(VERSION, COMMIT string) {
	version = VERSION
	versionCommit = COMMIT
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
