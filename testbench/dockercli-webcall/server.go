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
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/phayes/freeport"
)

type Foo struct{}

func (f *Foo) GetDockerFile(args map[string]string) string {
	return `FROM ubuntu:20.04
			CMD tail -f /dev/null`
}

func (f *Foo) GetDockerFileContextFiles(args map[string]string) map[string]string {
	return make(map[string]string, 0)
}

var addr = flag.String("addr", "localhost:9090", "http service address")
var dockerClient = NewDockerClient("1.2.3")

func echo(w http.ResponseWriter, r *http.Request) {

	// log.Level = logrus.DebugLevel

	// get free internal port
	port, err := freeport.GetFreePort()
	if err != nil {
		return
	}

	args := make(map[string]string)

	uuid := uuid.New()
	uuidStr := uuid.String()
	name := "jhghjg"
	dockerRunArgs := []string{
		"--rm",
		"-p", strconv.Itoa(port) + ":" + strconv.Itoa(8080),
		"--label", "udesk_uuid=" + uuidStr,
		"--label", "udesk_entry_port=" + strconv.Itoa(port),
		"--label", "udesk_owner=" + "foo",
		"--label", "udesk_name=" + name,
	}

	// app := NewApp()
	// r.appHub.addApp(uuid, app)

	dockerClient.Create("ttyd", "foo", args, dockerRunArgs, []string{}, func() {})
	x, _ := dockerClient.GetContainer(uuidStr)
	dockerClient.Start(x.ID)

	// app := gatekeeper.NewOauthProxyApp()
	// _ = app.Run(os.Args)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	fmt.Println("finished")
}

func main() {
	// port, err := freeport.GetFreePort()
	// if err != nil {
	// 	return
	// }

	// args := make(map[string]string)

	// uuid := uuid.New()
	// uuidStr := uuid.String()
	// name := "jhghjg"
	// dockerRunArgs := []string{
	// 	"-d",
	// 	"--rm",
	// 	"-p", strconv.Itoa(port) + ":" + strconv.Itoa(8080),
	// 	"--label", "udesk_uuid=" + uuidStr,
	// 	"--label", "udesk_entry_port=" + strconv.Itoa(port),
	// 	"--label", "udesk_owner=" + "foo",
	// 	"--label", "udesk_name=" + name,
	// }

	// // app := NewApp()
	// // r.appHub.addApp(uuid, app)

	// dockerClient.Run("ttyd", "foo", args, dockerRunArgs, []string{}, func() {})
	// time.Sleep(5 * time.Second)
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
