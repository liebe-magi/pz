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

	"os/user"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type Config struct {
	User    UserConfig
	Setting SettingConfig
}

type UserConfig struct {
	Email string `toml:"email"`
	Pass  string `toml:"password"`
}

type SettingConfig struct {
	DefalutTemp string `toml:"defaultTemplate"`
	// AutoSubmit  bool   `toml:"autoSubmit"`
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Check and change the settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := printConfig()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func printConfig() error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	configString, err := encodeConfig(config)
	if err != nil {
		return err
	}
	fmt.Println(configString)
	return nil
}

func updateConfig(key, value string) error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	if key == "email" {
		config.User.Email = value
	} else if key == "password" {
		config.User.Pass = value
	} else if key == "defaultTemplate" {
		config.Setting.DefalutTemp = value
		// } else if key == "autoSubmit" {
		// if value == "true" {
		// config.Setting.AutoSubmit = true
		// } else if value == "false" {
		// config.Setting.AutoSubmit = false
		// } else {
		// return fmt.Errorf("Unknown value: %v", value)
		// }
	} else {
		return fmt.Errorf("Unknown key: %v", key)
	}
	configString, err := encodeConfig(config)
	if err != nil {
		return err
	}
	p, err := getConfigPath()
	if err != nil {
		return err
	}
	if err = writeFile(p, configString); err != nil {
		return err
	}
	return nil
}

func createConfigFile() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// mkdir: ~/.config/pz
	if f, err := os.Stat(configDir); os.IsNotExist(err) || !f.IsDir() {
		if err = os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	// write: ~/.config/pz/config.toml
	p, err := getConfigPath()
	if err != nil {
		return err
	}
	if !Exists(p) {
		config := Config{}
		configString, err := encodeConfig(config)
		if err != nil {
			return err
		}
		if err = writeFile(p, configString); err != nil {
			return err
		}
	}
	return nil
}

func getConfigPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(u.HomeDir, ".config", "pz", "config.toml")
	return p, nil
}

func getConfigDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(u.HomeDir, ".config", "pz")
	return p, nil
}

func getConfig() (Config, error) {
	p, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}
	var config Config
	_, err = toml.DecodeFile(p, &config)
	return config, nil
}

func encodeConfig(config Config) (string, error) {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if err := encoder.Encode(config); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func Exists(filename string) bool {
	if f, err := os.Stat(filename); os.IsNotExist(err) || f.IsDir() {
		return false
	}
	return true
}

func init() {
	rootCmd.AddCommand(configCmd)
}
