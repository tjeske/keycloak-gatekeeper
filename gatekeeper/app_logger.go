package gatekeeper

import (
	"container/ring"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type sessionInfo struct {
	context         string
	ringBuffer      *Ringbuf
	stepsRingBuffer *Ringbuf
	lastAccess      time.Time
}

type AppLogger struct {
	data        *ring.Ring
	mux         sync.Mutex
	userMap     map[string][]sessionInfo
	maxSteps    int64
	currentStep int64
}

func NewAppLogger(size int) (result *AppLogger) {
	return &AppLogger{
		data:    ring.New(1000),
		userMap: make(map[string][]sessionInfo),
	}
}

var r = regexp.MustCompile(`Step (?P<currentStep>\d+)/(?P<maxSteps>\d+) : `)

func (al *AppLogger) Write(p []byte) (n int, err error) {
	al.mux.Lock()
	defer al.mux.Unlock()
	newElement := string(p)

	match := r.FindStringSubmatch(newElement)
	if match != nil {
		currentStep, _ := strconv.ParseInt(match[1], 10, 64)
		maxSteps, _ := strconv.ParseInt(match[2], 10, 64)
		al.currentStep = currentStep
		al.maxSteps = maxSteps
	}

	al.data.Value = newElement
	al.data = al.data.Next()
	for _, v := range al.userMap {
		for _, sessionInfo := range v {
			sessionInfo.ringBuffer.Input <- string(p)
			if al.maxSteps != 0 {
				sessionInfo.stepsRingBuffer.Input <- string((al.currentStep / al.maxSteps) * 100)
			} else {
				sessionInfo.stepsRingBuffer.Input <- string(1)
			}
		}
	}
	return len(p), nil
}

func (al *AppLogger) GetLoggerStream(user, sessionId string) <-chan interface{} {
	al.mux.Lock()
	defer al.mux.Unlock()
	sessionInfos := al.userMap[user]
	if sessionInfos != nil {
		for _, sessionInfo := range sessionInfos {
			if sessionInfo.context == sessionId {
				// found
				return sessionInfo.ringBuffer.Output
			}
		}
	}
	// create new sessionInfo
	newSessionInfo := sessionInfo{context: sessionId, ringBuffer: NewRingBuffer(100), stepsRingBuffer: NewRingBuffer((1))}
	if _, ok := al.userMap[user]; ok {
		sessionInfos := al.userMap[user]
		sessionInfos = append(sessionInfos, newSessionInfo)
	} else {
		sessionInfos := make([]sessionInfo, 0)
		sessionInfos = append(sessionInfos, newSessionInfo)
		al.userMap[user] = sessionInfos
	}

	return newSessionInfo.ringBuffer.Output
}

func (al *AppLogger) GetProvisioningStatus(user, sessionId string) <-chan interface{} {
	al.mux.Lock()
	defer al.mux.Unlock()
	sessionInfos := al.userMap[user]
	if sessionInfos != nil {
		for _, sessionInfo := range sessionInfos {
			if sessionInfo.context == sessionId {
				// found
				return sessionInfo.ringBuffer.Output
			}
		}
	}
	// create new sessionInfo
	newSessionInfo := sessionInfo{context: sessionId, ringBuffer: NewRingBuffer(100), stepsRingBuffer: NewRingBuffer((1))}
	if _, ok := al.userMap[user]; ok {
		sessionInfos := al.userMap[user]
		sessionInfos = append(sessionInfos, newSessionInfo)
	} else {
		sessionInfos := make([]sessionInfo, 0)
		sessionInfos = append(sessionInfos, newSessionInfo)
		al.userMap[user] = sessionInfos
	}

	return newSessionInfo.stepsRingBuffer.Output
}
