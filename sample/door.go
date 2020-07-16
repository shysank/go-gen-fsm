package sample

import (
	"bytes"
	. "github.com/cynic89/go-gen-fsm"
	"time"
)

type Door struct {
	code  string
	sofar bytes.Buffer
}

const (
	LockTimeout = 1 * time.Second
)

func (d *Door) Init(args ...interface{}) State {
	d.code = args[0].(string)
	return "Locked"
}

func (d *Door) Locked_Button(digit rune) (State, time.Duration) {
	time.Sleep(100 * time.Millisecond)
	d.sofar.WriteRune(digit)
	sofarStr := d.sofar.String()
	if sofarStr == d.code {
		return "Open", LockTimeout
	}

	if len(sofarStr) < len(d.code) {
		return "Locked", -1
	}

	d.sofar.Reset()
	return "Locked", -1

}

func (d *Door) Open_Timeout() State {
	d.sofar.Reset()
	return "Locked"
}

func (d *Door) Open_Reset(pass string) State {
	d.sofar.Reset()
	d.code = pass
	return "Locked"
}

func Button(g *GenFSM, digit rune) {
	g.SendEvent("Button", digit)
}

func ResetLock(g *GenFSM, pass string) {
	g.SendEvent("Reset", pass)
}
