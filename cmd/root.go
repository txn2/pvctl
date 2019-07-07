// Copyright © 2019 TXN2
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
	"fmt"
	"os"

	"github.com/prometheus/common/log"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "v0.0.0"
	header  = `                _   _
 _ ____   _____| |_| |
| '_ \ \ / / __| __| |
| |_) \ V / (__| |_| |
| .__/ \_/ \___|\__|_|
|_| ` + version + `
`

	cfgFile string
	backend string

	rootCmd = &cobra.Command{
		Use:   "pvctl [command]",
		Short: "Provision CLI",
		Long: header + `
Command Line Interface to the Provision API. 
See: https://github.com/txn2/provision
`,
		Run: func(c *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = c.Help()
				os.Exit(0)
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// global application flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pvctl.yaml)")
	rootCmd.PersistentFlags().StringVar(&backend, "backend", "http://api-provision:8070", "location of the Provision service")

	// config binding
	err := viper.BindPFlag("backend", rootCmd.PersistentFlags().Lookup("backend"))
	if err != nil {
		log.Fatalf("could not bind to configuration: %s", err.Error())
		os.Exit(1)
	}

	// Add sub commands
	rootCmd.AddCommand(applyCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pvctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pvctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
