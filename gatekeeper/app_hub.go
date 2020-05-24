package gatekeeper

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

// type AppStatus struct {
// 	uuid uuid.UUID
// 	dc   *backend.DockerClient
// }

// AppHub maintains the set of active clients and broadcasts messages to the
// clients.
type AppHub struct {
	// Registered clients.
	// clients map[*Client]bool

	apps map[uuid.UUID]*App

	appsMutex sync.Mutex

	// // Register requests from the clients.
	// register chan *AppLogClient

	// // Unregister requests from clients.
	// unregister chan *AppLogClient

	// registerApp chan *AppLog

	// unregisterApp chan *AppLog
}

func newHub() *AppHub {
	return &AppHub{
		// clients:       make(map[*Client]bool),
		apps: make(map[uuid.UUID]*App),
		// register:      make(chan *AppLogClient),
		// unregister:    make(chan *AppLogClient),
		// registerApp:   make(chan *AppLog),
		// unregisterApp: make(chan *AppLog),
	}
}

// func (h *Hub) run() {
// 	for {
// 		select {
// 		case client := <-h.register:
// 			h.clients[client] = true
// 		case client := <-h.unregister:
// 			if _, ok := h.clients[client]; ok {
// 				delete(h.clients, client)
// 			}
// 		case appStatus := <-h.registerApp:
// 			h.apps[appStatus] = true
// 		case appStatus := <-h.unregisterApp:
// 			if _, ok := h.apps[appStatus]; ok {
// 				delete(h.apps, appStatus)
// 			}
// 		}
// 	}
// }

func (h *AppHub) addApp(uuid uuid.UUID, app *App) error {
	h.appsMutex.Lock()
	defer h.appsMutex.Unlock()

	h.apps[uuid] = app
	return nil
}

func (h *AppHub) removeApp(uuid uuid.UUID) error {
	h.appsMutex.Lock()
	defer h.appsMutex.Unlock()
	if app, ok := h.apps[uuid]; ok {
		app.dc.RemoveContainer(uuid.String())
		delete(h.apps, uuid)
	} else {
		app := NewApp()
		app.dc.RemoveContainer(uuid.String())
	}

	return nil
}

func (h *AppHub) pauseApp(uuid uuid.UUID) error {
	if app, ok := h.apps[uuid]; ok {
		app.dc.PauseContainer(uuid.String())
	} else {
		return errors.New("unknown app")
	}
	return nil
}

func (h *AppHub) unpauseApp(uuid uuid.UUID) error {
	if app, ok := h.apps[uuid]; ok {
		app.dc.UnpauseContainer(uuid.String())
	} else {
		return errors.New("unknown app")
	}
	return nil
}

func (h *AppHub) getApp(uuid uuid.UUID) (*App, error) {
	h.appsMutex.Lock()
	defer h.appsMutex.Unlock()

	if app, ok := h.apps[uuid]; ok {
		return app, nil
	}
	app := NewApp()
	h.apps[uuid] = app
	return app, nil
}
