/*
Copyright © 2021 MagicalLiebe <magical.liebe@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// emailCmd represents the email command
var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Set your e-mail address for paiza.jp",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := updateConfig("email", args[0])
		if err != nil {
			log.Fatalln(err)
		}
		err = printConfig()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	configCmd.AddCommand(emailCmd)
}
