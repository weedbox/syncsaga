package syncsaga

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/weedbox/timebank"
)

type ReadyGroupOpt func(*ReadyGroup)
type ReadyGroupCallback func(*ReadyGroup)
type ReadyGroupValidator func(*ReadyGroup) bool

type ReadyGroupAction struct {
	ParticipantID int64
	IsReady       bool
}

type ReadyGroup struct {
	mu              sync.RWMutex
	participants    map[int64]bool
	timeoutInterval int
	timebank        *timebank.TimeBank
	isCompleted     int32
	actionCh        chan *ReadyGroupAction
	validator       ReadyGroupValidator
	onUpdated       ReadyGroupCallback
	onCompleted     ReadyGroupCallback
	onTimeout       ReadyGroupCallback
}

func WithTimeout(timeout int, callback ReadyGroupCallback) ReadyGroupOpt {
	return func(rg *ReadyGroup) {
		rg.timeoutInterval = timeout

		if callback != nil {
			rg.onTimeout = callback
		}
	}
}

func WithValidator(v ReadyGroupValidator) ReadyGroupOpt {
	return func(rg *ReadyGroup) {
		rg.validator = v
	}
}

func WithUpdatedCallback(callback ReadyGroupCallback) ReadyGroupOpt {
	return func(rg *ReadyGroup) {
		rg.onUpdated = callback
	}
}

func WithCompletedCallback(callback ReadyGroupCallback) ReadyGroupOpt {
	return func(rg *ReadyGroup) {
		rg.onCompleted = callback
	}
}

func NewReadyGroup(opts ...ReadyGroupOpt) *ReadyGroup {

	rg := &ReadyGroup{
		participants:    make(map[int64]bool),
		timeoutInterval: 0,
		timebank:        timebank.NewTimeBank(),
		validator:       func(rg *ReadyGroup) bool { return rg.defValidate() },
		onUpdated:       func(*ReadyGroup) {},
		onCompleted:     func(*ReadyGroup) {},
		onTimeout:       func(*ReadyGroup) {},
	}

	for _, o := range opts {
		o(rg)
	}

	return rg
}

func (rg *ReadyGroup) defValidate() bool {

	rg.mu.RLock()
	defer rg.mu.RUnlock()

	for _, ready := range rg.participants {
		if !ready {
			return false
		}
	}

	return true
}

func (rg *ReadyGroup) validate() {

	rg.mu.RLock()
	defer rg.mu.RUnlock()

	if rg.validator(rg) {
		rg.Done()
	}
}

func (rg *ReadyGroup) updateState(participantID int64, isReady bool) {

	rg.mu.Lock()
	defer rg.mu.Unlock()

	if _, ok := rg.participants[participantID]; !ok {
		return
	}

	rg.participants[participantID] = isReady
}

func (rg *ReadyGroup) Add(participantID int64, isReady bool) {

	rg.mu.Lock()
	defer rg.mu.Unlock()

	rg.participants[participantID] = isReady
}

func (rg *ReadyGroup) Start() {

	if rg.actionCh != nil {
		close(rg.actionCh)
	}

	rg.actionCh = make(chan *ReadyGroupAction, 256)

	go func() {
		for action := range rg.actionCh {
			rg.updateState(action.ParticipantID, action.IsReady)
			rg.validate()
			go rg.onUpdated(rg)
		}
	}()

	// No time limit
	if rg.timeoutInterval == 0 {
		return
	}

	// Initializing timeout handler
	duration := time.Duration(rg.timeoutInterval) * time.Second
	rg.timebank.NewTask(duration, func(isCancelled bool) {
		if isCancelled {
			return
		}

		rg.onTimeout(rg)
	})
}

func (rg *ReadyGroup) Stop() {
	close(rg.actionCh)
	rg.timebank.Cancel()
}

func (rg *ReadyGroup) Ready(participantID int64) {
	rg.actionCh <- &ReadyGroupAction{
		ParticipantID: participantID,
		IsReady:       true,
	}
}

func (rg *ReadyGroup) Discard(participantID int64) {
	rg.actionCh <- &ReadyGroupAction{
		ParticipantID: participantID,
		IsReady:       false,
	}
}

func (rg *ReadyGroup) Done() {

	if atomic.LoadInt32(&rg.isCompleted) == 1 {
		return
	}

	atomic.StoreInt32(&rg.isCompleted, 1)

	if rg.timeoutInterval > 0 {
		rg.timebank.Cancel()
	}

	rg.onCompleted(rg)
}

func (rg *ReadyGroup) GetParticipantStates() map[int64]bool {

	rg.mu.RLock()
	defer rg.mu.RUnlock()

	states := make(map[int64]bool)

	for k, v := range rg.participants {
		states[k] = v
	}

	return states
}
