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

	"github.com/spf13/cobra"
	"github.com/txn2/pvctl/util"
)

var (
	applyPaths []string

	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply Provision objects.",
		Long: header + `
Apply Provision Account, User and Asset objects from YAML configuration files.`,
		Run: func(cmd *cobra.Command, args []string) {

			pvOs := util.NewPvObjectStore(provisionBackend)

			for _, pth := range applyPaths {
				err := pvOs.LoadObjectsFromPath(pth)
				if err != nil {
					fmt.Printf("apply command returned error: %s\n", err.Error())
					os.Exit(1)
				}
			}

			if pvOs.ObjectsInStore() < 1 {
				fmt.Println("No objects found for apply.")
				return
			}

			fmt.Printf("apply found %d object(s) loaded and ready to send.\n", pvOs.ObjectsInStore())

			err := pvOs.SendObjects()
			if err != nil {
				fmt.Printf("apply error sending object: %s\n", err.Error())
				os.Exit(1)
			}

		},
	}
)

func init() {

	// global application flags
	applyCmd.Flags().StringArrayVarP(&applyPaths, "file", "f", []string{}, "Source directory or yaml file to read from")
	_ = applyCmd.MarkFlagRequired("file")
}
