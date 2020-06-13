package gatekeeper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v5"
	"github.com/docker/go-units"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	storge "github.com/tjeske/keycloak-gatekeeper/storage"
)

type udeskOauthProxy struct {
	*oauthProxy
	appHub *AppHub
}

// für /searchUser
type Answer struct {
	Results []Profile `json:"results"`
}

type Profile struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

var formDecoder = schema.NewDecoder()

var mutex = &sync.Mutex{}

func newUdeskProxy() *udeskOauthProxy {
	hub := newHub()
	proxy := &udeskOauthProxy{appHub: hub}
	return proxy
}

func (r *udeskOauthProxy) createReverseProxy() error {

	// TODO: find better solution
	engine, ok := r.oauthProxy.router.(*chi.Mux)
	if !ok {
		panic("cannot cast to *chi.Mux")
	}

	engine.With(proxyDenyMiddleware).Get("/udesk/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/udesk/admin/", 301)
	})

	fs := http.StripPrefix("/udesk/admin", http.FileServer(http.Dir("frontend/dist")))
	engine.With(r.authenticationMiddleware(), proxyDenyMiddleware).Get("/udesk/admin/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))

	engine.With(r.authenticationMiddleware()).Get("/udesk/getTemplates", r.getAllTemplates)

	engine.With(r.authenticationMiddleware()).Get("/udesk/getTemplate/{query}", r.getTemplate)

	engine.With(r.authenticationMiddleware()).Post("/udesk/updateTemplate/{query}", r.updateTemplate)

	engine.With(r.authenticationMiddleware(),
		r.identityHeadersMiddleware(r.config.AddClaims)).Get("/udesk/startApp", r.startApp)

	engine.With(r.authenticationMiddleware()).Get("/udesk/removeApp/{query}", r.removeApp)

	engine.With(r.authenticationMiddleware()).Get("/udesk/pauseApp/{query}", r.pauseApp)

	engine.With(r.authenticationMiddleware()).Get("/udesk/unpauseApp/{query}", r.unpauseApp)

	engine.With(r.authenticationMiddleware()).Get("/udesk/dockerStatus", r.dockerStatus)

	engine.Get("/udesk/searchUser/{query}", r.searchUser)

	// engine.With(r.authenticationMiddleware()).Get("/udesk/echo", r.appLogging)

	engine.With(r.authenticationMiddleware()).Get("/udesk/appLog/{query}", r.appLogging2)

	engine.Get("/udesk/logout", r.logout)

	return nil
}

func (r *udeskOauthProxy) getAllTemplates(w http.ResponseWriter, req *http.Request) {
	apps := storageProvider.GetAllTemplates()
	js, err := json.Marshal(struct {
		Data *[]storge.TemplateName `json:"data"`
	}{apps})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (r *udeskOauthProxy) getTemplate(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "query")
	app := storageProvider.GetTemplateByName(name)
	js, err := json.Marshal(struct {
		Data *storge.Template `json:"data"`
	}{app})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type Person struct {
	Name         string `schema:"name"`
	InternalPort string `schema:"internalPort"`
}

func (r *udeskOauthProxy) updateTemplate(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	p := new(Person)
	err := formDecoder.Decode(p, req.PostForm)
	fmt.Println(p)

	// name := chi.URLParam(req, "query")
	// app := storageProvider.GetTemplateByName(name)
	js, err := json.Marshal(req.Form)
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

	appTemplate := storageProvider.GetTemplateByName(templateName)
	if appTemplate == nil {
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

	uuid := uuid.New()
	uuidStr := uuid.String()
	name := req.URL.Query().Get("name")
	dockerRunArgs := []string{
		"--rm",
		"-p", strconv.Itoa(port) + ":" + strconv.Itoa(appTemplate.InternalPort),
		"--label", "udesk_uuid=" + uuidStr,
		"--label", "udesk_entry_port=" + strconv.Itoa(port),
		"--label", "udesk_owner=" + user.name,
		"--label", "udesk_name=" + name,
	}

	app := NewApp()
	r.appHub.addApp(uuid, app)

	app.dc.Create(appTemplate.Name, user.name, args, appTemplate, dockerRunArgs, []string{}, func() {})
	x, _ := dockerClient.GetContainer(uuidStr)
	app.dc.Start(x.ID)
	app.ProvisionFinished()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	fmt.Println("finished")
}

func (r *udeskOauthProxy) removeApp(w http.ResponseWriter, req *http.Request) {
	uuidStr := chi.URLParam(req, "query")
	err := r.appHub.removeApp(uuid.MustParse(uuidStr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) pauseApp(w http.ResponseWriter, req *http.Request) {
	uuidStr := chi.URLParam(req, "query")
	err := r.appHub.pauseApp(uuid.MustParse(uuidStr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *udeskOauthProxy) unpauseApp(w http.ResponseWriter, req *http.Request) {
	uuidStr := chi.URLParam(req, "query")
	err := r.appHub.unpauseApp(uuid.MustParse(uuidStr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

		// uuid
		uuid := "UNKNOWN"
		if uuidLabel, ok := container.Labels["udesk_uuid"]; ok {
			uuid = uuidLabel
		}

		y = append(y, []string{
			name,
			owner,
			units.HumanDuration(time.Now().UTC().Sub(t)) + " ago",
			container.State,
			uuid,
		})
	}

abc:
	for uuid := range r.appHub.apps {
		for _, x := range y {
			if x[4] == uuid.String() {
				continue abc
			}
		}
		y = append(y, []string{
			"jhgjhg",
			"owner",
			" ago",
			"provisioning",
			uuid.String(),
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
	fmt.Println("KGHHGJG")
	// u, err := r.getIdentity(req)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	client := gocloak.NewClient("http://localhost:7080/")
	//token := u.token.RawHeader + "." + u.token.RawPayload + "." + strings.TrimRight(base64.URLEncoding.EncodeToString(u.token.Signature), "=")
	token, err := client.LoginAdmin("foo", "foo", "master")
	users, err := client.GetUsers(
		token.AccessToken,
		"master",
		gocloak.GetUsersParams{})
	fmt.Println(users)
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

// func (r *udeskOauthProxy) appLogging(w http.ResponseWriter, req *http.Request) {
// 	c, err := upgrader.Upgrade(w, req, nil)
// 	if err != nil {
// 		log.Print("upgrade:", err)
// 		return
// 	}
// 	defer c.Close()
// 	for {
// 		user, err := r.getIdentity(req)
// 		data := <-appLogger.GetLoggerStream(user.name, user.token.RawHeader)
// 		err = c.WriteMessage(websocket.TextMessage, []byte(strings.ReplaceAll(data.(string), "\n", "\n\r")))
// 		if err != nil {
// 			log.Println("write:", err)
// 			break
// 		}
// 	}
// }

func (r *udeskOauthProxy) appLogging2(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader2.Upgrade(w, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	uuidStr := chi.URLParam(req, "query")

	app, err := r.appHub.getApp(uuid.MustParse(uuidStr))
	client := &AppLogClient{app: app, conn: conn, ringBuffer: NewRingBuffer(100)}
	app.registerLogClient(client)

	go client.streamLog(uuidStr)
	go client.closeStreamLog()
}

func (r *udeskOauthProxy) logout(w http.ResponseWriter, req *http.Request) {
	r.logoutHandler(w, req)
}
