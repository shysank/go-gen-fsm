package go_gen_fsm

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type HandlerResolver struct {
	nameMatcher NameMatcher
}

type NameMatcher interface {
	Matches(name string) (bool, State, Event)
}

func NewHandlerResolver() HandlerResolver {
	return HandlerResolver{DelimiterMatcher{"_"}}
}

func (h *HandlerResolver) Resolve(m reflect.Method) (bool, State, Event) {
	var returnValid bool
	outputCount := m.Type.NumOut()
	if outputCount < 1 || outputCount > 2 {
		returnValid = false
	}

	op1 := m.Type.Out(0).Name()
	if op1 == "State" {
		if outputCount == 1 {
			returnValid = true
		}

		if outputCount == 2 {
			op2 := m.Type.Out(1).Name()
			if op2 == "Duration" {
				returnValid = true
			}
		}
	}

	if !returnValid {
		return false, State(""), Event("")
	}

	return h.nameMatcher.Matches(m.Name)
}

type DelimiterMatcher struct {
	delim string
}

func (d DelimiterMatcher) Matches(name string) (bool, State, Event) {
	match, err := regexp.Match(d.pattern(), []byte(name))
	if !match || err != nil {
		return false, State(""), Event("")
	}
	parts := strings.Split(name, d.delim)
	return true, State(parts[0]), Event(parts[1])
}

func (d DelimiterMatcher) pattern() string {
	return fmt.Sprintf("^[A-Z][A-Za-z0-9]*%s[A-Za-z][A-Za-z0-9]*$", d.delim)
}
