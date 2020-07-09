package go_gen_fsm

import (
	"fmt"
	"reflect"
	"time"
)

type GenFSM struct {
	fsm           FSM
	state         string
	eventsChannel <-chan Event
}

type FSM interface {
	Init(args ...interface{}) string
}

type Event struct {
	kind string
	data []interface{}
}

func SendEvent(ch chan<- Event, kind string, data ...interface{}) {
	e := Event{kind, data}
	ch <- e
}

func Start(fsm FSM, args ...interface{}) (chan<- Event, <-chan struct{}) {
	eventsChannel := make(chan Event)
	close := make(chan struct{})
	initialState := fsm.Init(args)

	genFsm := GenFSM{fsm: fsm, state: initialState, eventsChannel: eventsChannel}

	fsmType := reflect.TypeOf(fsm)

	go func() {

		for {

			select {

			case e := <-eventsChannel:

				currentState := genFsm.state
				eventKind := e.kind
				methodToInvoke := currentState + "_" + eventKind
				m, present := fsmType.MethodByName(methodToInvoke)
				if !present {
					panic(fmt.Sprintf("%v Method not found", methodToInvoke))
				}

				var values []reflect.Value
				values = append(values, reflect.ValueOf(fsm))
				for _, d := range e.data {
					values = append(values, reflect.ValueOf(d))
				}

				returnValue := m.Func.Call(values)
				genFsm.state = returnValue[0].String()

				_, present = fsmType.MethodByName(returnValue[0].String() + "_" + "timeout")
				if present {
					fmt.Println("Registering timeout")
					time.AfterFunc(4*time.Second, func() {
						SendEvent(eventsChannel, "timeout")
					})
				}

			case <-time.After(15 * time.Second):
				close <- struct{}{}


			}
		}

	}()

	return eventsChannel, close
}
