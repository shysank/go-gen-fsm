package go_gen_fsm

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type HandlerMatcher interface {
	Matches(m reflect.Method) bool
	Parts(m reflect.Method) (State, Event)
}

type DefaultMatcher struct {
	delim string
}

func (d *DefaultMatcher) Matches(m reflect.Method) bool {
	match, err := regexp.Match(d.pattern(), []byte(m.Name))
	if !match || err != nil {
		return false
	}

	ops := m.Type.NumOut()
	if ops < 1 || ops > 2 {
		return false
	}

	op1 := m.Type.Out(0).Name()
	if op1 == "State" {
		if ops == 1 {
			return true
		}
		
		if ops == 2 {
			op2 := m.Type.Out(1).Name()
			if op2 == "Duration" {
				return true
			}
		}
	}
	return false
}

func (d *DefaultMatcher) Parts(m reflect.Method) (State, Event) {
	parts := strings.Split(m.Name, d.delim)
	return State(parts[0]), Event(parts[1])
}

func (d *DefaultMatcher) pattern() string {
	return fmt.Sprintf("[A-Z][A-za-z]*%s[A-za-z]+", d.delim)
}
