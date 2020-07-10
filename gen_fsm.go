package go_gen_fsm

import (
	"fmt"
	"reflect"
	"time"
)

type GenFSM struct {
	fsm           FSM
	currentState  string
	eventsChannel chan Event
}

type FSM interface {
	Init(args ...interface{}) string
}

type Event struct {
	kind string
	data []interface{}
}

func (g *GenFSM) SendEvent(kind string, data ...interface{}) {
	e := Event{kind, data}
	g.eventsChannel <- e
}

func Start(fsm FSM, args ...interface{}) GenFSM {
	eventsChannel := make(chan Event)
	initialState := fsm.Init(args)

	genFsm := GenFSM{fsm: fsm, currentState: initialState, eventsChannel: eventsChannel}

	go genFsm.handleEvents()

	return genFsm
}

func (g *GenFSM) handleEvents() {
	fsmType := reflect.TypeOf(g.fsm)
	for {

		select {

		case e := <-g.eventsChannel:

			eventHandler := inferEventHandler(g.currentState, e.kind)
			m, present := fsmType.MethodByName(eventHandler)
			if !present {
				panic(fmt.Sprintf("Handler not found for event %s in state %s. Is there a method `%s` definded?", e.kind, g.currentState, eventHandler))
			}

			values := []reflect.Value{reflect.ValueOf(g.fsm)}
			for _, d := range e.data {
				values = append(values, reflect.ValueOf(d))
			}

			returnValue := m.Func.Call(values)
			g.currentState = returnValue[0].String()

			_, present = fsmType.MethodByName(returnValue[0].String() + "_" + "timeout")
			if present {
				fmt.Println("Registering timeout")
				time.AfterFunc(4*time.Second, func() {
					g.SendEvent("timeout")
				})
			}
		}
	}
}

func inferEventHandler(state, event string) string {
	return state + "_" + event
}
