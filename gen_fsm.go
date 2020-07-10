package go_gen_fsm

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type GenFSM struct {
	fsm           FSM
	currentState  State
	eventsChannel chan EventMessage

	handlers map[State][]EventHandler
}

type State string
type Event string
type EventHandler struct {
	event       Event
	handlerFunc reflect.Method
}

type FSM interface {
	Init(args ...interface{}) State
}

type EventMessage struct {
	kind Event
	data []interface{}
}

const (
	TIMEOUT_EVENT = "timeout"
)

func (g *GenFSM) SendEvent(kind Event, data ...interface{}) {
	g.eventsChannel <- EventMessage{kind, data}
}

func Start(fsm FSM, args ...interface{}) GenFSM {
	eventsChannel := make(chan EventMessage)
	initialState := fsm.Init(args)

	genFsm := GenFSM{fsm: fsm, currentState: initialState, eventsChannel: eventsChannel, handlers: make(map[State][]EventHandler)}
	genFsm.registerHandlers()
	go genFsm.handleEvents()

	return genFsm
}

func (g *GenFSM) registerHandlers() {
	fsmType := reflect.TypeOf(g.fsm)
	nMethods := fsmType.NumMethod()

	for i := 0; i < nMethods; i++ {
		m := fsmType.Method(i)
		match, _ := regexp.Match("[A-Z][A-za-z]*_[A-za-z]+", []byte(m.Name))
		if match {
			mParts := strings.Split(m.Name, "_")
			state := State(mParts[0])
			event := Event(mParts[1])
			eventHandler := EventHandler{event: event, handlerFunc: m}
			fmt.Printf("Adding handler %s\n", m.Name)

			if h, ok := g.handlers[state]; ok {
				h = append(h, eventHandler)
			} else {
				g.handlers[state] = []EventHandler{eventHandler}
			}
		}
	}
}

func (g *GenFSM) getHandler(event Event) (EventHandler, error) {
	handlers, ok := g.handlers[g.currentState]
	if !ok {
		return EventHandler{}, errors.New(fmt.Sprintf("No handlers found for state %s", g.currentState))
	}

	for _, h := range handlers {
		if h.event == event {
			return h, nil
		}
	}

	return EventHandler{}, errors.New(fmt.Sprintf("No handlers found for state %s and event %s", g.currentState, event))
}

func (g *GenFSM) handleEvents() {
	for {
		select {
		case e := <-g.eventsChannel:
			eventHandler, err := g.getHandler(e.kind)
			if err != nil {
				fmt.Println(err.Error())
			}

			values := []reflect.Value{reflect.ValueOf(g.fsm)}
			for _, d := range e.data {
				values = append(values, reflect.ValueOf(d))
			}

			returnValues := eventHandler.handlerFunc.Func.Call(values)
			g.currentState = State(returnValues[0].String())

			g.scheduleTimeout()

		}
	}
}

func (g *GenFSM) scheduleTimeout() {
	_, err := g.getHandler(TIMEOUT_EVENT)
	if err == nil {
		time.AfterFunc(4*time.Second, func() {
			g.SendEvent(TIMEOUT_EVENT)
		})
	}
}
