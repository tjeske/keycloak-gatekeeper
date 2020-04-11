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

package storage

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/tjeske/containerflight/util"
	"gopkg.in/yaml.v2"
)

type fileProvider struct {
	configFile string
	apps       []*App
}

var mu sync.Mutex

func NewFileProvider(configFile string) StorageProvider {

	absConfigFile, err := filepath.Abs(configFile)
	util.CheckErr(err)

	fileProvider := fileProvider{configFile: absConfigFile}

	fileProvider.loadConfigFile()
	return &fileProvider
}

func (p *fileProvider) loadConfigFile() {
	yamlConfigFileReader, err := filesystem.Open(p.configFile)
	util.CheckErr(err)

	// yamlSpec := getConfig(yamlConfigFileReader)

	// read the config file
	yamlFileBytes, err := ioutil.ReadAll(yamlConfigFileReader)
	util.CheckErr(err)
	str := string(yamlFileBytes)

	// unmarshal yaml file
	config := []*App{}
	err = yaml.UnmarshalStrict([]byte(str), &config)
	util.CheckErr(err)

	p.apps = config
}

func (p *fileProvider) saveConfigFile() {
	yamlConfigFileWriter, err := filesystem.Open(p.configFile)
	defer yamlConfigFileWriter.Close()
	util.CheckErr(err)

	// marshal yaml file
	yamlFileBytes, err := yaml.Marshal(p.apps)
	util.CheckErr(err)

	// write the config file
	_, err = yamlConfigFileWriter.Write(yamlFileBytes)
	util.CheckErr(err)
}

func (p *fileProvider) UpdateApp(app App) {
	mu.Lock()
	defer mu.Unlock()

	found := false
	for i, app := range p.apps {
		if app.Name == app.Name {
			p.apps[i] = app
			found = true
		}
	}
	if !found {
		p.apps = append(p.apps, &app)
	}
	p.saveConfigFile()

	log.Debugf("Update configuration for app '%s': %+v", app.Name, app)
}

func (p *fileProvider) GetAllTemplates() []*App {
	mu.Lock()
	defer mu.Unlock()

	return p.apps
}

func (p *fileProvider) GetTemplateByName(name string) *App {
	mu.Lock()
	defer mu.Unlock()

	for _, app := range p.apps {
		if app.Name == name {
			return app
		}
	}
	return nil
}

func (p *fileProvider) GetAppConfigByEntryPoint(entryPoint string) *App {
	mu.Lock()
	defer mu.Unlock()
	for _, app := range p.apps {
		if app.EntryPoint == entryPoint {
			return app
		}
	}
	return nil
}
