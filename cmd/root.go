package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

var (
	currentVersion string
)

func init() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Set("alsologtostderr", "true")
	flag.Set("logtostderr", "true")
	flag.Parse()
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
