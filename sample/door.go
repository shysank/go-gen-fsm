package sample

import (
	"bytes"
	"fmt"
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
	arg := args[0].([]interface{})
	d.code = arg[0].(string)
	return "Locked"
}

func (d *Door) Locked_Button(digit rune) (State, time.Duration) {
	time.Sleep(100 * time.Millisecond)
	d.sofar.WriteRune(digit)
	sofarStr := d.sofar.String()
	if sofarStr == d.code {
		fmt.Printf("Opened. Value Entered: %s\n", sofarStr)
		return "Open", LockTimeout
	}

	if len(sofarStr) < len(d.code) {
		fmt.Printf("Partial Entry. Value Entered: %s\n", sofarStr)
		return "Locked", -1
	}

	fmt.Println("Wrong Code")

	d.sofar.Reset()
	return "Locked", -1

}

func (d *Door) Open_Timeout() State {
	fmt.Printf("Timeout Epired. Going to Lock the door again\n")
	d.sofar.Reset()
	return "Locked"
}

func (d *Door) Open_Reset(pass string) State {
	fmt.Printf("Resetting lock with new pass %s\n", pass)
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
