package syncsaga

import (
	"sync"
	"testing"
	"time"
)

func Test_ReadyGroup(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	rg := NewReadyGroup(
		WithTimeout(0, nil),
		WithValidator(func(rg *ReadyGroup) bool {

			t.Log("validate")

			// Check states
			states := rg.GetParticipantStates()
			for _, ready := range states {
				if !ready {
					return false
				}
			}

			return true
		}),
		WithUpdatedCallback(func(rg *ReadyGroup) {
			t.Log("updated")
		}),
		WithCompletedCallback(func(rg *ReadyGroup) {
			t.Log("completed")
			wg.Done()
		}),
	)

	// Initializing participants
	for i := int64(0); i < 10; i++ {
		rg.Add(i, false)
	}

	rg.Start()

	go func() {
		time.Sleep(3 * time.Second)

		for i := int64(0); i < 10; i++ {
			rg.Ready(i)
		}
	}()

	wg.Wait()
}

func Test_ReadyGroup_Timeout(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	rg := NewReadyGroup(
		WithTimeout(1, func(rg *ReadyGroup) {
			t.Log("timeout")

			// Check states
			states := rg.GetParticipantStates()
			for id, _ := range states {
				rg.Ready(id)
			}
		}),
		WithValidator(func(rg *ReadyGroup) bool {
			t.Log("validate")

			// Check states
			states := rg.GetParticipantStates()
			for _, ready := range states {
				if !ready {
					return false
				}
			}

			return true
		}),
		WithCompletedCallback(func(rg *ReadyGroup) {
			t.Log("completed")
			wg.Done()
		}),
	)

	// Initializing participants
	for i := int64(0); i < 10; i++ {
		rg.Add(i, false)
	}

	rg.Start()

	wg.Wait()
}

func Test_ReadyGroup_Updated(t *testing.T) {

	var wg sync.WaitGroup

	rg := NewReadyGroup(
		WithTimeout(0, nil),
		WithUpdatedCallback(func(rg *ReadyGroup) {
			t.Log("updated")
			wg.Done()
		}),
		WithCompletedCallback(func(rg *ReadyGroup) {
			t.Log("completed")
		}),
	)

	// Initializing participants
	for i := int64(0); i < 10; i++ {
		rg.Add(i, false)
	}

	rg.Start()

	wg.Add(10)
	go func() {
		for i := int64(0); i < 10; i++ {
			rg.Ready(i)
		}
	}()

	wg.Wait()
}
