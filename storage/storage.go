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
	"log"

	"github.com/spf13/afero"
)

// "mock connectors" for unit-tesing
var logFatalf = log.Fatalf
var filesystem = afero.NewOsFs()

type StorageProvider interface {
	UpdateApp(template Template)
	GetAllTemplates() *[]TemplateName
	GetTemplateByName(name string) *Template
}

var Provider StorageProvider

func SetProvider(newProvider StorageProvider) {
	Provider = newProvider
}
