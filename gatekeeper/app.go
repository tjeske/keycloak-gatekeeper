package gatekeeper

import (
	"container/ring"
	"regexp"
	"sync"

	"github.com/tjeske/keycloak-gatekeeper/backend"
)

// status of an app
type App struct {
	dc           *backend.DockerClient
	provisionLog *ring.Ring

	provisioned bool

	logClientsMutex sync.Mutex

	// connected clients to get log information
	logClients map[*AppLogClient]bool
	// maxSteps    int64
	// currentStep int64
}

func NewApp() (result *App) {
	app := &App{
		provisionLog: ring.New(1000),
		logClients:   make(map[*AppLogClient]bool),
	}
	var abc = backend.NewDockerClientWithWriter("1.2.3", app)
	app.dc = abc
	return app
}

var r = regexp.MustCompile(`Step (?P<currentStep>\d+)/(?P<maxSteps>\d+) : `)

func (al *App) Write(p []byte) (n int, err error) {
	al.logClientsMutex.Lock()
	defer al.logClientsMutex.Unlock()
	newElement := string(p)

	// // match := r.FindStringSubmatch(newElement)
	// // if match != nil {
	// // 	currentStep, _ := strconv.ParseInt(match[1], 10, 64)
	// // 	maxSteps, _ := strconv.ParseInt(match[2], 10, 64)
	// // 	al.currentStep = currentStep
	// // 	al.maxSteps = maxSteps
	// // }

	al.provisionLog.Value = newElement
	al.provisionLog = al.provisionLog.Next()
	for client := range al.logClients {
		client.ringBuffer.Input <- string(p)
		// if al.maxSteps != 0 {
		// 	sessionInfo.stepsRingBuffer.Input <- string((al.currentStep / al.maxSteps) * 100)
		// } else {
		// 	sessionInfo.stepsRingBuffer.Input <- string(1)
		// }
	}
	return len(p), nil
}

// func (al *AppLogger) GetProvisioningStatus(user, sessionId string) <-chan interface{} {
// 	al.mux.Lock()
// 	defer al.mux.Unlock()
// 	sessionInfos := al.userMap[user]
// 	if sessionInfos != nil {
// 		for _, sessionInfo := range sessionInfos {
// 			if sessionInfo.context == sessionId {
// 				// found
// 				return sessionInfo.ringBuffer.Output
// 			}
// 		}
// 	}
// 	// create new sessionInfo
// 	newSessionInfo := sessionInfo{context: sessionId, ringBuffer: NewRingBuffer(100), stepsRingBuffer: NewRingBuffer((1))}
// 	if _, ok := al.userMap[user]; ok {
// 		sessionInfos := al.userMap[user]
// 		sessionInfos = append(sessionInfos, newSessionInfo)
// 	} else {
// 		sessionInfos := make([]sessionInfo, 0)
// 		sessionInfos = append(sessionInfos, newSessionInfo)
// 		al.userMap[user] = sessionInfos
// 	}

// 	return newSessionInfo.stepsRingBuffer.Output
// }

func (a *App) registerLogClient(client *AppLogClient) {
	a.logClientsMutex.Lock()
	defer a.logClientsMutex.Unlock()

	a.logClients[client] = true
}

func (a *App) unregisterLogClient(client *AppLogClient) {
	a.logClientsMutex.Lock()
	defer a.logClientsMutex.Unlock()

	delete(a.logClients, client)
}

func (a *App) ProvisionFinished() {
	a.logClientsMutex.Lock()
	defer a.logClientsMutex.Unlock()

	for client := range a.logClients {
		close(client.ringBuffer.Input)
	}
	a.provisioned = true
}
