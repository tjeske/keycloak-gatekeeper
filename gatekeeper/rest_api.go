package gatekeeper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v4"
	"github.com/docker/go-units"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/tjeske/keycloak-gatekeeper/backend"
	storge "github.com/tjeske/keycloak-gatekeeper/storage"
)

type udeskOauthProxy struct {
	*oauthProxy
	dockerClient *backend.DockerClient
}

// für /searchUser
type Answer struct {
	Results []Profile `json:"results"`
}

type Profile struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

var mutex = &sync.Mutex{}

func newUdeskProxy() *udeskOauthProxy {
	dc := backend.NewDockerClientWithWriter("1.2.3", rb)
	proxy := &udeskOauthProxy{dockerClient: dc}
	return proxy
}
func (r *udeskOauthProxy) createReverseProxy() error {

	// TODO: find better solution
	engine, ok := r.oauthProxy.router.(*chi.Mux)
	if !ok {
		panic("cannot cast to *chi.Mux")
	}

	engine.With(proxyDenyMiddleware).Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", 301)
	})

	fs := http.StripPrefix("/admin", http.FileServer(http.Dir("frontend/dist")))
	engine.With(r.authenticationMiddleware(), proxyDenyMiddleware).Get("/admin/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))

	engine.Get("/getTemplates", r.getTemplates)

	engine.With(r.authenticationMiddleware(),
		r.identityHeadersMiddleware(r.config.AddClaims)).Get("/startApp", r.startApp)

	engine.With(r.authenticationMiddleware()).Get("/removeApp/{query}", r.removeApp)

	engine.With(r.authenticationMiddleware()).Get("/pauseApp/{query}", r.pauseApp)

	engine.With(r.authenticationMiddleware()).Get("/unpauseApp/{query}", r.unpauseApp)

	engine.With(r.authenticationMiddleware()).Get("/switchApp/{query}", r.switchApp)

	engine.With(r.authenticationMiddleware()).Get("/dockerStatus", r.dockerStatus)

	engine.With(r.authenticationMiddleware()).Get("/searchUser/{query}", r.searchUser)

	return nil
}

func (r *udeskOauthProxy) getTemplates(w http.ResponseWriter, req *http.Request) {
	apps := storageProvider.GetAllTemplates()
	js, err := json.Marshal(struct {
		Data []*storge.App `json:"data"`
	}{apps})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (r *udeskOauthProxy) startApp(w http.ResponseWriter, req *http.Request) {

	templateName := req.URL.Query().Get("templateName")
	if templateName == "" {
		http.Error(w, "cannot find template name in request", http.StatusInternalServerError)
		return
	}

	app := storageProvider.GetTemplateByName(templateName)
	if app == nil {
		http.Error(w, "cannot find template", http.StatusInternalServerError)
		return
	}

	// get free internal port
	mutex.Lock()
	port, err := freeport.GetFreePort()
	mutex.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get identity
	user, err := r.getIdentity(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	args := make(map[string]string)
	for k, v := range req.URL.Query() {
		if k == "user" {
			args[k] = user.name
		} else if k == "templateName" {
			// skip
		} else {
			args[k] = strings.Join(v, " ")
		}
	}

	uuid := uuid.New().String()
	name := req.URL.Query().Get("name")
	dockerRunArgs := []string{
		"-d",
		"-p", strconv.Itoa(port) + ":" + strconv.Itoa(app.InternalPort),
		"--label", "udesk_uuid=" + uuid,
		"--label", "udesk_entry_port=" + strconv.Itoa(port),
		"--label", "udesk_owner=" + user.name,
		"--label", "udesk_name=" + name,
	}

	go r.dockerClient.Run(app.Name, user.name, args, app, dockerRunArgs, []string{})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func (r *udeskOauthProxy) removeApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	err := r.dockerClient.RemoveContainer(containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) pauseApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	err := r.dockerClient.PauseContainer(containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) unpauseApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	err := r.dockerClient.UnpauseContainer(containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) switchApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	fmt.Println(containerID)
	r.dropCookie(w, req.Host, "udesk_current_app", containerID, 0)
	fmt.Println(req.Host)
	http.Redirect(w, req, "http://"+req.Host, http.StatusSeeOther)
}

func (r *udeskOauthProxy) dockerStatus(w http.ResponseWriter, req *http.Request) {

	containers := r.dockerClient.GetStatus()
	y := [][]string{}
	for _, container := range containers {

		t := time.Unix(container.Created, 0)

		// name
		name := "UNKNOWN"
		if nameLabel, ok := container.Labels["udesk_name"]; ok {
			name = nameLabel
		}

		// owner
		owner := "UNKNOWN"
		if ownerLabel, ok := container.Labels["udesk_owner"]; ok {
			owner = ownerLabel
		}

		// entrypoint
		uuid := "UNKNOWN"
		if entryPointLabel, ok := container.Labels["udesk_uuid"]; ok {
			uuid = entryPointLabel
		}

		y = append(y, []string{
			name,
			owner,
			units.HumanDuration(time.Now().UTC().Sub(t)) + " ago",
			container.State,
			uuid,
		})
	}

	js, err := json.Marshal(struct {
		Data [][]string `json:"data"`
	}{y})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (r *udeskOauthProxy) searchUser(w http.ResponseWriter, req *http.Request) {

	u, err := r.getIdentity(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := gocloak.NewClient("http://auth.familie-jeske.net:7080/")
	token := u.token.RawHeader + "." + u.token.RawPayload + "." + strings.TrimRight(base64.URLEncoding.EncodeToString(u.token.Signature), "=")
	users, err := client.GetUsers(
		token,
		"master",
		gocloak.GetUsersParams{})

	filteredUserOrGroups := []Profile{}
	for _, user := range users {
		if strings.HasPrefix(strings.ToLower(*user.Username), strings.ToLower(chi.URLParam(req, "query"))) {
			description := ""
			if user.FirstName != nil {
				description += *user.FirstName
			}
			if user.LastName != nil {
				if description != "" {
					description += " "
				}
				description += *user.LastName
			}
			if description != "" {
				description = *user.Username
			}
			filteredUserOrGroups = append(filteredUserOrGroups, Profile{*user.Username, description})
		}
	}

	js, err := json.Marshal(Answer{filteredUserOrGroups})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
