package go_gen_fsm

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type GenFSM struct {
	fsm            FSM
	handlerMatcher HandlerMatcher

	currentState State
	handlers     map[State][]EventHandler

	eventsChannel chan EventMessage
	doneChannel   chan struct{}
	errorChannel  chan error

	timeout time.Duration
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
	Kind Event
	Data []interface{}
}

const (
	TIMEOUT        = "Timeout"
	NOOP           = "NOOP"
	DefaultTimeout = 1 * time.Second
)

func Start(fsm FSM, args ...interface{}) *GenFSM {
	handlers := make(map[State][]EventHandler)
	eventsChannel := make(chan EventMessage)
	doneChannel := make(chan struct{}, 1)
	errorChannel := make(chan error, 10)

	defaultMatcher := &DefaultMatcher{"_"}

	initialState := fsm.Init(args)

	genFsm := GenFSM{fsm: fsm, handlerMatcher: defaultMatcher,
		currentState: initialState, handlers: handlers, eventsChannel:
		eventsChannel, doneChannel: doneChannel, errorChannel: errorChannel,
		timeout: DefaultTimeout}

	genFsm.registerHandlers()
	go genFsm.handleEvents()

	return &genFsm
}

func (g *GenFSM) registerHandlers() {
	fsmType := reflect.TypeOf(g.fsm)
	nMethods := fsmType.NumMethod()

	for i := 0; i < nMethods; i++ {
		m := fsmType.Method(i)
		match := g.handlerMatcher.Matches(m)
		if match {
			state, event := g.handlerMatcher.Parts(m)
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

	return EventHandler{}, errors.New(fmt.Sprintf("No handlers found for state `%s` and event `%s` ", g.currentState, event))
}

func (g *GenFSM) handleEvents() {
	for {
		select {
		case e := <-g.eventsChannel:
			if e.Kind == NOOP {
				g.doneChannel <- struct{}{}
			}
			eventHandler, err := g.getHandler(e.Kind)
			if err != nil {
				g.errorChannel <- err
				continue
			}

			values := []reflect.Value{reflect.ValueOf(g.fsm)}
			for _, d := range e.Data {
				values = append(values, reflect.ValueOf(d))
			}

			returnValues := eventHandler.handlerFunc.Func.Call(values)
			g.currentState = State(returnValues[0].String())

			g.scheduleTimeout()

		}
		g.errorChannel <- nil

	}
}

func (g *GenFSM) scheduleTimeout() {
	_, err := g.getHandler(TIMEOUT)
	if err == nil {
		time.AfterFunc(DefaultTimeout, func() {
			fmt.Println("Sending Timeout Event")
			g.SendEvent(TIMEOUT)
		})
	}
}

func (g *GenFSM) SendEvent(kind Event, data ...interface{}) {
	g.eventsChannel <- EventMessage{kind, data}
}

func (g *GenFSM) Wait() {
	g.SendEvent(NOOP)
	<-g.doneChannel
}

func (g *GenFSM) GetCurrentState() State {
	return g.currentState
}
