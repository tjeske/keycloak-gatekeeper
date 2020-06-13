package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nerzal/gocloak/v5"
	"github.com/go-chi/chi"
)

type Answer struct {
	Results []Profile `json:"results"`
}

type Profile struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {

	client := gocloak.NewClient("http://auth.familie-jeske.net:7080/")
	token, err := client.LoginAdmin("admin", "admin", "master")
	if err != nil {
		panic("Something wrong with the credentials or url")
	}

	r := chi.NewRouter()

	r.Route("/", func(e chi.Router) {
		e.Get("/udesk/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hi"))
		})
		e.Get("/udesk/searchUser/{query}", func(w http.ResponseWriter, r *http.Request) {

			users, err := client.GetUsers(
				token.AccessToken,
				"master",
				gocloak.GetUsersParams{})

			filteredUserOrGroups := []Profile{}
			for _, user := range users {
				if strings.HasPrefix(strings.ToLower(*user.Username), strings.ToLower(chi.URLParam(r, "query"))) {
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
		})
	})

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "dist")
	FileServer(r, "/admin", http.Dir(filesDir))

	http.ListenAndServe(":3333", r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
