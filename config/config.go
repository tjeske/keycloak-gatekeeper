// Copyright © 2019 Tobias Jeske
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

package config

import (
	"io"
	"io/ioutil"
	"path/filepath"

	yaml "github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/tjeske/containerflight/util"
	"github.com/tjeske/keycloak-gatekeeper/backend"
)

// "mock connectors" for unit-tesing
var logFatalf = log.Fatalf
var filesystem = afero.NewOsFs()

type ConfigProvider interface {
}

// specification of an app file
type yamlSpec struct {
	Storage Storage
	Backend Backend
}

type Storage struct {
	Type    string
	File    File
	MongoDb MongoDb
}

type MongoDb struct {
	Host     string
	User     string
	Password string
	Database string
}

type File struct {
	AppsFile string
}

type Backend struct {
	Type   string
	Docker backend.BackendDockerConfig
}

// NewConfig returns a representation of an application config file
func NewConfig(configFile string) *yamlSpec {

	absConfigFile, err := filepath.Abs(configFile)
	util.CheckErr(err)

	yamlConfigFileReader, err := filesystem.Open(absConfigFile)
	util.CheckErr(err)

	config := getConfig(yamlConfigFileReader)

	validate(config)

	return &config
}

// // NewFakeAppInfo returns a fake representation of an application config file for unit-testing
// func NewFakeAppInfo(fs *afero.Fs, appConfigFile string, appConfigStr string) *AppInfo {
// 	origFS := filesystem
// 	defer func() { filesystem = origFS }()
// 	filesystem = *fs

// 	afero.WriteFile(filesystem, appConfigFile, []byte(appConfigStr), 0644)

// 	return NewConfig(appConfigFile)
// }

// validate config file
func validate(config yamlSpec) {
}

// read and parse config file
func getConfig(yamlConfigReader io.Reader) yamlSpec {

	// read the config file
	yamlFileBytes, err := ioutil.ReadAll(yamlConfigReader)
	util.CheckErr(err)
	str := string(yamlFileBytes)

	// unmarshal yaml file
	config := yamlSpec{}
	err = yaml.UnmarshalStrict([]byte(str), &config)
	util.CheckErr(err)

	return config
}

// func (cfg *AppInfo) GetContainer(entryPoint string) (*App, error) {
// 	containers := cfg.AppConfig.Containers
// 	for i := range containers {
// 		if containers[i].EntryPoint == entryPoint {
// 			return &containers[i], nil
// 		}
// 	}
// 	return nil, errors.New("Container not found")
// }

func (m *MongoDb) GetMongoDbHost() string {
	if m.Host != "" {
		return m.Host
	}
	return "localhost:27017"
}

func (m *MongoDb) GetMongoDbUser() string {
	if m.Host != "" {
		return m.User
	}
	return "root"
}

func (m *MongoDb) GetMongoDbPassword() string {
	return m.Password
}

func (m *MongoDb) GetMongoDbDataBase() string {
	if m.Database != "" {
		return m.Database
	}
	return "ccb"
}
