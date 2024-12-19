package wshandler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/fxamacker/cbor/v2"
)

var (
	awarenessState = make(map[string]map[string]any)
	awarenessChans = make(map[string]chan []string)
	awarenessMutex = sync.RWMutex{}
)

func subscribeAwareness(ctx context.Context, id string, responseChan chan []byte) {
	awarenessMutex.Lock()
	awarenessState[id] = make(map[string]any)
	awarenessChan := make(chan []string)
	awarenessChans[id] = awarenessChan
	awarenessMutex.Unlock()

	go func() {
		var msg []byte
		func() {
			awarenessMutex.RLock()
			defer awarenessMutex.RUnlock()
			awarenessInitMsg := WSResponseMessage{
				Type:    wsResponseMessageAwareness,
				Message: awarenessState,
			}
			var err error
			msg, err = cbor.Marshal(awarenessInitMsg)
			if err != nil {
				slog.Error("cbor failed to marshal awarenessinit", "error", err.Error())
			}
		}()
		responseChan <- msg
	}()

	cleanup := sync.OnceFunc(func() {
		awarenessMutex.Lock()
		defer awarenessMutex.Unlock()
		delete(awarenessState, id)
		delete(awarenessChans, id)
		chans := []chan []string{}
		for _, ch := range awarenessChans {
			chans = append(chans, ch)
		}
		for _, ch := range chans {
			go func() {
				ch <- []string{id}
			}()
		}
	})

	go func() {
		<-ctx.Done()
		cleanup()
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case keys := <-awarenessChan:
				if len(keys) == 0 || (len(keys) == 1 && keys[0] == id) {
					continue
				}

				var msg []byte
				func() {
					awarenessMutex.RLock()
					defer awarenessMutex.RUnlock()
					responseMsg := make(map[string]any)
					for _, key := range keys {
						if val, ok := awarenessState[key]; ok {
							responseMsg[key] = val
						} else {
							responseMsg[key] = nil
						}
					}
					awarenessMsg := WSResponseMessage{
						Type:    wsResponseMessageAwareness,
						Message: responseMsg,
					}
					var err error
					msg, err = cbor.Marshal(awarenessMsg)
					if err != nil {
						slog.Error("cbor failed to marshal awareness", "error", err.Error())
					}
				}()

				responseChan <- msg
			}
		}
	}()
}

func updateAwareness(id string, msg map[string]any) {
	chans := []chan []string{}
	func() {
		awarenessMutex.Lock()
		defer awarenessMutex.Unlock()
		awarenessState[id] = msg
		for _, ch := range awarenessChans {
			chans = append(chans, ch)
		}
	}()

	for _, ch := range chans {
		go func() {
			ch <- []string{id}
		}()
	}
}
