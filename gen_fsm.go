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

	timer *time.Timer
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
	NOOP    = "NOOP"
	TIMEOUT = "Timeout"
	STOP    = "Stop"
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
		eventsChannel, doneChannel: doneChannel, errorChannel: errorChannel}

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

			var stateEventHandlers []EventHandler
			if _, ok := g.handlers[state]; ok {
				stateEventHandlers = append(stateEventHandlers, eventHandler)
			} else {
				stateEventHandlers = append(stateEventHandlers, eventHandler)
			}
			g.handlers[state] = stateEventHandlers
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
				continue
			}
			if e.Kind == STOP {
				close(g.eventsChannel)
				close(g.errorChannel)
				close(g.doneChannel)
				return
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
			if len(returnValues) == 2 {
				timeout := returnValues[1].Interface().(time.Duration)
				g.scheduleTimeout(timeout)
			}

		}
		g.errorChannel <- nil

	}
}

func (g *GenFSM) scheduleTimeout(timeout time.Duration) {
	if timeout == -1 {
		return
	}

	fmt.Printf("Schedule a timeout event after %s\n", timeout.String())
	g.timer = time.NewTimer(timeout)
	go func() {
		<-g.timer.C
		fmt.Printf("Sending timeout event\n")
		g.SendEvent(TIMEOUT)
	}()
}

func (g *GenFSM) SendEvent(kind Event, data ...interface{}) {
	g.eventsChannel <- EventMessage{kind, data}
}

func (g *GenFSM) Wait() {
	g.SendEvent(NOOP)
	<-g.doneChannel
}

func (g *GenFSM) Stop() {
	g.SendEvent(STOP)
	if g.timer != nil {
		g.timer.Stop()
	}
}

func (g *GenFSM) GetCurrentState() State {
	return g.currentState
}
