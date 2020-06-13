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
	"fmt"
	"regexp"
)

type Template struct {
	Name         string            `json:"name"`
	Params       map[string]string `json:"params"`
	Dockerfile   string            `json:"dockerfile"`
	Files        map[string]string `json:"files"`
	Access       map[string]access `json:"access"`
	InternalPort int               `json:"internalPort"`
}

type access struct {
	Permissions string `json:"permissions"`
	Args        map[string]string
}

var parameterRegex = regexp.MustCompile(`([^\\])(?P<param>(\$\{[[:word:]]+\}))`)
var escapedParameterRegex = regexp.MustCompile(`\\\$\{[[:word:]]+\}`)

// search and replace parameters in string
func replaceParameters(str *string, args map[string]string) {
	*str = parameterRegex.ReplaceAllStringFunc(*str, func(match string) string {
		split := parameterRegex.FindStringSubmatch(match)
		fmt.Println(split)
		param := split[2][2 : len(split[2])-1]
		value, ok := args[param]
		if ok {
			return split[1] + value
		}

		return "<<ERROR!>>"
	})
	*str = escapedParameterRegex.ReplaceAllStringFunc(*str, func(match string) string {
		return match[1:len(match)]
	})
}

// GetDockerRunArgs returns for an app file the resolved docker run arguments
func (a *Template) GetDockerRunArgs() (dockerRunArgs []string) {
	return []string{}
}

func (a *Template) GetDockerFile(args map[string]string) string {
	dockerFile := a.Dockerfile
	replaceParameters(&dockerFile, args)
	return dockerFile
}

func (a *Template) GetDockerFileContextFiles(args map[string]string) map[string]string {
	resolvedFiles := make(map[string]string)
	for file, content := range a.Files {
		replaceParameters(&content, args)
		resolvedFiles[file] = content
	}
	return resolvedFiles
}
