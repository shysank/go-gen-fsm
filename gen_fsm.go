package go_gen_fsm

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type GenFSM struct {
	fsm             FSM
	handlerResolver HandlerResolver

	currentState State
	handlers     map[State][]EventHandler

	eventsChannel chan EventMessage
	errorChannel  chan error

	sync sync

	timer *time.Timer
}

type FSM interface {
	Init(args ...interface{}) State
}

type State string
type Event string
type EventHandler struct {
	event       Event
	handlerFunc reflect.Method
}
type EventMessage struct {
	Kind Event
	Data []interface{}
}
type sync struct {
	req  chan interface{}
	resp chan interface{}
}

const (
	NOOP         = "NOOP"
	TIMEOUT      = "Timeout"
	STOP         = "Stop"
	genericState = State("Handle")
	genericEvent = Event("Info")
)

func Start(fsm FSM, args ...interface{}) *GenFSM {
	g := newGenFsm()
	g.fsm = fsm
	g.currentState = fsm.Init(args...)

	g.registerHandlers()

	go g.doStart()

	return g
}

func newGenFsm() *GenFSM {
	g := new(GenFSM)
	g.handlers = make(map[State][]EventHandler)
	g.eventsChannel = make(chan EventMessage)
	g.errorChannel = make(chan error, 10)
	g.sync = sync{make(chan interface{}, 1), make(chan interface{}, 1)}
	g.handlerResolver = NewHandlerResolver()
	return g
}

func (g *GenFSM) registerHandlers() {
	fsmType := reflect.TypeOf(g.fsm)
	nMethods := fsmType.NumMethod()

	for i := 0; i < nMethods; i++ {
		m := fsmType.Method(i)
		match, state, event := g.handlerResolver.Resolve(m)
		if match {
			eventHandler := EventHandler{event: event, handlerFunc: m}

			var stateEventHandlers []EventHandler
			if h, ok := g.handlers[state]; ok {
				stateEventHandlers = append(h, eventHandler)
			} else {
				stateEventHandlers = append(stateEventHandlers, eventHandler)
			}
			g.handlers[state] = stateEventHandlers
		}
	}
}

func (g *GenFSM) doStart() {
	var shutdown bool
	for {
		if shutdown {
			break
		}
		select {
		case e, ok := <-g.eventsChannel:
			if !ok {
				shutdown = true
				break
			}
			g.handleEvent(e)

		case r, _ := <-g.sync.req:
			g.handleSync(r)
		}

	}
}

func (g *GenFSM) handleEvent(e EventMessage) {
	eventHandler, err := g.getHandler(e.Kind)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		g.errorChannel <- err
		fmt.Printf("Checking if there is a generic `handle_info` handler\n")
		eventHandler, err = g.getGenericHandler()
		if err != nil {
			fmt.Printf("Cannot find a generic handler `handle_info`\n")
			g.errorChannel <- err
			return
		}

	}

	g.cancelTimer()

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
	g.errorChannel <- nil
}

func getHandler(handlers map[State][]EventHandler, state State, event Event) (EventHandler, error) {
	stateHandlers, ok := handlers[state]
	if !ok {
		return EventHandler{}, errors.New(fmt.Sprintf("No handlers found for state %s", state))
	}

	for _, h := range stateHandlers {
		if h.event == event {
			return h, nil
		}
	}

	return EventHandler{}, errors.New(fmt.Sprintf("No handlers found for state `%s` and event `%s` ", state, event))
}

func (g *GenFSM) getHandler(event Event) (EventHandler, error) {
	return getHandler(g.handlers, g.currentState, event)
}

func (g *GenFSM) getGenericHandler() (EventHandler, error) {
	return getHandler(g.handlers, genericState, genericEvent)
}

func (g *GenFSM) handleSync(req interface{}) {
	var resp interface{}
	if reqStr, ok := req.(string); ok {
		switch reqStr {
		case NOOP:
			resp = "Noop"
			g.sync.resp <- resp
		case STOP:
			resp = "Shutdown"
			g.sync.resp <- resp
			g.closeAllChannels()
		}
	}
}

func (g *GenFSM) closeAllChannels() {
	close(g.eventsChannel)
	close(g.errorChannel)
	close(g.sync.req)
	close(g.sync.resp)
}

func (g *GenFSM) scheduleTimeout(timeout time.Duration) {
	if timeout == -1 {
		return
	}

	g.timer = time.NewTimer(timeout)
	go func() {
		<-g.timer.C
		g.SendEvent(TIMEOUT)
	}()
}

func (g *GenFSM) cancelTimer() {
	if g.timer != nil {
		g.timer.Stop()
	}
}

func (g *GenFSM) SendEvent(kind Event, data ...interface{}) {
	g.eventsChannel <- EventMessage{kind, data}
}

func (g *GenFSM) SendSyncReq(req interface{}) (resp interface{}) {
	g.sync.req <- req
	resp = <-g.sync.resp
	return resp
}

func (g *GenFSM) Wait() {
	g.SendSyncReq(NOOP)
}

func (g *GenFSM) Stop() {
	g.cancelTimer()
	g.SendSyncReq(STOP)
}

func (g *GenFSM) GetCurrentState() State {
	return g.currentState
}
