package gatekeeper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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
}

// für /searchUser
type Answer struct {
	Results []Profile `json:"results"`
}

type Profile struct {
	Title       string `json:"title"`
	Description string `json:"description"`
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

	engine.With(r.authenticationMiddleware()).Get("/getTemplates", r.getTemplates)

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
	apps := storageProvider.ReturnAllApps()
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

	for k, v := range req.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
	}

	apps := storageProvider.ReturnAllApps()
	js, err := json.Marshal(struct {
		Data []*storge.App `json:"data"`
	}{apps})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateName := req.URL.Query().Get("templateName")
	container := storageProvider.GetAppConfigByName(templateName)
	if container == nil {
		// cannot find container
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	args := make(map[string]string)

	user, err := r.getIdentity(req)
	if err != nil {
		fmt.Println(req)
	}
	args["user"] = user.name
	uuid := uuid.New().String()
	name := req.URL.Query().Get("name")
	dockerRunArgs := []string{
		"-d",
		"-p", strconv.Itoa(port) + ":" + strconv.Itoa(container.InternalPort),
		"--label", "udesk_uuid=" + uuid,
		"--label", "udesk_entry_port=" + strconv.Itoa(port),
		"--label", "udesk_owner=" + user.name,
		"--label", "udesk_name=" + name,
	}

	// rb.Input <- "blah"
	// rb.Flush() // if needed-- useful for testing

	var dc = backend.NewDockerClientWithWriter("1.2.3", rb)
	go dc.Run(container.Name, user.name, args, container, dockerRunArgs, []string{})

	runtimeCache[userContainer{user: user.name, container: container}] = port
	time.Sleep(1 * time.Second)
	r.dropCookie(w, req.Host, "udesk_current_app", uuid, 0)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (r *udeskOauthProxy) removeApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	err := dockerClient.RemoveContainer(containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) pauseApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	err := dockerClient.PauseContainer(containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) unpauseApp(w http.ResponseWriter, req *http.Request) {
	containerID := chi.URLParam(req, "query")
	err := dockerClient.UnpauseContainer(containerID)
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

	containers := dockerClient.GetStatus()
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
