// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"log"

	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/client"
	"github.com/spf13/cobra"
)

// testitCmd represents the testit command
var testitCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connectivity to secrets bridge.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		bridgeConf, err := cmd.Flags().GetString("bridge-conf")
		if err != nil {
			log.Fatalln("--bridge-conf invalid:", err)
		}

		bridge, err := bridge.NewFromString(bridgeConf)
		if err != nil {
			log.Fatalln("--bridge-conf has an invalid value:", err)
		}

		c := client.NewClient(bridge)
		err = c.ChooseEndpoint()
		if err != nil {
			log.Fatalln("error pinging server:", err)
		}

		fmt.Println("bridge server responding")
	},
}

func init() {
	RootCmd.AddCommand(testitCmd)
}
