/*
Copyright 2015 All rights reserved.
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

package main

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/tjeske/containerflight/util"
	"github.com/tjeske/keycloak-gatekeeper/config"
	"github.com/tjeske/keycloak-gatekeeper/gatekeeper"
	"github.com/tjeske/keycloak-gatekeeper/storage"
)

var log = logrus.New()

func main() {
	log.Level = logrus.DebugLevel

	// determine directory of executable
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	util.CheckErr(err)

	// read config
	cfg := config.NewConfig(executableDir + "/config.conf")

	// configure storage backend
	storageType := cfg.Storage.Type
	switch storageType {
	case "file":
		{
			fileProvider := storage.NewFileProvider(executableDir + "/apps.conf")
			storage.SetProvider(fileProvider)

			gatekeeper.SetStorageProvider(fileProvider)
		}
	case "mongodb":
		{
			mongoDbConfig := cfg.Storage.MongoDb
			mongoDbProvider := storage.NewMongoDbProvider(mongoDbConfig.GetMongoDbHost(), mongoDbConfig.GetMongoDbUser(), mongoDbConfig.GetMongoDbPassword(), mongoDbConfig.GetMongoDbDataBase())
			storage.SetProvider(mongoDbProvider)
		}
	default:
		{
			log.Fatal("Storage backend unknown")
		}
	}

	apps := storage.Provider.ReturnAllApps()
	for _, app := range apps {
		log.Println(app.Name)
	}

	app := gatekeeper.NewOauthProxyApp()
	_ = app.Run(os.Args)
}
