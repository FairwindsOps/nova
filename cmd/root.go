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
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fairwindsops/nova/pkg/containers"
	nova_helm "github.com/fairwindsops/nova/pkg/helm"
	"github.com/fairwindsops/nova/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

var (
	version       string
	versionCommit string
	cfgFile       string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(
		findCmd,
		genConfigCmd,
	)

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		os.Stderr.WriteString("\n\nWant more? Automate Nova for free with Fairwinds Insights!\n🚀 https://fairwinds.com/insights-signup/nova 🚀\n")
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file to use. If empty, flags will be used instead")
	rootCmd.PersistentFlags().String("output-file", "", "Path on local filesystem to write file output to")
	err := viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("output-file"))
	if err != nil {
		klog.Exitf("Failed to bind output-file flag: %v", err)
	}

	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace to look in. If empty, scan will be cluster-wide")
	err = viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	if err != nil {
		klog.Exitf("Failed to bind namespace flag: %v", err)
	}

	rootCmd.PersistentFlags().String("format", "json", "An output format (table, json)")
	err = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	if err != nil {
		klog.Exitf("Failed to bind format flag: %v", err)
	}

	rootCmd.PersistentFlags().StringToStringP("desired-versions", "d", nil, "A map of chart=override_version to override the helm repository when checking.")
	err = viper.BindPFlag("desired-versions", rootCmd.PersistentFlags().Lookup("desired-versions"))
	if err != nil {
		klog.Exitf("Failed to bind desired-versions flag: %v", err)
	}

	rootCmd.PersistentFlags().StringSliceP("url", "u", []string{}, "URL for a helm chart repo")
	err = viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	if err != nil {
		klog.Exitf("Failed to bind url flag: %v", err)
	}

	rootCmd.PersistentFlags().Bool("poll-artifacthub", true, "When true, polls artifacthub to match against helm releases in the cluster. If false, you must provide a url list via --url/-u. Default is true.")
	err = viper.BindPFlag("poll-artifacthub", rootCmd.PersistentFlags().Lookup("poll-artifacthub"))
	if err != nil {
		klog.Exitf("Failed to bind poll-artifacthub flag: %v", err)
	}

	rootCmd.PersistentFlags().String("context", "", "A context to use in the kubeconfig.")
	err = viper.BindPFlag("context", rootCmd.PersistentFlags().Lookup("context"))
	if err != nil {
		klog.Exitf("Failed to bind context flag: %v", err)
	}

	rootCmd.PersistentFlags().Bool("wide", false, "Output chart name and namespace")
	err = viper.BindPFlag("wide", rootCmd.PersistentFlags().Lookup("wide"))
	if err != nil {
		klog.Exitf("Failed to bind wide flag: %v", err)
	}
	rootCmd.PersistentFlags().BoolP("include-all", "a", false, "Show all charts even if no latest version is found.")
	err = viper.BindPFlag("include-all", rootCmd.PersistentFlags().Lookup("include-all"))
	if err != nil {
		klog.Exitf("Failed to bind include-all flag: %v", err)
	}

	findCmd.Flags().Bool("containers", false, "Show old container image versions instead of helm chart versions. There will be no helm output if this flag is set.")
	err = viper.BindPFlag("containers", findCmd.Flags().Lookup("containers"))
	if err != nil {
		klog.Exitf("Failed to bind containers flag: %v", err)
	}
	findCmd.Flags().Bool("helm", false, "Show old helm chart versions. You can combine this flag with --containers to have both output in a single run.")
	err = viper.BindPFlag("helm", findCmd.Flags().Lookup("helm"))
	if err != nil {
		klog.Exitf("Failed to bind containers flag: %v", err)
	}
	findCmd.Flags().Bool("show-non-semver", false, "When finding container images, show all containers even if they don't follow semver.")
	err = viper.BindPFlag("show-non-semver", findCmd.Flags().Lookup("show-non-semver"))
	if err != nil {
		klog.Exitf("Failed to bind show-non-semver flag: %v", err)
	}

	findCmd.Flags().Bool("show-errored-containers", false, "When finding container images, show errors encountered when scanning.")
	err = viper.BindPFlag("show-errored-containers", findCmd.Flags().Lookup("show-errored-containers"))
	if err != nil {
		klog.Exitf("Failed to bind show-errored-containers flag: %v", err)
	}

	rootCmd.PersistentFlags().Bool("show-old", false, "Only show charts that are not on the latest version")
	err = viper.BindPFlag("show-old", rootCmd.PersistentFlags().Lookup("show-old"))
	if err != nil {
		klog.Exitf("Failed to bind show-old flag: %v", err)
	}

	klog.InitFlags(nil)
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlag(flag.Lookup("alsologtostderr"))
	pflag.CommandLine.AddGoFlag(flag.Lookup("logtostderr"))
	pflag.CommandLine.AddGoFlag(flag.Lookup("v"))
}

