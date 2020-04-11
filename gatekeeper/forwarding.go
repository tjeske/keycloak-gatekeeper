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

package gatekeeper

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/tjeske/keycloak-gatekeeper/backend"
	"github.com/tjeske/keycloak-gatekeeper/config"
	mystorage "github.com/tjeske/keycloak-gatekeeper/storage"
	"go.uber.org/zap"
)

type userContainer struct {
	user      string
	container *mystorage.App
}

// var appInfo *config.Config

// func init() {
// 	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	appInfo = config.NewConfig(dir + "/apps.conf")
// }

var storageProvider mystorage.StorageProvider

func SetStorageProvider(newStorageProvider mystorage.StorageProvider) {
	storageProvider = newStorageProvider
}

var configProvider config.ConfigProvider

func setConfigProvider(newConfigProvider config.ConfigProvider) {
	configProvider = newConfigProvider
}

var dockerClient = backend.NewDockerClient("1.2.3")

// proxyMiddleware is responsible for handles reverse proxy request to the upstream endpoint
func (r *oauthProxy) proxyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req)

		if strings.HasPrefix(req.URL.Path, "/udesk") {
			return
		}
		// TODO: nur forwarden wenn kein udesk Zugriff (kein localhost/udesk/...)
		var uuid = ""
		for _, cookie := range req.Cookies() {
			if cookie.Name == "udesk_current_app" {
				uuid = cookie.Value
			}
		}

		_, err := r.getIdentity(req)

		c := dockerClient.GetContainer(uuid)
		if uuid == "" || c == nil {
			http.Redirect(w, req, "http://"+req.Host+"/udesk/admin/controlpanel.html", http.StatusSeeOther)
			return
		}

		// @step: retrieve the request scope
		scope := req.Context().Value(contextScopeName)
		if scope != nil {
			sc := scope.(*RequestScope)
			if sc.AccessDenied {
				return
			}
		}

		// @step: add the proxy forwarding headers
		req.Header.Add("X-Forwarded-For", realIP(req))
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", req.Header.Get("X-Forwarded-Proto"))

		if len(r.config.CorsOrigins) > 0 {
			// if CORS is enabled by gatekeeper, do not propagate CORS requests upstream
			req.Header.Del("Origin")
		}
		// @step: add any custom headers to the request
		for k, v := range r.config.Headers {
			req.Header.Set(k, v)
		}

		// *****************

		// appName := "unknown"
		// rawParts := strings.Split(req.URL.Path, "/")
		// var parts []string
		// for _, part := range rawParts {
		// 	if part != "" {
		// 		parts = append(parts, part)
		// 	}
		// }
		// if len(parts) > 0 {
		// 	appName = parts[0]
		// 	// req.URL.Path = strings.Join(parts[1:], "/")
		// }
		// fmt.Println("!!!!!!" + appName)
		// fmt.Println("!!!!!!" + req.URL.Path)

		// container := storageProvider.GetAppConfigByEntryPoint(req.Host)
		// if container == nil {
		// 	// cannot find container
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// _, ok := container.Access[user.name]
		// if !ok {
		// 	r.log.Warn("unauthorized access from " + user.name)
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// port, ok := runtimeCache[userContainer{user: user.name, container: container}]
		// fmt.Println(runtimeCache)
		// if !ok {
		// 	port, err = freeport.GetFreePort()
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	y := container.Access[user.name]
		// 	args := y.Args
		// 	if args == nil {
		// 		args = make(map[string]string)
		// 	}
		// 	args["user"] = user.name
		// 	dockerRunArgs := []string{
		// 		"-d",
		// 		"-p", strconv.Itoa(port) + ":" + strconv.Itoa(container.InternalPort),
		// 		"--label", "bcb_entrypoint=" + container.EntryPoint,
		// 	}
		// 	dockerClient.Run(container.Name, user.name, args, container, dockerRunArgs, []string{})
		// 	runtimeCache[userContainer{user: user.name, container: container}] = port
		// 	time.Sleep(1 * time.Second)
		// 	r.dropCookie(w, req.Host, "test", "testvalue", 0)
		// }

		// find container with label abc
		// read port label

		endpoint, err := url.Parse("http://localhost:" + c.Labels["udesk_entry_port"])

		// @note: by default goproxy only provides a forwarding proxy, thus all requests have to be absolute and we must update the host headers
		req.URL.Host = endpoint.Host
		req.URL.Scheme = endpoint.Scheme
		if v := req.Header.Get("Host"); v != "" {
			req.Host = v
			req.Header.Del("Host")
		} else if !r.config.PreserveHost {
			req.Host = endpoint.Host
		}

		if isUpgradedConnection(req) {
			r.log.Debug("upgrading the connnection", zap.String("client_ip", req.RemoteAddr))
			if err := tryUpdateConnection(req, w, endpoint); err != nil {
				r.log.Error("failed to upgrade connection", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		proxy, err := r.createUpstreamProxy(endpoint)
		if err != nil {
			return
		}

		spew.Config = spew.ConfigState{SortKeys: true}
		spew.Dump(req)
		proxy.ServeHTTP(w, req)
	})
}
