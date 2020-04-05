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
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/jose"
	"github.com/coreos/go-oidc/oidc"
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

var runtimeCache = make(map[userContainer]int)

var dockerClient = backend.NewDockerClient("1.2.3")

// proxyMiddleware is responsible for handles reverse proxy request to the upstream endpoint
func (r *oauthProxy) proxyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req)

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

		user, err := r.getIdentity(req)
		if err != nil {
			fmt.Println(user)
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
		fmt.Println(req.Cookies())
		var uuid = ""
		for _, cookie := range req.Cookies() {
			if cookie.Name == "udesk_current_app" {
				uuid = cookie.Value
				fmt.Println(cookie)
			}
		}

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

		c := dockerClient.GetContainer(uuid)
		if uuid == "" || c == nil {
			http.Redirect(w, req, "http://"+req.Host+"/admin/controlpanel.html", http.StatusSeeOther)
			return
		}
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

// forwardProxyHandler is responsible for signing outbound requests
func (r *oauthProxy) forwardProxyHandler() func(*http.Request, *http.Response) {
	client, err := r.client.OAuthClient()
	if err != nil {
		r.log.Fatal("failed to create oauth client", zap.Error(err))
	}
	// the loop state
	var state struct {
		// the access token
		token jose.JWT
		// the refresh token if any
		refresh string
		// the identity of the user
		identity *oidc.Identity
		// the expiry time of the access token
		expiration time.Time
		// whether we need to login
		login bool
		// whether we should wait for expiration
		wait bool
	}
	state.login = true

	// create a routine to refresh the access tokens or login on expiration
	go func() {
		for {
			state.wait = false

			// step: do we have a access token
			if state.login {
				r.log.Info("requesting access token for user",
					zap.String("username", r.config.ForwardingUsername))

				// step: login into the service
				resp, err := client.UserCredsToken(r.config.ForwardingUsername, r.config.ForwardingPassword)
				if err != nil {
					r.log.Error("failed to login to authentication service", zap.Error(err))
					// step: back-off and reschedule
					<-time.After(time.Duration(5) * time.Second)
					continue
				}

				// step: parse the token
				token, identity, err := parseToken(resp.AccessToken)
				if err != nil {
					r.log.Error("failed to parse the access token", zap.Error(err))
					// step: we should probably hope and reschedule here
					<-time.After(time.Duration(5) * time.Second)
					continue
				}

				// step: update the loop state
				state.token = token
				state.identity = identity
				state.expiration = identity.ExpiresAt
				state.wait = true
				state.login = false
				state.refresh = resp.RefreshToken

				r.log.Info("successfully retrieved access token for subject",
					zap.String("subject", state.identity.ID),
					zap.String("email", state.identity.Email),
					zap.String("expires", state.expiration.Format(time.RFC3339)))

			} else {
				r.log.Info("access token is about to expiry",
					zap.String("subject", state.identity.ID),
					zap.String("email", state.identity.Email))

				// step: if we a have a refresh token, we need to login again
				if state.refresh != "" {
					r.log.Info("attempting to refresh the access token",
						zap.String("subject", state.identity.ID),
						zap.String("email", state.identity.Email),
						zap.String("expires", state.expiration.Format(time.RFC3339)))

					// step: attempt to refresh the access
					token, newRefreshToken, expiration, _, err := getRefreshedToken(r.client, state.refresh)
					if err != nil {
						state.login = true
						switch err {
						case ErrRefreshTokenExpired:
							r.log.Warn("the refresh token has expired, need to login again",
								zap.String("subject", state.identity.ID),
								zap.String("email", state.identity.Email))
						default:
							r.log.Error("failed to refresh the access token", zap.Error(err))
						}
						continue
					}

					// step: update the state
					state.token = token
					state.expiration = expiration
					state.wait = true
					state.login = false
					if newRefreshToken != "" {
						state.refresh = newRefreshToken
					}

					// step: add some debugging
					r.log.Info("successfully refreshed the access token",
						zap.String("subject", state.identity.ID),
						zap.String("email", state.identity.Email),
						zap.String("expires", state.expiration.Format(time.RFC3339)))

				} else {
					r.log.Info("session does not support refresh token, acquiring new token",
						zap.String("subject", state.identity.ID),
						zap.String("email", state.identity.Email))

					// we don't have a refresh token, we must perform a login again
					state.wait = false
					state.login = true
				}
			}

			// wait for an expiration to come close
			if state.wait {
				// set the expiration of the access token within a random 85% of actual expiration
				duration := getWithin(state.expiration, 0.85)
				r.log.Info("waiting for expiration of access token",
					zap.String("token_expiration", state.expiration.Format(time.RFC3339)),
					zap.String("renewal_duration", duration.String()))

				<-time.After(duration)
			}
		}
	}()

	return func(req *http.Request, resp *http.Response) {
		hostname := req.Host
		req.URL.Host = hostname
		// is the host being signed?
		if len(r.config.ForwardingDomains) == 0 || containsSubString(hostname, r.config.ForwardingDomains) {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", state.token.Encode()))
			req.Header.Set("X-Forwarded-Agent", prog)
		}
	}
}