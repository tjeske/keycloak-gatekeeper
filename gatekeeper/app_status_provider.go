package gatekeeper

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/tjeske/keycloak-gatekeeper/backend"
)

// "mock connectors" for unit-tesing
var logFatalf = log.Fatalf
var filesystem = afero.NewOsFs()

type Status int

const (
	Running Status = iota
	Stopped
	Paused
	Provisioning
)

type appInfo struct {
	name          string
	owner         string
	created       time.Time
	status        Status
	provCompleted int
}

// extract information from container engine
type AppStatusProviderType struct {
	dockerClient *backend.DockerClient
	mux          *sync.Mutex
	apps         map[uuid.UUID]appInfo
	appsProv     map[uuid.UUID]appInfo
}

var AppStatusProvider AppStatusProviderType

func SetProvider(newProvider AppStatusProviderType) {
	AppStatusProvider = newProvider
}

func NewAppStatusProvider() AppStatusProviderType {
	var dockerClient = backend.NewDockerClient("1.2.3")
	p := AppStatusProviderType{
		dockerClient: dockerClient,
		apps:         make(map[uuid.UUID]appInfo),
		appsProv:     make(map[uuid.UUID]appInfo),
	}
	return p
}

func (r *AppStatusProviderType) GetStatusAll() map[uuid.UUID]appInfo {
	r.mux.Lock()
	defer r.mux.Unlock()

	res := make(map[uuid.UUID]appInfo)
	for k, v := range r.apps {
		res[k] = v
	}
	for k, v := range r.appsProv {
		res[k] = v
	}
	return res
}

func (r *AppStatusProviderType) Refresh() {
	containers := r.dockerClient.GetStatus()
	r.mux.Lock()
	defer r.mux.Unlock()

	r.apps = make(map[uuid.UUID]appInfo)

	for _, container := range containers {

		createdTime := time.Unix(container.Created, 0)

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
		if uuidLabel, ok := container.Labels["udesk_uuid"]; ok {
			uuid, err := uuid.Parse(uuidLabel)
			if err == nil {
				r.apps[uuid] = appInfo{
					name:    name,
					owner:   owner,
					created: createdTime,
					//status: container.State,
				}
			}
		}
	}
}