func initConfig() {
	if cfgFile == "" {
		klog.V(2).Infof("config not set, using flags only")
		return
	}

	if strings.HasPrefix(cfgFile, "https://") || strings.HasPrefix(cfgFile, "http://") {
		klog.V(2).Infof("detected URL for config location")
		var err error
		cfgFile, err = downloadConfig(cfgFile)
		if err != nil {
			klog.Exitf("failed to download config: %s", err.Error())
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
	if err != nil {
		return "", err
	}

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

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find out-of-date deployed releases.",
	Long:  "Find deployed helm releases that have updated charts available in chart repos",
	Run: func(cmd *cobra.Command, args []string) {
		if !viper.GetBool("poll-artifacthub") && len(viper.GetStringSlice("url")) == 0 {
			klog.Exitf("--poll-artifacthub=false requires urls provided to the --url flag. none were provided.")
		}
		klog.V(5).Infof("Settings: %v", viper.AllSettings())
		klog.V(5).Infof("All Keys: %v", viper.AllKeys())

		kubeContext := viper.GetString("context")

		format := viper.GetString("format")
		if !(format == output.TableFormat || format == output.JSONFormat) {
			klog.Exitf("--format flag value is not valid. Run `nova find --help` to see flag options")
		}

		if viper.GetBool("helm") && viper.GetBool("containers") {
			output, err := handleHelmAndContainers(kubeContext)
			if err != nil {
				klog.Exit(err)
			}
			outputFile := viper.GetString("output-file")
			if outputFile != "" {
				err = output.ToFile(outputFile)
				if err != nil {
					klog.Exitf("error outputting to file: %s", err)
				}
			} else {
				output.Print(format, viper.GetBool("wide"), viper.GetBool("show-old"))
			}
			return
		}

		if viper.GetBool("containers") {
			output, err := handleContainers(kubeContext)
			if err != nil {
				klog.Exit(err)
			}
			output.Print(format)
			return
		}

		output, err := handleHelm(kubeContext)
		if err != nil {
			klog.Exit(err)
		}
		outputFile := viper.GetString("output-file")
		if outputFile != "" {
			err = output.ToFile(outputFile)
			if err != nil {
				klog.Exitf("error outputting to file: %s", err)
			}
		} else {
			output.Print(format, viper.GetBool("wide"), viper.GetBool("show-old"))
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
			klog.Exit(err)
		}
	},
}

// Execute is the main entry point into the command
func Execute(VERSION, COMMIT string) {
	version = VERSION
	versionCommit = COMMIT
	if err := rootCmd.Execute(); err != nil {
		klog.Exit(err)
	}
}

func handleContainers(kubeContext string) (*output.ContainersOutput, error) {
	// Set up a context we can use to cancel all operations to external container registries if we need to
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(signals)
		cancel()
	}()
	go func() {
		select {
		case <-signals:
			fmt.Print("\nCancelling operations to external container registries\n")
			cancel()
		case <-ctx.Done():
		}
	}()
	iClient := containers.NewClient(kubeContext)
	namespace := viper.GetString("namespace")
	if viper.IsSet("namespace") {
		klog.V(3).Infof("Scanning namespace %v", namespace)
	} else {
		klog.V(3).Infof("Scanning whole cluster")
	}
	containers, err := iClient.Find(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("ERROR during images.Find() %w", err)
	}
	showNonSemver := viper.GetBool("show-non-semver")
	showErrored := viper.GetBool("show-errored-containers")
	includeAll := viper.GetBool("include-all")
	return output.NewContainersOutput(containers.Images, containers.ErrImages, showNonSemver, showErrored, includeAll), nil
}

func handleHelm(kubeContext string) (*output.Output, error) {
	h := nova_helm.NewHelm(kubeContext)
	if viper.IsSet("desired-versions") {
		klog.V(3).Infof("desired-versions is set - attempting to load them")
		klog.V(8).Infof("raw desired-versions: %v", viper.Get("desired-versions"))

		desiredVersion := viper.GetStringMapString("desired-versions")
		for k, v := range desiredVersion {
			klog.V(2).Infof("version override for %s: %s", k, v)
			h.DesiredVersions = append(h.DesiredVersions, nova_helm.DesiredVersion{
				Name:    k,
				Version: v,
			})
		}
	}
	namespace := viper.GetString("namespace")
	if viper.IsSet("namespace") {
		klog.V(3).Infof("Scanning namespace %v", namespace)
	} else {
		klog.V(3).Infof("Scanning whole cluster")
	}
	releases, chartNames, err := h.GetReleaseOutput(namespace)
	if err != nil {
		return nil, fmt.Errorf("error getting helm releases: %s", err)
	}
	out := output.NewOutputWithHelmReleases(releases)
	out.IncludeAll = viper.GetBool("include-all")

	if viper.GetBool("poll-artifacthub") {
		ahClient, err := nova_helm.NewArtifactHubPackageClient(version)
		if err != nil {
			return nil, fmt.Errorf("error setting up artifact hub client: %s", err)
		}
		packageRepos, err := ahClient.MultiSearch(chartNames)
		if err != nil {
			return nil, fmt.Errorf("Error getting artifacthub package repos: %v", err)
		}
		packages := ahClient.GetPackages(packageRepos)
		klog.V(2).Infof("found %d possible package matches", len(packages))
		for _, release := range releases {
			o := nova_helm.FindBestArtifactHubMatch(release, packages)
			if o != nil {
				h.OverrideDesiredVersion(o)
				out.HelmReleases = append(out.HelmReleases, *o)
			}
		}
	}
	if len(viper.GetStringSlice("url")) > 0 {
		repos := viper.GetStringSlice("url")
		helmRepos := nova_helm.NewRepos(repos)
		outputObjects := h.GetHelmReleasesVersion(helmRepos, releases)
		out.HelmReleases = append(out.HelmReleases, outputObjects...)
	}
	return &out, nil
}

func handleHelmAndContainers(kubeContext string) (*output.HelmAndContainersOutput, error) {
	helmOutput, err := handleHelm(kubeContext)
	if err != nil {
		return nil, err
	}
	containersOutput, err := handleContainers(kubeContext)
	if err != nil {
		return nil, err
	}
	return output.NewHelmAndContainersOutput(*helmOutput, *containersOutput), nil
}
