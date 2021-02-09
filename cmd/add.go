/*
Copyright © 2021 reeve0930 <reeve0930@gmail.com>

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
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type templateConfig struct {
	Lang string `toml:"language"`
	File string `toml:"file"`
	Run  string `toml:"run"`
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a template files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := addTemplate(args[0]); err != nil {
			log.Fatalln(err)
		}
	},
}

func addTemplate(name string) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}
	p := filepath.Join(configDir, name)
	if f, err := os.Stat(p); os.IsNotExist(err) || !f.IsDir() {
		err = os.Mkdir(p, 0755)
		if err != nil {
			return err
		}
		main := filepath.Join(p, "main.xx")
		if !Exists(main) {
			if _, err = os.Create(main); err != nil {
				return err
			}
		}
		configFile := filepath.Join(p, "template.toml")
		if !Exists(configFile) {
			if _, err = os.Create(configFile); err != nil {
				return err
			}
			tempConfig := templateConfig{File: "main.xx"}
			var buffer bytes.Buffer
			encoder := toml.NewEncoder(&buffer)
			if err := encoder.Encode(tempConfig); err != nil {
				return err
			}
			if err = writeFile(configFile, buffer.String()); err != nil {
				return err
			}
			fmt.Printf("create template files: %v\n", p)
		}
	} else {
		fmt.Printf("template: '%v' already exists\n", name)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
